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

import (
	"errors"
	"time"
)

var (
	// ErrWatcherStopped is returned when routing table watcher has been stopped
	ErrWatcherStopped = errors.New("watcher stopped")
)

// EventType defines routing table event
type EventType int

const (
	// Create is emitted when a new route has been created
	Create EventType = iota
	// Delete is emitted when an existing route has been deleted
	Delete
	// Update is emitted when an existing route has been updated
	Update
)

// String returns human readable event type
func (t EventType) String() string {
	switch t {
	case Create:
		return "create"
	case Delete:
		return "delete"
	case Update:
		return "update"
	default:
		return "unknown"
	}
}

// Event is returned by a call to Next on the watcher.
type Event struct {
	// Unique id of the event
	Id string
	// Type defines type of event
	Type EventType
	// Timestamp is event timestamp
	Timestamp time.Time
	// Route is table route
	Route Route
}

// Watcher defines routing watcher interface
// Watcher returns updates to the routing table
type Watcher interface {
	// Next is a blocking call that returns watch result
	Next() (*Event, error)
	// Chan returns event channel
	Chan() (<-chan *Event, error)
	// Stop stops watcher
	Stop()
}

// WatchOption is used to define what routes to watch in the table
type WatchOption func(*WatchOptions)

// WatchOptions are table watcher options
type WatchOptions struct {
	// Service allows to watch specific service routes
	Service string
}

// WatchService sets watch service routes to watch
// Service is the vine service name
func WatchService(s string) WatchOption {
	return func(o *WatchOptions) {
		o.Service = s
	}
}
