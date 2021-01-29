// Copyright 2020 lack
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package registry

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	json "github.com/json-iterator/go"

	regpb "github.com/lack-io/vine/proto/apis/registry"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/util/mdns"
)

var (
	// use a .vine domain rather than .local
	mdnsDomain = "vine"
)

type mdnsTxt struct {
	Service   string
	Version   string
	node      []*regpb.Node
	Endpoints []*regpb.Endpoint
	Metadata  map[string]string
	Apis      []*regpb.OpenAPI
}

type mdnsEntry struct {
	id   string
	node *mdns.Server
}

type mdnsRegistry struct {
	opts Options
	// the mdns domain
	domain string

	sync.Mutex
	services map[string][]*mdnsEntry

	mtx sync.RWMutex

	// watchers
	watchers map[string]*mdnsWatcher

	// listener
	listener chan *mdns.ServiceEntry
}

type mdnsWatcher struct {
	id   string
	wo   WatchOptions
	ch   chan *mdns.ServiceEntry
	exit chan struct{}
	// the mdns domain
	domain string
	// the registry
	registry *mdnsRegistry
}

func encode(txt *mdnsTxt) ([]string, error) {
	b, err := json.Marshal(txt)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	defer buf.Reset()

	w := zlib.NewWriter(&buf)
	if _, err := w.Write(b); err != nil {
		return nil, err
	}
	w.Close()

	encoded := hex.EncodeToString(buf.Bytes())

	// individual txt limit
	if len(encoded) <= 255 {
		return []string{encoded}, nil
	}

	// split encoded string
	var record []string

	for len(encoded) > 255 {
		record = append(record, encoded[:255])
		encoded = encoded[255:]
	}

	record = append(record, encoded)

	return record, nil
}

func decode(record []string) (*mdnsTxt, error) {
	encoded := strings.Join(record, "")

	hr, err := hex.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	br := bytes.NewReader(hr)
	zr, err := zlib.NewReader(br)
	if err != nil {
		return nil, err
	}

	rbuf, err := ioutil.ReadAll(zr)
	if err != nil {
		return nil, err
	}

	var txt *mdnsTxt

	if err := json.Unmarshal(rbuf, &txt); err != nil {
		return nil, err
	}

	return txt, nil
}

func newRegistry(opts ...Option) Registry {
	options := Options{
		Context: context.Background(),
		Timeout: time.Millisecond * 100,
	}

	for _, o := range opts {
		o(&options)
	}

	// set the domain
	domain := mdnsDomain

	d, ok := options.Context.Value("mdns.domain").(string)
	if ok {
		domain = d
	}

	return &mdnsRegistry{
		opts:     options,
		domain:   domain,
		services: make(map[string][]*mdnsEntry),
		watchers: make(map[string]*mdnsWatcher),
	}
}

func (m *mdnsRegistry) Init(opts ...Option) error {
	for _, o := range opts {
		o(&m.opts)
	}
	return nil
}

func (m *mdnsRegistry) Options() Options {
	return m.opts
}

func (m *mdnsRegistry) Register(service *regpb.Service, opts ...RegisterOption) error {
	m.Lock()
	defer m.Unlock()

	entries, ok := m.services[service.Name]
	// first entry, create wildcard used for list queries
	if !ok {
		s, err := mdns.NewMDNSService(
			service.Name,
			"_services",
			m.domain+".",
			"",
			9999,
			[]net.IP{net.ParseIP("0.0.0.0")},
			nil,
		)
		if err != nil {
			return err
		}

		svc, err := mdns.NewServer(&mdns.Config{Zone: &mdns.DNSSDService{MDNSService: s}})
		if err != nil {
			return err
		}

		// append the wildcard entry
		entries = append(entries, &mdnsEntry{id: "*", node: svc})
	}

	var gerr error

	for _, node := range service.Nodes {
		var seen bool
		var e *mdnsEntry

		for _, entry := range entries {
			if node.Id == entry.id {
				seen = true
				e = entry
				break
			}
		}

		// already registered, continue
		if seen {
			continue
			// doesn't exist
		} else {
			e = &mdnsEntry{}
		}

		txt, err := encode(&mdnsTxt{
			Service:   service.Name,
			Version:   service.Version,
			Endpoints: service.Endpoints,
			node:      service.Nodes,
			Metadata:  node.Metadata,
			Apis:      service.Apis,
		})

		if err != nil {
			gerr = err
			continue
		}

		host, pt, err := net.SplitHostPort(node.Address)
		if err != nil {
			gerr = err
			continue
		}
		port, _ := strconv.Atoi(pt)

		log.Infof("[mdns] registry create new service with ip: %s for: %s", net.ParseIP(host).String(), host)

		// we got here, new node
		s, err := mdns.NewMDNSService(
			node.Id,
			service.Name,
			m.domain+".",
			"",
			port,
			[]net.IP{net.ParseIP(host)},
			txt,
		)
		if err != nil {
			gerr = err
			continue
		}

		svc, err := mdns.NewServer(&mdns.Config{Zone: s, LocalhostChecking: true})
		if err != nil {
			gerr = err
			continue
		}

		e.id = node.Id
		e.node = svc
		entries = append(entries, e)
	}

	// save
	m.services[service.Name] = entries

	return gerr
}

func (m *mdnsRegistry) Deregister(service *regpb.Service, opts ...DeregisterOption) error {
	m.Lock()
	defer m.Unlock()

	var newEntries []*mdnsEntry

	// loop existing entries, check if any match, shutdown those that do
	for _, entry := range m.services[service.Name] {
		var remove bool

		for _, node := range service.Nodes {
			if node.Id == entry.id {
				entry.node.Shutdown()
				remove = true
				break
			}
		}

		// keep it?
		if !remove {
			newEntries = append(newEntries, entry)
		}
	}

	// last entry is the wildcard for list queries. Remove it.
	if len(newEntries) == 1 && newEntries[0].id == "*" {
		newEntries[0].node.Shutdown()
		delete(m.services, service.Name)
	} else {
		m.services[service.Name] = newEntries
	}

	return nil
}

func (m *mdnsRegistry) GetService(service string, opts ...GetOption) ([]*regpb.Service, error) {
	serviceMap := make(map[string]*regpb.Service)
	entries := make(chan *mdns.ServiceEntry, 10)
	done := make(chan bool)

	p := mdns.DefaultParams(service)
	// set context with timeout
	var cancel context.CancelFunc
	p.Context, cancel = context.WithTimeout(context.Background(), m.opts.Timeout)
	defer cancel()
	// set entries channel
	p.Entries = entries
	// set the domain
	p.Domain = m.domain

	go func() {
		for {
			select {
			case e := <-entries:
				// list record so skip
				if p.Service == "_services" {
					continue
				}
				if p.Domain != m.domain {
					continue
				}
				if e.TTL == 0 {
					continue
				}

				txt, err := decode(e.InfoFields)
				if err != nil {
					continue
				}

				if txt.Service != service {
					continue
				}

				s, ok := serviceMap[txt.Version]
				if !ok {
					s = &regpb.Service{
						Name:      txt.Service,
						Version:   txt.Version,
						Metadata:  txt.Metadata,
						Nodes:     txt.node,
						Endpoints: txt.Endpoints,
						Apis:      txt.Apis,
					}
				}
				addr := ""
				// prefer ipv4 addrs
				if len(e.AddrV4) > 0 {
					addr = e.AddrV4.String()
					// else use ipv6
				} else if len(e.AddrV6) > 0 {
					addr = "[" + e.AddrV6.String() + "]"
				} else {
					log.Infof("[mdns]: invalid endpoint received: %v", e)
					continue
				}
				s.Nodes = append(s.Nodes, &regpb.Node{
					Id:       strings.TrimSuffix(e.Name, "."+p.Service+"."+p.Domain+"."),
					Address:  fmt.Sprintf("%s:%d", addr, e.Port),
					Metadata: txt.Metadata,
				})

				serviceMap[txt.Version] = s
			case <-p.Context.Done():
				close(done)
				return
			}
		}
	}()

	// execute the query
	if err := mdns.Query(p); err != nil {
		return nil, err
	}

	// wait for completion
	<-done

	// create list and return
	services := make([]*regpb.Service, 0, len(serviceMap))

	for _, service := range serviceMap {
		services = append(services, service)
	}

	return services, nil
}

func (m *mdnsRegistry) ListServices(opts ...ListOption) ([]*regpb.Service, error) {
	serviceMap := make(map[string]bool)
	entries := make(chan *mdns.ServiceEntry, 10)
	done := make(chan bool)

	p := mdns.DefaultParams("_services")
	// set context with timeout
	var cancel context.CancelFunc
	p.Context, cancel = context.WithTimeout(context.Background(), m.opts.Timeout)
	defer cancel()
	// set entries channel
	p.Entries = entries
	// set domain
	p.Domain = m.domain

	var services []*regpb.Service

	go func() {
		for {
			select {
			case e := <-entries:
				if e.TTL == 0 {
					continue
				}

				if !strings.HasSuffix(e.Name, p.Domain+".") {
					continue
				}

				txt, err := decode(e.InfoFields)
				if err != nil {
					continue
				}

				name := strings.TrimSuffix(e.Name, "."+p.Service+"."+p.Domain+".")
				if !serviceMap[name] {
					serviceMap[name] = true
					services = append(services, &regpb.Service{
						Name:      name,
						Version:   txt.Version,
						Metadata:  txt.Metadata,
						Nodes:     txt.node,
						Endpoints: txt.Endpoints,
					})
				}
			case <-p.Context.Done():
				close(done)
				return
			}
		}
	}()

	// execute query
	if err := mdns.Query(p); err != nil {
		return nil, err
	}

	// wait till done
	<-done

	return services, nil
}

func (m *mdnsRegistry) Watch(opts ...WatchOption) (Watcher, error) {
	var wo WatchOptions
	for _, o := range opts {
		o(&wo)
	}

	md := &mdnsWatcher{
		id:       uuid.New().String(),
		wo:       wo,
		ch:       make(chan *mdns.ServiceEntry, 32),
		exit:     make(chan struct{}),
		domain:   m.domain,
		registry: m,
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()

	// save the watcher
	m.watchers[md.id] = md

	// check of the listener exists
	if m.listener != nil {
		return md, nil
	}

	// start the listener
	go func() {
		// go to infinity
		for {
			m.mtx.Lock()

			// just return if there are no watchers
			if len(m.watchers) == 0 {
				m.listener = nil
				m.mtx.Unlock()
				return
			}

			// check existing listener
			if m.listener != nil {
				m.mtx.Unlock()
				return
			}

			// reset the listener
			exit := make(chan struct{})
			ch := make(chan *mdns.ServiceEntry, 32)
			m.listener = ch

			m.mtx.Unlock()

			// send messages to the watchers
			go func() {
				send := func(w *mdnsWatcher, e *mdns.ServiceEntry) {
					select {
					case w.ch <- e:
					default:
					}
				}

				for {
					select {
					case <-exit:
						return
					case e, ok := <-ch:
						if !ok {
							return
						}
						m.mtx.RLock()
						// send service entry to all watchers
						for _, w := range m.watchers {
							send(w, e)
						}
						m.mtx.RUnlock()
					}
				}

			}()

			// start listening, blocking call
			mdns.Listen(ch, exit)

			// mdns.Listen has unblocked
			// kill the saved listener
			m.mtx.Lock()
			m.listener = nil
			close(ch)
			m.mtx.Unlock()
		}
	}()

	return md, nil
}

func (m *mdnsRegistry) String() string {
	return "mdns"
}

func (m *mdnsWatcher) Next() (*regpb.Result, error) {
	for {
		select {
		case e := <-m.ch:
			txt, err := decode(e.InfoFields)
			if err != nil {
				continue
			}

			if len(txt.Service) == 0 || len(txt.Version) == 0 {
				continue
			}

			// Filter watch options
			// wo.Service: Only keep services we care about
			if len(m.wo.Service) > 0 && txt.Service != m.wo.Service {
				continue
			}
			var action string
			if e.TTL == 0 {
				action = "delete"
			} else {
				action = "create"
			}

			service := &regpb.Service{
				Name:      txt.Service,
				Version:   txt.Version,
				Endpoints: txt.Endpoints,
			}

			// skip anything without the domain we care about
			suffix := fmt.Sprintf(".%s.%s.", service.Name, m.domain)
			if !strings.HasSuffix(e.Name, suffix) {
				continue
			}

			var addr string
			if len(e.AddrV4) > 0 {
				addr = e.AddrV4.String()
			} else if len(e.AddrV6) > 0 {
				addr = "[" + e.AddrV6.String() + "]"
			} else {
				addr = e.Addr.String()
			}

			service.Nodes = append(service.Nodes, &regpb.Node{
				Id:       strings.TrimSuffix(e.Name, suffix),
				Address:  fmt.Sprintf("%s:%d", addr, e.Port),
				Metadata: txt.Metadata,
			})

			return &regpb.Result{
				Action:  action,
				Service: service,
			}, nil
		case <-m.exit:
			return nil, ErrWatcherStopped
		}
	}
}

func (m *mdnsWatcher) Stop() {
	select {
	case <-m.exit:
		return
	default:
		close(m.exit)
		// remove self from the registry
		m.registry.mtx.Lock()
		delete(m.registry.watchers, m.id)
		m.registry.mtx.Unlock()
	}
}

// NewRegistry returns a new default registry which is mdns
func NewRegistry(opts ...Option) Registry {
	return newRegistry(opts...)
}
