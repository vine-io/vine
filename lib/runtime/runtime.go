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

// Package runtime is a service runtime manager
package runtime

import (
	"errors"
	"time"
)

var (
	// DefaultRuntime is default vine runtime
	DefaultRuntime Runtime = NewRuntime()
	// DefaultName is default runtime service name
	DefaultName = "go.vine.runtime"

	ErrAlreadyExists = errors.New("already exists")
)

// Runtime is a service runtime manager
type Runtime interface {
	// Init initializes runtime
	Init(...Option) error
	// Create registers a service
	Create(*Service, ...CreateOption) error
	// Read returns the service
	Read(...ReadOption) ([]*Service, error)
	// Update the service in place
	Update(*Service, ...UpdateOption) error
	// Remove a service
	Delete(*Service, ...DeleteOption) error
	// Logs returns the logs for a service
	Logs(*Service, ...LogsOption) (LogStream, error)
	// Start starts the runtime
	Start() error
	// Stop shuts down the runtime
	Stop() error
	// String describes runtime
	String() string
}

// Stream returns a log stream
type LogStream interface {
	Error() error
	Chan() chan LogRecord
	Stop() error
}

type LogRecord struct {
	Message  string
	Metadata map[string]string
}

// Scheduler is a runtime service scheduler
type Scheduler interface {
	// Notify publishes schedule events
	Notify() (<-chan Event, error)
	// Close stops the scheduler
	Close() error
}

// EventType defines schedule event
type EventType int

const (
	// Create is emitted when a new build has been craeted
	Create EventType = iota
	// Update is emitted when a new update become available
	Update
	// Delete is emitted when a build has been deleted
	Delete
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

// Event is notification event
type Event struct {
	// ID of the event
	ID string
	// Type is event type
	Type EventType
	// Timestamp is event timestamp
	Timestamp time.Time
	// Service the event relates to
	Service *Service
	// Options to use when processing the event
	Options *CreateOptions
}

// Service is runtime service
type Service struct {
	// Name of the service
	Name string
	// Version of the service
	Version string
	// url location of source
	Source string
	// Metadata stores metadata
	Metadata map[string]string
}
