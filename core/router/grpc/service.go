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

package grpc

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/lack-io/vine/core/client"
	rr "github.com/lack-io/vine/core/router"
	pb "github.com/lack-io/vine/proto/services/router"
)

type svc struct {
	sync.RWMutex
	opts       rr.Options
	callOpts   []client.CallOption
	router     pb.RouterService
	table      *table
	exit       chan bool
	errChan    chan error
	advertChan chan *rr.Advert
}

// NewRouter creates new service router and returns it
func NewRouter(opts ...rr.Option) rr.Router {
	// get default options
	options := rr.DefaultOptions()

	// apply requested options
	for _, o := range opts {
		o(&options)
	}

	// NOTE: might need some client opts here
	cli := client.DefaultClient

	// set options client
	if options.Client != nil {
		cli = options.Client
	}

	// NOTE: should we have Client/Service option in router.Options?
	s := &svc{
		opts:   options,
		router: pb.NewRouterService(rr.DefaultName, cli),
	}

	// set the router address to call
	if len(options.Address) > 0 {
		s.callOpts = []client.CallOption{
			client.WithAddress(options.Address),
		}
	}
	// set the table
	s.table = &table{pb.NewTableService(rr.DefaultName, cli), s.callOpts}

	return s
}

// Init initializes router with given options
func (s *svc) Init(opts ...rr.Option) error {
	s.Lock()
	defer s.Unlock()

	for _, o := range opts {
		o(&s.opts)
	}

	return nil
}

// Options returns router options
func (s *svc) Options() rr.Options {
	s.Lock()
	opts := s.opts
	s.Unlock()

	return opts
}

// Table returns routing table
func (s *svc) Table() rr.Table {
	return s.table
}

// Start starts the service
func (s *svc) Start() error {
	s.Lock()
	defer s.Unlock()
	return nil
}

func (s *svc) advertiseEvents(advertChan chan *rr.Advert, stream pb.Router_AdvertiseService) error {
	go func() {
		<-s.exit
		stream.Close()
	}()

	var advErr error

	for {
		resp, err := stream.Recv()
		if err != nil {
			if err != io.EOF {
				advErr = err
			}
			break
		}

		events := make([]*rr.Event, len(resp.Events))
		for i, event := range resp.Events {
			route := rr.Route{
				Service: event.Route.Service,
				Address: event.Route.Address,
				Gateway: event.Route.Gateway,
				Network: event.Route.Network,
				Link:    event.Route.Link,
				Metric:  event.Route.Metric,
			}

			events[i] = &rr.Event{
				Id:        event.Id,
				Type:      rr.EventType(event.Type),
				Timestamp: time.Unix(0, event.Timestamp),
				Route:     route,
			}
		}

		advert := &rr.Advert{
			Id:        resp.Id,
			Type:      rr.AdvertType(resp.Type),
			Timestamp: time.Unix(0, resp.Timestamp),
			TTL:       time.Duration(resp.Ttl),
			Events:    events,
		}

		select {
		case advertChan <- advert:
		case <-s.exit:
			close(advertChan)
			return nil
		}
	}

	// close the channel on exit
	close(advertChan)

	return advErr
}

// Advertise advertises routes to the network
func (s *svc) Advertise() (<-chan *rr.Advert, error) {
	s.Lock()
	defer s.Unlock()

	stream, err := s.router.Advertise(context.Background(), &pb.Request{}, s.callOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed getting advert stream: %s", err)
	}

	// create advertise and event channels
	advertChan := make(chan *rr.Advert)
	go s.advertiseEvents(advertChan, stream)

	return advertChan, nil
}

// Process processes incoming adverts
func (s *svc) Process(advert *rr.Advert) error {
	events := make([]*pb.Event, 0, len(advert.Events))
	for _, event := range advert.Events {
		route := &pb.Route{
			Service: event.Route.Service,
			Address: event.Route.Address,
			Gateway: event.Route.Gateway,
			Network: event.Route.Network,
			Link:    event.Route.Link,
			Metric:  event.Route.Metric,
		}
		e := &pb.Event{
			Id:        event.Id,
			Type:      pb.EventType(event.Type),
			Timestamp: event.Timestamp.UnixNano(),
			Route:     route,
		}
		events = append(events, e)
	}

	advertReq := &pb.Advert{
		Id:        s.Options().Id,
		Type:      pb.AdvertType(advert.Type),
		Timestamp: advert.Timestamp.UnixNano(),
		Events:    events,
	}

	if _, err := s.router.Process(context.Background(), advertReq, s.callOpts...); err != nil {
		return err
	}

	return nil
}

// Remote router cannot be stopped
func (s *svc) Stop() error {
	s.Lock()
	defer s.Unlock()

	select {
	case <-s.exit:
		return nil
	default:
		close(s.exit)
	}

	return nil
}

// Lookup looks up routes in the routing table and returns them
func (s *svc) Lookup(q ...rr.QueryOption) ([]rr.Route, error) {
	// call the router
	query := rr.NewQuery(q...)

	resp, err := s.router.Lookup(context.Background(), &pb.LookupRequest{
		Query: &pb.Query{
			Service: query.Service,
			Gateway: query.Gateway,
			Network: query.Network,
		},
	}, s.callOpts...)

	// errored out
	if err != nil {
		return nil, err
	}

	routes := make([]rr.Route, len(resp.Routes))
	for i, route := range resp.Routes {
		routes[i] = rr.Route{
			Service: route.Service,
			Address: route.Address,
			Gateway: route.Gateway,
			Network: route.Network,
			Link:    route.Link,
			Metric:  route.Metric,
		}
	}

	return routes, nil
}

// Watch returns a watcher which allows to track updates to the routing table
func (s *svc) Watch(opts ...rr.WatchOption) (rr.Watcher, error) {
	rsp, err := s.router.Watch(context.Background(), &pb.WatchRequest{}, s.callOpts...)
	if err != nil {
		return nil, err
	}
	options := rr.WatchOptions{
		Service: "*",
	}
	for _, o := range opts {
		o(&options)
	}
	return newWatcher(rsp, options)
}

// Returns the router implementation
func (s *svc) String() string {
	return "service"
}
