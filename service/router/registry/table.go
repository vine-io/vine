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
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/router"
)

// table is an in-memory routing table
type table struct {
	sync.RWMutex
	// routes stores service routes
	routes map[string]map[uint64]router.Route
	// watchers stores table watchers
	watchers map[string]*tableWatcher
}

// newTable creates a new routing table and returns it
func newTable(opts ...router.Option) *table {
	return &table{
		routes:   make(map[string]map[uint64]router.Route),
		watchers: make(map[string]*tableWatcher),
	}
}

// sendEvent sends events to all subscribed watchers
func (t *table) sendEvent(e *router.Event) {
	t.RLock()
	defer t.RUnlock()

	for len(e.Id) == 0 {
		e.Id = uuid.New().String()
	}

	for _, w := range t.watchers {
		select {
		case w.resChan <- e:
		case <-w.done:
		// don't block forever
		case <-time.After(time.Second):
		}
	}
}

// Create creates new route in the routing table
func (t *table) Create(r router.Route) error {
	service := r.Service
	sum := r.Hash()

	t.Lock()
	defer t.Unlock()

	// check if there any routes in the table for route destination
	if _, ok := t.routes[service]; !ok {
		t.routes[service] = make(map[uint64]router.Route)
	}

	// add new route to the table for the route destination
	if _, ok := t.routes[service][sum]; !ok {
		t.routes[service][sum] = r
		logger.Debugf("Router emitting %s for route: %s", router.Create, r.Address)
		go t.sendEvent(&router.Event{Type: router.Create, Timestamp: time.Now(), Route: r})
		return nil
	}

	return router.ErrDuplicateRoute
}

// Delete deletes the route from the routing table
func (t *table) Delete(r router.Route) error {
	service := r.Service
	sum := r.Hash()

	t.Lock()
	defer t.Unlock()

	if _, ok := t.routes[service]; !ok {
		return router.ErrRouteNotFound
	}

	if _, ok := t.routes[service][sum]; !ok {
		return router.ErrRouteNotFound
	}

	delete(t.routes[service], sum)
	logger.Debugf("Router emitting %s for route: %s", router.Delete, r.Address)
	go t.sendEvent(&router.Event{Type: router.Delete, Timestamp: time.Now(), Route: r})

	return nil
}

// Update updates routing table with the new route
func (t *table) Update(r router.Route) error {
	service := r.Service
	sum := r.Hash()

	t.Lock()
	defer t.Unlock()

	// check if the route destination has any routes in the table
	if _, ok := t.routes[service]; !ok {
		t.routes[service] = make(map[uint64]router.Route)
	}

	if _, ok := t.routes[service][sum]; !ok {
		t.routes[service][sum] = r
		logger.Debugf("Router emitting %s for route: %s", router.Update, r.Address)
		go t.sendEvent(&router.Event{Type: router.Update, Timestamp: time.Now(), Route: r})
		return nil
	}

	// just update route, but dont emit Update event
	t.routes[service][sum] = r

	return nil
}

// List returns a list of all routes in the table
func (t *table) List() ([]router.Route, error) {
	t.RLock()
	defer t.RUnlock()

	var routes []router.Route
	for _, rmap := range t.routes {
		for _, route := range rmap {
			routes = append(routes, route)
		}
	}

	return routes, nil
}

// isMatch checks if the route matches given query options
func isMatch(route router.Route, address, gateway, network, _router string, strategy router.Strategy) bool {
	// matches the values provided
	match := func(a, b string) bool {
		return a == "*" || a == b
	}

	// a simple struct to hold our values
	type compare struct {
		a string
		b string
	}

	// by default assume we are querying all routes
	link := "*"
	// if AdvertiseLocal change the link query accordingly
	if strategy == router.AdvertiseLocal {
		link = "local"
	}

	// compare the following values
	values := []compare{
		{gateway, route.Gateway},
		{network, route.Network},
		{_router, route.Router},
		{address, route.Address},
		{link, route.Link},
	}

	for _, v := range values {
		// attempt to match each value
		if !match(v.a, v.b) {
			return false
		}
	}

	return true
}

// findRoutes finds all the routes for given network and router and returns them
func findRoutes(routes map[uint64]router.Route, address, gateway, network, _router string, strategy router.Strategy) []router.Route {
	// routeMap stores the routes we're going to advertise
	routeMap := make(map[string][]router.Route)

	for _, route := range routes {
		if isMatch(route, address, gateway, network, _router, strategy) {
			// add matching route to the routeMap
			routeKey := route.Service + "@" + route.Network
			// append the first found route to routeMap
			_, ok := routeMap[routeKey]
			if !ok {
				routeMap[routeKey] = append(routeMap[routeKey], route)
				continue
			}

			// if AdvertiseAll, keep appending
			if strategy == router.AdvertiseAll || strategy == router.AdvertiseLocal {
				routeMap[routeKey] = append(routeMap[routeKey], route)
				continue
			}

			// now we're going to find the best routes
			if strategy == router.AdvertiseBest {
				// if the current optimal route metric is higher then routing table route, replace it
				if len(routeMap[routeKey]) > 0 {
					// NOTE: we know than when AdvertiseBest is set, we only ever have one item in current
					if routeMap[routeKey][0].Metric > route.Metric {
						routeMap[routeKey][0] = route
						continue
					}
				}
			}
		}
	}

	var results []router.Route
	for _, route := range routeMap {
		results = append(results, route...)
	}

	return results
}

// Lookup queries routing table and returns all routes that match the lookup query
func (t *table) Query(q ...router.QueryOption) ([]router.Route, error) {
	t.RLock()
	defer t.RUnlock()

	// create new query options
	opts := router.NewQuery(q...)

	// creates a cwslicelist of query results
	results := make([]router.Route, 0, len(t.routes))

	// if No routes are queried, return early
	if opts.Strategy == router.AdvertiseNone {
		return results, nil
	}

	if opts.Service != "*" {
		if _, ok := t.routes[opts.Service]; !ok {
			return nil, router.ErrRouteNotFound
		}
		return findRoutes(t.routes[opts.Service], opts.Address, opts.Gateway, opts.Network, opts.Router, opts.Strategy), nil
	}

	// search through all destinations
	for _, routes := range t.routes {
		results = append(results, findRoutes(routes, opts.Address, opts.Gateway, opts.Network, opts.Router, opts.Strategy)...)
	}

	return results, nil
}

// Watch returns routing table entry watcher
func (t *table) Watch(opts ...router.WatchOption) (router.Watcher, error) {
	// by default watch everything
	wopts := router.WatchOptions{
		Service: "*",
	}

	for _, o := range opts {
		o(&wopts)
	}

	w := &tableWatcher{
		id:      uuid.New().String(),
		opts:    wopts,
		resChan: make(chan *router.Event, 10),
		done:    make(chan struct{}),
	}

	// when the watcher is stopped delete it
	go func() {
		<-w.done
		t.Lock()
		delete(t.watchers, w.id)
		t.Unlock()
	}()

	// save the watcher
	t.Lock()
	t.watchers[w.id] = w
	t.Unlock()

	return w, nil
}
