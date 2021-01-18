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
