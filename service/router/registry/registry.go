// Copyright 2020 The vine Authors
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
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/registry"
	"github.com/lack-io/vine/service/router"
)

var (
	// AdvertiseEventsTick is time interval in which the router advertises route updates
	AdvertiseEventsTick = 10 * time.Second
	// DefaultAdvertTTL is default advertisement TTL
	DefaultAdvertTTL = 2 * time.Minute
)

// regRouter implements router.Router
type regRouter struct {
	sync.RWMutex

	running   bool
	table     *table
	options   router.Options
	exit      chan bool
	eventChan chan *router.Event

	// advert subscribers
	sub         sync.RWMutex
	subscribers map[string]chan *router.Advert
}

// newRouter creates new router and returns it
func newRouter(opts ...router.Option) router.Router {
	// get default options
	options := router.DefaultOptions()

	// apply requested options
	for _, o := range opts {
		o(&options)
	}

	return &regRouter{
		options:     options,
		table:       newTable(),
		subscribers: make(map[string]chan *router.Advert),
	}
}

// Init initializes router with given options
func (r *regRouter) Init(opts ...router.Option) error {
	r.Lock()
	defer r.Unlock()

	for _, o := range opts {
		o(&r.options)
	}

	return nil
}

// Options returns router options
func (r *regRouter) Options() router.Options {
	r.RLock()
	defer r.RUnlock()

	options := r.options

	return options
}

// Table returns routing table
func (r *regRouter) Table() router.Table {
	return r.table
}

// manageRoute applies action on a given route
func (r *regRouter) manageRoute(route router.Route, action string) error {
	switch action {
	case "create":
		if err := r.table.Create(route); err != nil && err != router.ErrDuplicateRoute {
			return fmt.Errorf("failed adding route for service %s: %s", route.Service, err)
		}
	case "delete":
		if err := r.table.Delete(route); err != nil && err != router.ErrRouteNotFound {
			return fmt.Errorf("failed deleting route for service %s: %s", route.Service, err)
		}
	case "update":
		if err := r.table.Update(route); err != nil {
			return fmt.Errorf("failed updating route for service %s: %s", route.Service, err)
		}
	default:
		return fmt.Errorf("failed to manage route for service %s: unknown action %s", route.Service, action)
	}

	return nil
}

// manageServiceRoutes applies action to all routes of the service.
// It returns error of the action fails with error.
func (r *regRouter) manageRoutes(service *registry.Service, action string) error {
	// action is the routing table action
	action = strings.ToLower(action)

	// take route action on each service node
	for _, node := range service.Nodes {
		route := router.Route{
			Service: service.Name,
			Address: node.Address,
			Gateway: "",
			Network: r.options.Network,
			Router:  r.options.Id,
			Link:    router.DefaultLink,
			Metric:  router.DefaultLocalMetric,
		}

		if err := r.manageRoute(route, action); err != nil {
			return err
		}
	}

	return nil
}

// manageRegistryRoutes applies action to all routes of each service found in the registry.
// It returns error if either the services failed to be listed or the routing table action fails.
func (r *regRouter) manageRegistryRoutes(reg registry.Registry, action string) error {
	services, err := reg.ListServices()
	if err != nil {
		return fmt.Errorf("failed listing services: %v", err)
	}

	// add each service node as a separate route
	for _, service := range services {
		// get the service to retrieve all its info
		srvs, err := reg.GetService(service.Name)
		if err != nil {
			continue
		}
		// manage the routes for all returned services
		for _, srv := range srvs {
			if err := r.manageRoutes(srv, action); err != nil {
				return err
			}
		}
	}

	return nil
}

// watchRegistry watches registry and updates routing table based on the received events.
// It returns error if either the registry watcher fails with error or if the routing table update fails.
func (r *regRouter) watchRegistry(w registry.Watcher) error {
	exit := make(chan bool)

	defer func() {
		close(exit)
	}()

	go func() {
		defer w.Stop()

		select {
		case <-exit:
			return
		case <-r.exit:
			return
		}
	}()

	for {
		res, err := w.Next()
		if err != nil {
			if err != registry.ErrWatcherStopped {
				return err
			}
			break
		}

		if err := r.manageRoutes(res.Service, res.Action); err != nil {
			return err
		}
	}

	return nil
}

// watchTable watches routing table entries and either adds or deletes locally registered service to/from network registry
// It returns error if the locally registered services either fails to be added/deleted to/from network registry.
func (r *regRouter) watchTable(w router.Watcher) error {
	exit := make(chan bool)

	defer func() {
		close(exit)
	}()

	// wait in the background for the router to stop
	// when the router stops, stop the watcher and exit
	go func() {
		defer w.Stop()

		select {
		case <-r.exit:
			return
		case <-exit:
			return
		}
	}()

	for {
		event, err := w.Next()
		if err != nil {
			if err != router.ErrWatcherStopped {
				return err
			}
			break
		}

		select {
		case <-r.exit:
			close(r.eventChan)
			return nil
		case r.eventChan <- event:
			// process event
		}
	}

	return nil
}

// publishAdvert publishes router advert to advert channel
func (r *regRouter) publishAdvert(advType router.AdvertType, events []*router.Event) {
	a := &router.Advert{
		Id:        r.options.Id,
		Type:      advType,
		TTL:       DefaultAdvertTTL,
		Timestamp: time.Now(),
		Events:    events,
	}

	r.sub.RLock()
	for _, sub := range r.subscribers {
		// now send the message
		select {
		case sub <- a:
		case <-r.exit:
			r.sub.RUnlock()
			return
		}
	}
	r.sub.RUnlock()
}

// adverts maintains a map of router adverts
type adverts map[uint64]*router.Event

// advertiseEvents advertises routing table events
// It suppresses unhealthy flapping events and advertises healthy events upstream.
func (r *regRouter) advertiseEvents() error {
	// ticker to periodically scan event for advertising
	ticker := time.NewTicker(AdvertiseEventsTick)
	defer ticker.Stop()

	// adverts is a map of advert events
	adverts := make(adverts)

	// routing table watcher
	w, err := r.Watch()
	if err != nil {
		return err
	}
	defer w.Stop()

	go func() {
		var err error

		for {
			select {
			case <-r.exit:
				return
			default:
				if w == nil {
					// routing table watcher
					w, err = r.Watch()
					if err != nil {
						logger.Errorf("Error creating watcher: %v", err)
						time.Sleep(time.Second)
						continue
					}
				}

				if err := r.watchTable(w); err != nil {
					logger.Errorf("Error watching table: %v", err)
					time.Sleep(time.Second)
				}

				if w != nil {
					// reset
					w.Stop()
					w = nil
				}
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			// If we're not advertising any events then sip processing them entirely
			if r.options.Advertise == router.AdvertiseNone {
				continue
			}

			var events []*router.Event

			// collect all events which are not flapping
			for key, event := range adverts {
				// if we only advertise local routes skip processing anything not link local
				if r.options.Advertise == router.AdvertiseLocal && event.Route.Link != "local" {
					continue
				}

				// copy the event and append
				e := new(router.Event)
				// this is ok, because router.Event only contains builtin types
				// and no references so this creates a deep copy of struct Event
				*e = *event
				events = append(events, e)
				// delete the advert from adverts
				delete(adverts, key)
			}

			// advertise events to subscribers
			if len(events) > 0 {
				logger.Debugf("Router publishing %d events", len(events))
				go r.publishAdvert(router.RouteUpdate, events)
			}
		case e := <-r.eventChan:
			// if event is nil, continue
			if e == nil {
				continue
			}

			// If we're not advertising any events then skip processing them entirely
			if r.options.Advertise == router.AdvertiseNone {
				continue
			}

			// if we only advertise local routes skip processing anything not link local
			if r.options.Advertise == router.AdvertiseLocal && e.Route.Link != "local" {
				continue
			}

			logger.Debugf("Router processing table event %s for service %s %s", e.Type, e.Route.Service, e.Route.Address)

			// check if we have already registered the route
			hash := e.Route.Hash()
			ev, ok := adverts[hash]
			if !ok {
				ev = e
				adverts[hash] = e
				continue
			}

			// override the route event only if the previous event was different
			if ev.Type != e.Type {
				ev = e
			}
		case <-r.exit:
			if w != nil {
				w.Stop()
			}
			return nil
		}
	}
}

// drain all the events, only called on Stop
func (r *regRouter) drain() {
	for {
		select {
		case <-r.eventChan:
		default:
			return
		}
	}
}

// Start starts the router
func (r *regRouter) Start() error {
	r.Lock()
	defer r.Unlock()

	if r.running {
		return nil
	}

	// add all local service routes into the routing table
	if err := r.manageRegistryRoutes(r.options.Registry, "create"); err != nil {
		return fmt.Errorf("failed adding registry routes: %s", err)
	}

	// add default gateway into routing table
	if r.options.Gateway != "" {
		// note, the only non-default value is the gateway
		route := router.Route{
			Service: "*",
			Address: "*",
			Gateway: r.options.Gateway,
			Network: "*",
			Router:  r.options.Id,
			Link:    router.DefaultLink,
			Metric:  router.DefaultLocalMetric,
		}
		if err := r.table.Create(route); err != nil {
			return fmt.Errorf("failed adding default gateway route: %s", err)
		}
	}

	// create error and exit channels
	r.exit = make(chan bool)

	// registry watcher
	w, err := r.options.Registry.Watch()
	if err != nil {
		return fmt.Errorf("failed creating registry watcher: %v", err)
	}

	go func() {
		var err error

		for {
			select {
			case <-r.exit:
				if w != nil {
					w.Stop()
				}
				return
			default:
				if w == nil {
					w, err = r.options.Registry.Watch()
					if err != nil {
						logger.Errorf("failed creating registry watcher: %v", err)
						time.Sleep(time.Second)
						continue
					}
				}

				if err := r.watchRegistry(w); err != nil {
					logger.Errorf("Error watching the registry: %v", err)
					time.Sleep(time.Second)
				}

				if w != nil {
					w.Stop()
					w = nil
				}
			}
		}
	}()

	r.running = true

	return nil
}

// Advertise stars advertising the routes to the network and returns the advertisements channel to consume from.
// If the router is already advertising it returns the channel to consume from.
// It returns error if either the router is not running or if the routing table fails to list the routes to advertise.
func (r *regRouter) Advertise() (<-chan *router.Advert, error) {
	r.Lock()
	defer r.Unlock()

	if !r.running {
		return nil, errors.New("not running")
	}

	// already advertising
	if r.eventChan != nil {
		advertChan := make(chan *router.Advert, 128)
		r.subscribers[uuid.New().String()] = advertChan
		return advertChan, nil
	}

	// list all the routes and pack them into even slice to advertise
	events, err := r.flushRouteEvents(router.Create)
	if err != nil {
		return nil, fmt.Errorf("failed to flush routes: %s", err)
	}

	// create event channels
	r.eventChan = make(chan *router.Event)

	// create advert channel
	advertChan := make(chan *router.Advert, 128)
	r.subscribers[uuid.New().String()] = advertChan

	// advertise your presence
	go r.publishAdvert(router.Announce, events)

	go func() {
		select {
		case <-r.exit:
			return
		default:
			if err := r.advertiseEvents(); err != nil {
				logger.Errorf("Error adveritising events: %v", err)
			}
		}
	}()

	return advertChan, nil

}

// Process updates the routing table using the advertised values
func (r *regRouter) Process(a *router.Advert) error {
	// NOTE: event sorting might not be necessary
	// copy update events intp new slices
	events := make([]*router.Event, len(a.Events))
	copy(events, a.Events)
	// sort events by timestamp
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	logger.Debugf("Router %s processing advert from: %s", r.options.Id, a.Id)

	for _, event := range events {
		// skip if the router is the origin of this route
		if event.Route.Router == r.options.Id {
			logger.Debugf("Router skipping processing its own route: %s", r.options.Id)
			continue
		}
		// create a copy of the route
		route := event.Route
		action := event.Type

		logger.Debugf("Router %s applying %s from router %s for service %s %s", r.options.Id, action, route.Router, route.Service, route.Address)

		if err := r.manageRoute(route, action.String()); err != nil {
			return fmt.Errorf("failed applying action %s to routing table: %s", action, err)
		}
	}

	return nil
}

// flushRouteEvents returns a slice of events, one per each route in the routing table
func (r *regRouter) flushRouteEvents(evType router.EventType) ([]*router.Event, error) {
	// get a list of routes for each service in our routing table
	// for the configured advertising strategy
	q := []router.QueryOption{
		router.QueryStrategy(r.options.Advertise),
	}

	routes, err := r.Table().Query(q...)
	if err != nil && err != router.ErrRouteNotFound {
		return nil, err
	}

	logger.Debugf("Router advertising %d routes with strategy %s", len(routes), r.options.Advertise)

	// build a list of events to advertise
	events := make([]*router.Event, len(routes))
	var i int

	for _, route := range routes {
		event := &router.Event{
			Type:      evType,
			Timestamp: time.Now(),
			Route:     route,
		}
		events[i] = event
		i++
	}

	return events, nil
}

// Lookup routes in the routing table
func (r *regRouter) Lookup(q ...router.QueryOption) ([]router.Route, error) {
	return r.table.Query(q...)
}

// Watch routes
func (r *regRouter) Watch(opts ...router.WatchOption) (router.Watcher, error) {
	return r.table.Watch(opts...)
}

// Stop stops the router
func (r *regRouter) Stop() error {
	r.Lock()
	defer r.Unlock()

	select {
	case <-r.exit:
		return nil
	default:
		close(r.exit)

		// extract the events
		r.drain()

		r.sub.Lock()
		// close advert subscribers
		for id, sub := range r.subscribers {
			// close the channel
			close(sub)
			// delete the subscriber
			delete(r.subscribers, id)
		}
		r.sub.Unlock()
	}

	// remove event chan
	r.eventChan = nil

	return nil
}

// String prints debugging information about router
func (r *regRouter) String() string {
	return "registry"
}

// NewRouter creates new Router and returns it
func NewRouter(opts ...router.Option) router.Router {
	return newRouter(opts...)
}
