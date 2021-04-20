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

package registry

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	rr "github.com/lack-io/vine/core/router"
	log "github.com/lack-io/vine/lib/logger"
)

var (
	// ErrRouteNotFound is returned when no route was found in the routing table
	ErrRouteNotFound = errors.New("route not found")
	// ErrDuplicateRoute is returned when the route already exists
	ErrDuplicateRoute = errors.New("duplicate route")
)

// table is an in-memory routing table
type table struct {
	sync.RWMutex
	// routes stores service routes
	routes map[string]map[uint64]rr.Route
	// watchers stores table watchers
	watchers map[string]*tableWatcher
}

// newTable creates a new routing table and returns it
func newTable(opts ...rr.Option) *table {
	return &table{
		routes:   make(map[string]map[uint64]rr.Route),
		watchers: make(map[string]*tableWatcher),
	}
}

// sendEvent sends events to all subscribed watchers
func (t *table) sendEvent(e *rr.Event) {
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
func (t *table) Create(route rr.Route) error {
	service := route.Service
	sum := route.Hash()

	t.Lock()
	defer t.Unlock()

	// check if there any routes in the table for route destination
	if _, ok := t.routes[service]; !ok {
		t.routes[service] = make(map[uint64]rr.Route)
	}

	// add new route to the table for the route destination
	if _, ok := t.routes[service][sum]; !ok {
		t.routes[service][sum] = route
		log.Debugf("Router emitting %s for route: %s", rr.Create, rr.Address)
		go t.sendEvent(&rr.Event{Type: rr.Create, Timestamp: time.Now(), Route: route})
		return nil
	}

	return ErrDuplicateRoute
}

// Delete deletes the route from the routing table
func (t *table) Delete(route rr.Route) error {
	service := route.Service
	sum := route.Hash()

	t.Lock()
	defer t.Unlock()

	if _, ok := t.routes[service]; !ok {
		return ErrRouteNotFound
	}

	if _, ok := t.routes[service][sum]; !ok {
		return ErrRouteNotFound
	}

	delete(t.routes[service], sum)
	log.Debugf("Router emitting %s for route: %s", rr.Delete, rr.Address)
	go t.sendEvent(&rr.Event{Type: rr.Delete, Timestamp: time.Now(), Route: route})

	return nil
}

// Update updates routing table with the new route
func (t *table) Update(route rr.Route) error {
	service := route.Service
	sum := route.Hash()

	t.Lock()
	defer t.Unlock()

	// check if the route destination has any routes in the table
	if _, ok := t.routes[service]; !ok {
		t.routes[service] = make(map[uint64]rr.Route)
	}

	if _, ok := t.routes[service][sum]; !ok {
		t.routes[service][sum] = route
		log.Debugf("Router emitting %s for route: %s", rr.Update, rr.Address)
		go t.sendEvent(&rr.Event{Type: rr.Update, Timestamp: time.Now(), Route: route})
		return nil
	}

	// just update route, but dont emit Update event
	t.routes[service][sum] = route

	return nil
}

// List returns a list of all routes in the table
func (t *table) List() ([]rr.Route, error) {
	t.RLock()
	defer t.RUnlock()

	var routes []rr.Route
	for _, rmap := range t.routes {
		for _, route := range rmap {
			routes = append(routes, route)
		}
	}

	return routes, nil
}

// isMatch checks if the route matches given query options
func isMatch(route rr.Route, address, gateway, network, router string, strategy rr.Strategy) bool {
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
	if strategy == rr.AdvertiseLocal {
		link = "local"
	}

	// compare the following values
	values := []compare{
		{gateway, route.Gateway},
		{network, route.Network},
		{router, route.Router},
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
func findRoutes(routes map[uint64]rr.Route, address, gateway, network, router string, strategy rr.Strategy) []rr.Route {
	// routeMap stores the routes we're going to advertise
	routeMap := make(map[string][]rr.Route)

	for _, route := range routes {
		if isMatch(route, address, gateway, network, router, strategy) {
			// add matching route to the routeMap
			routeKey := route.Service + "@" + route.Network
			// append the first found route to routeMap
			_, ok := routeMap[routeKey]
			if !ok {
				routeMap[routeKey] = append(routeMap[routeKey], route)
				continue
			}

			// if AdvertiseAll, keep appending
			if strategy == rr.AdvertiseAll || strategy == rr.AdvertiseLocal {
				routeMap[routeKey] = append(routeMap[routeKey], route)
				continue
			}

			// now we're going to find the best routes
			if strategy == rr.AdvertiseBest {
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

	var results []rr.Route
	for _, route := range routeMap {
		results = append(results, route...)
	}

	return results
}

// Lookup queries routing table and returns all routes that match the lookup query
func (t *table) Query(q ...rr.QueryOption) ([]rr.Route, error) {
	t.RLock()
	defer t.RUnlock()

	// create new query options
	opts := rr.NewQuery(q...)

	// creates a cwslicelist of query results
	results := make([]rr.Route, 0, len(t.routes))

	// if No routes are queried, return early
	if opts.Strategy == rr.AdvertiseNone {
		return results, nil
	}

	if opts.Service != "*" {
		if _, ok := t.routes[opts.Service]; !ok {
			return nil, ErrRouteNotFound
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
func (t *table) Watch(opts ...rr.WatchOption) (rr.Watcher, error) {
	// by default watch everything
	wopts := rr.WatchOptions{
		Service: "*",
	}

	for _, o := range opts {
		o(&wopts)
	}

	w := &tableWatcher{
		id:      uuid.New().String(),
		opts:    wopts,
		resChan: make(chan *rr.Event, 10),
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

// tableWatcher implements routing table router
type tableWatcher struct {
	sync.RWMutex
	id      string
	opts    rr.WatchOptions
	resChan chan *rr.Event
	done    chan struct{}
}

// Next returns the next noticed action taken on table
// TODO: right now we only allow to watch particular service
func (w *tableWatcher) Next() (*rr.Event, error) {
	for {
		select {
		case res := <-w.resChan:
			switch w.opts.Service {
			case res.Route.Service, "*":
				return res, nil
			default:
				continue
			}
		case <-w.done:
			return nil, rr.ErrWatcherStopped
		}
	}
}

// Chan returns watcher events channel
func (w *tableWatcher) Chan() (<-chan *rr.Event, error) {
	return w.resChan, nil
}

// Stop stops routing table watcher
func (w *tableWatcher) Stop() {
	w.Lock()
	defer w.Unlock()

	select {
	case <-w.done:
		return
	default:
		close(w.done)
	}
}
