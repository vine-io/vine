// MIT License
//
// Copyright (c) 2021 Lack
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

// Service represents a vine service
type Service struct {
	Name      string            `json:"name,omitempty"`
	Version   string            `json:"version,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Endpoints []*Endpoint       `json:"endpoints,omitempty"`
	Nodes     []*Node           `json:"nodes,omitempty"`
	TTL       int64             `json:"ttl,omitempty"`
}

// Node represents the node the service is on
type Node struct {
	Id       string            `json:"id,omitempty"`
	Address  string            `json:"address,omitempty"`
	Port     int64             `json:"port,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Endpoint is a endpoint provided by a service
type Endpoint struct {
	Name     string            `json:"name,omitempty"`
	Request  *Value            `json:"request,omitempty"`
	Response *Value            `json:"response,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Value is an opaque value for a request or response
type Value struct {
	Name   string   `json:"name,omitempty"`
	Type   string   `json:"type,omitempty"`
	Values []*Value `json:"values,omitempty"`
}

// Result is returns by the watcher
type Result struct {
	Action    string   `json:"action,omitempty"`
	Service   *Service `json:"service,omitempty"`
	Timestamp int64    `json:"timestamp,omitempty"`
}

type EventType string

const (
	EventCreate EventType = "Create"
	EventUpdate EventType = "Update"
	EventDelete EventType = "Delete"
)

// Event is registry event
type Event struct {
	// Event Id
	Id string `json:"id,omitempty"`
	// type of event
	Type EventType `json:"type,omitempty"`
	// unix timestamp of event
	Timestamp int64 `json:"timestamp,omitempty"`
	// service entry
	Service *Service `json:"service,omitempty"`
}
