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

	regpb "github.com/lack-io/vine/proto/apis/registry"
)

var (
	DefaultRegistry Registry

	// Not found error when GetService is called
	ErrNotFound = errors.New("service not found")
	// Watcher stopped error when watcher is stopped
	ErrWatcherStopped = errors.New("watcher stopped")
)

// The registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, ...}
type Registry interface {
	Init(...Option) error
	Options() Options
	Register(*regpb.Service, ...RegisterOption) error
	Deregister(*regpb.Service, ...DeregisterOption) error
	GetService(string, ...GetOption) ([]*regpb.Service, error)
	ListServices(...ListOption) ([]*regpb.Service, error)
	Watch(...WatchOption) (Watcher, error)
	String() string
}

type Option func(*Options)

type RegisterOption func(*RegisterOptions)

type WatchOption func(*WatchOptions)

type DeregisterOption func(*DeregisterOptions)

type GetOption func(*GetOptions)

type ListOption func(*ListOptions)

type OpenAPIOption func(*OpenAPIOptions)

// Register a service node. Additionally supply options such as TTL.
func Register(s *regpb.Service, opts ...RegisterOption) error {
	return DefaultRegistry.Register(s, opts...)
}

// Deregister a service node
func Deregister(s *regpb.Service) error {
	return DefaultRegistry.Deregister(s)
}

// Retrieve a service. A slice is returned since we separate Name/Version.
func GetService(name string) ([]*regpb.Service, error) {
	return DefaultRegistry.GetService(name)
}

// List the services. Only returns service names
func ListServices() ([]*regpb.Service, error) {
	return DefaultRegistry.ListServices()
}

// Watch returns a watcher which allows you to track updates to the registry.
func Watch(opts ...WatchOption) (Watcher, error) {
	return DefaultRegistry.Watch(opts...)
}

func String() string {
	return DefaultRegistry.String()
}
