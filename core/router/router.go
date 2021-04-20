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

package router

import "time"

var (
	// DefaultAddress is default router address
	DefaultAddress = ":9093"
	// DefaultName is default router service name
	DefaultName = "go.vine.router"
	// DefaultNetwork is default router service name
	DefaultNetwork = "go.vine"
	// DefaultRouter is default network router
	DefaultRouter Router
)

// Router is an interface for a routing control plane
type Router interface {
	// Init initializes the router with options
	Init(...Option) error
	// Options returns the router options
	Options() Options
	// Table the routing table
	Table() Table
	// Advertise advertises routes
	Advertise() (<-chan *Advert, error)
	// Process process incoming adverts
	Process(*Advert) error
	// Lookup queries routes in the routing table
	Lookup(...QueryOption) ([]Route, error)
	// Watch returns a watcher which traces updates to the routing table
	Watch(opts ...WatchOption) (Watcher, error)
	// Start starts the router
	Start() error
	// Stop stops the router
	Stop() error
	// String returns the router implementation
	String() string
}

// Table is an interface for routing table
type Table interface {
	// Create new route in the routing table
	Create(Route) error
	// Delete existing route from the routing table
	Delete(Route) error
	// Update route in the routing table
	Update(Route) error
	// List all routes in the table
	List() ([]Route, error)
	// Query routes in the routing table
	Query(...QueryOption) ([]Route, error)
}

type Option func(*Options)

// StatusCode defines router status
type StatusCode int

const (
	// Running means the router is up and running
	Running StatusCode = iota
	// Advertising meas the router is advertising
	Advertising
	// Stopped means the router has been stopped
	Stopped
	// Error means the router has encountered error
	Error
)

// AdvertType is route advertisement type
type AdvertType int

const (
	// Announce is advertised when the router announces itself
	Announce AdvertType = iota
	// RouteUpdate advertise route updates
	RouteUpdate
)

// String returns human readable advertisement type
func (t AdvertType) String() string {
	switch t {
	case Announce:
		return "announce"
	case RouteUpdate:
		return "update"
	default:
		return "unknown"
	}
}

// Advert contains a list of events advertised by the router to the network
type Advert struct {
	// Id is the router id
	Id string
	// Type is type of advert
	Type AdvertType
	// Timestamp marks the time when the update is send
	Timestamp time.Time
	// TTL is advert TTL
	TTL time.Duration
	// Events is a list of routing table events to advertise
	Events []*Event
}

// Strategy is route advertisement strategy
type Strategy int

const (
	// AdvertiseAll advertises all routes to the network
	AdvertiseAll Strategy = iota
	// AdvertiseBest advertises optimal routes to the network
	AdvertiseBest
	// AdvertiseLocal will only advertise the local routes
	AdvertiseLocal
	// AdvertiseNone will not advertise any routes
	AdvertiseNone
)

// String returns human readable Strategy
func (s Strategy) String() string {
	switch s {
	case AdvertiseAll:
		return "all"
	case AdvertiseBest:
		return "best"
	case AdvertiseLocal:
		return "local"
	case AdvertiseNone:
		return "none"
	default:
		return "unknown"
	}
}
