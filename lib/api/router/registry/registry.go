// MIT License
//
// Copyright (c) 2020 Lack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package registry provides a dynamic api service router
package registry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/vine-io/vine/core/registry"
	"github.com/vine-io/vine/core/registry/cache"
	"github.com/vine-io/vine/lib/api"
	"github.com/vine-io/vine/lib/api/router"
	"github.com/vine-io/vine/lib/api/router/util"
	"github.com/vine-io/vine/lib/logger"
	"github.com/vine-io/vine/util/context/metadata"
)

// endpoint struct, that holds compiled pcre
type endpoint struct {
	hostregs []*regexp.Regexp
	pathregs []util.Pattern
	pcreregs []*regexp.Regexp
}

// router is the default router
type registryRouter struct {
	exit chan bool
	opts router.Options

	// registry cache
	rc cache.Cache

	sync.RWMutex
	eps map[string]*api.Service
	// compiled regexp for host and path
	ceps map[string]*endpoint
}

func (r *registryRouter) isClosed() bool {
	select {
	case <-r.exit:
		return true
	default:
		return false
	}
}

// refresh list of api services
func (r *registryRouter) refresh() {
	var attempts int

	ctx := context.TODO()
	for {
		services, err := r.opts.Registry.ListServices(ctx)
		if err != nil {
			attempts++
			logger.Errorf("unable to list services: %v", err)
			time.Sleep(time.Duration(attempts) * time.Second)
			continue
		}

		attempts = 0

		// for each service, get service and store endpoints
		for _, s := range services {
			service, err := r.rc.GetService(ctx, s.Name)
			if err != nil {
				logger.Errorf("unable to get service: %v", err)
				continue
			}
			r.store(service)
		}

		// refresh list in 10 minutes... cruft
		// use registry watching
		select {
		case <-time.After(time.Minute * 10):
		case <-r.exit:
			return
		}
	}
}

// process watch event
func (r *registryRouter) process(res *registry.Result) {
	// skip these things
	if res == nil || res.Service == nil {
		return
	}

	// get entry from cache
	service, err := r.rc.GetService(context.TODO(), res.Service.Name)
	if err != nil {
		logger.Errorf("unable to get service '%s': %v", res.Service.Name, err)
		return
	}

	// update our local endpoints
	r.store(service)
}

// store local endpoint cache
func (r *registryRouter) store(services []*registry.Service) {
	// endpoints
	eps := map[string]*api.Service{}

	// services
	names := map[string]bool{}

	// create a new endpoint mapping
	for _, service := range services {
		// set names we need later
		names[service.Name] = true

		// map per endpoint
		for _, sep := range service.Endpoints {
			// create a key service:endpoint_name
			key := fmt.Sprintf("%s.%s", service.Name, sep.Name)
			// decode endpoint
			end := api.Decode(sep.Metadata)

			// if we got nothing skip
			if err := api.Validate(end); err != nil {
				logger.Tracef("endpoint validation failed: %+v", err)
				continue
			}

			// try to get endpoint
			ep, ok := eps[key]
			if !ok {
				ep = &api.Service{Name: service.Name}
			}

			// overwrite the endpoint
			ep.Endpoint = end
			// append services
			ep.Services = append(ep.Services, service)
			// store it
			eps[key] = ep
		}
	}

	r.Lock()
	defer r.Unlock()

	// delete any existing eps for services we know
	for key, service := range r.eps {
		// skip what we don't care about
		if !names[service.Name] {
			continue
		}

		// ok we know this thing
		// delete delete delete
		delete(r.eps, key)
	}

	// now set the eps we have
	for name, ep := range eps {
		r.eps[name] = ep
		cep := &endpoint{}

		for _, h := range ep.Endpoint.Host {
			if h == "" || h == "*" {
				continue
			}
			hostreg, err := regexp.CompilePOSIX(h)
			if err != nil {
				logger.Tracef("endpoint have invalid host regexp: %+v", err)
				continue
			}
			cep.hostregs = append(cep.hostregs, hostreg)
		}

		for _, p := range ep.Endpoint.Path {
			var pcreok bool

			if p[0] == '^' && p[len(p)-1] == '$' {
				pcrereg, err := regexp.CompilePOSIX(p)
				if err == nil {
					cep.pcreregs = append(cep.pcreregs, pcrereg)
					pcreok = true
				}
			}

			rule, err := util.Parse(p)
			if err != nil && !pcreok {
				logger.Tracef("endpoint have invalid path pattern: %+v", err)
				continue
			} else if err != nil && pcreok {
				continue
			}

			tpl := rule.Compile()
			pathreg, err := util.NewPattern(tpl.Version, tpl.OpCodes, tpl.Pool, "")
			if err != nil {
				logger.Tracef("endpoint have invalid path pattern: %+v", err)
				continue
			}
			cep.pathregs = append(cep.pathregs, pathreg)
		}

		r.ceps[name] = cep
	}
}

// watch for endpoint changes
func (r *registryRouter) watch() {
	var attempts int

	for {
		if r.isClosed() {
			return
		}

		// watch for changes
		w, err := r.opts.Registry.Watch(context.TODO())
		if err != nil {
			attempts++
			logger.Errorf("error watching endpoints: %+v", err)
			time.Sleep(time.Duration(attempts) * time.Second)
			continue
		}

		ch := make(chan bool)

		go func() {
			select {
			case <-ch:
				w.Stop()
			case <-r.exit:
				w.Stop()
			}
		}()

		// reset if we get here
		attempts = 0

		for {
			// process next event
			res, err := w.Next()
			if err != nil {
				logger.Errorf("error getting next endpoint: %+v", err)
				close(ch)
				break
			}
			r.process(res)
		}
	}
}

func (r *registryRouter) Options() router.Options {
	return r.opts
}

func (r *registryRouter) Close() error {
	select {
	case <-r.exit:
		return nil
	default:
		close(r.exit)
		r.rc.Stop()
	}
	return nil
}

func (r *registryRouter) Register(ep *api.Endpoint) error {
	return nil
}

func (r *registryRouter) Deregister(ep *api.Endpoint) error {
	return nil
}

func (r *registryRouter) Endpoint(req *http.Request) (*api.Service, error) {
	if r.isClosed() {
		return nil, errors.New("router closed")
	}

	r.RLock()
	defer r.RUnlock()

	var idx int
	path := req.URL.Path
	if len(path) > 0 && path != "/" {
		idx = 1
	}
	paths := strings.Split(path[idx:], "/")

	// use the first match
	// TODO: weighted matching
	for n, e := range r.eps {
		cep, ok := r.ceps[n]
		if !ok {
			continue
		}
		ep := e.Endpoint
		var mMatch, hMatch, pMatch bool
		// 1. try method
		for _, m := range ep.Method {
			if m == req.Method {
				mMatch = true
				break
			}
		}
		if !mMatch {
			continue
		}
		logger.Debugf("api method match %s", req.Method)

		// 2. try host
		if len(ep.Host) == 0 {
			hMatch = true
		} else {
			for idx, h := range ep.Host {
				if h == "" || h == "*" {
					hMatch = true
					break
				} else {
					if cep.hostregs[idx].MatchString(req.Host) {
						hMatch = true
						break
					}
				}
			}
		}
		if !hMatch {
			continue
		}
		logger.Debugf("api host match %s", req.Host)

		// 3. try path via google.api path matching
		for _, pathreg := range cep.pathregs {
			matches, err := pathreg.Match(paths, "")
			if err != nil {
				logger.Debugf("api gpath not match %s != %v", path, pathreg)
				continue
			}
			logger.Debugf("api gpath match %s = %v", path, pathreg)
			pMatch = true
			ctx := req.Context()
			md, ok := metadata.FromContext(ctx)
			if !ok {
				md = make(metadata.Metadata)
			}
			for k, v := range matches {
				md.Set("x-api-field-"+k, v)
			}
			md.Set("x-api-body", ep.Body)
			*req = *req.Clone(metadata.NewContext(ctx, md))
			break
		}

		if !pMatch {
			// 4. try path via pcre path matching
			for _, pathreg := range cep.pcreregs {
				if !pathreg.MatchString(req.URL.Path) {
					logger.Debugf("api pcre path not match %s != %v", path, pathreg)
					continue
				}
				logger.Debugf("api pcre path match %s != %v", path, pathreg)
				pMatch = true
				break
			}
		}

		if !pMatch {
			continue
		}

		// TODO: Percentage traffic
		// we got here, so its a match
		return e, nil
	}

	// no match
	return nil, errors.New("not found")
}

func (r *registryRouter) Route(req *http.Request) (*api.Service, error) {
	if r.isClosed() {
		return nil, errors.New("router closed")
	}

	// try to get an endpoint
	ep, err := r.Endpoint(req)
	if err == nil {
		return ep, nil
	}

	// error not nil
	// ignore that shit
	// TODO: don't ignore that shit

	// get the service name
	rp, err := r.opts.Resolver.Resolve(req)
	if err != nil {
		return nil, err
	}

	// service name
	name := rp.Name

	ctx := req.Context()
	// get service
	services, err := r.rc.GetService(ctx, name)
	if err != nil {
		return nil, err
	}

	// only use endpoint matching when the meta handler is set aka api.Default
	switch r.opts.Handler {
	// rpc handlers
	case "meta", "api", "rpc":
		handler := r.opts.Handler

		// set default handler to api
		if r.opts.Handler == "meta" {
			handler = "rpc"
		}

		// construct api service
		return &api.Service{
			Name: name,
			Endpoint: &api.Endpoint{
				Name:    rp.Method,
				Handler: handler,
			},
			Services: services,
		}, nil
	// http handler
	case "http", "proxy", "web":
		// construct api service
		return &api.Service{
			Name: name,
			Endpoint: &api.Endpoint{
				Name:    req.URL.String(),
				Handler: r.opts.Handler,
				Host:    []string{req.Host},
				Method:  []string{req.Method},
				Path:    []string{req.URL.Path},
			},
			Services: services,
		}, nil
	}

	return nil, errors.New("unknown handler")
}

func newRouter(opts ...router.Option) *registryRouter {
	options := router.NewOptions(opts...)
	r := &registryRouter{
		exit: make(chan bool),
		opts: options,
		rc:   cache.New(options.Registry),
		eps:  make(map[string]*api.Service),
		ceps: make(map[string]*endpoint),
	}
	go r.watch()
	go r.refresh()
	return r
}

// NewRouter returns the default router
func NewRouter(opts ...router.Option) router.Router {
	return newRouter(opts...)
}
