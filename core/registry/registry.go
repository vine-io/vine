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
	"context"
	"errors"
)

var (
	DefaultRegistry Registry

	// ErrNotFound not found error when GetService is called
	ErrNotFound = errors.New("service not found")
	// ErrWatcherStopped watcher stopped error when watcher is stopped
	ErrWatcherStopped = errors.New("watcher stopped")
)

// Registry the registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, ...}
type Registry interface {
	Init(...Option) error
	Options() Options
	Register(context.Context, *Service, ...RegisterOption) error
	Deregister(context.Context, *Service, ...DeregisterOption) error
	GetService(context.Context, string, ...GetOption) ([]*Service, error)
	ListServices(context.Context, ...ListOption) ([]*Service, error)
	Watch(context.Context, ...WatchOption) (Watcher, error)
	String() string
}

// Register a service node. Additionally supply options such as TTL.
func Register(ctx context.Context, s *Service, opts ...RegisterOption) error {
	return DefaultRegistry.Register(ctx, s, opts...)
}

// Deregister a service node
func Deregister(ctx context.Context, s *Service) error {
	return DefaultRegistry.Deregister(ctx, s)
}

// GetService retrieve a service. A slice is returned since we separate Name/Version.
func GetService(ctx context.Context, name string) ([]*Service, error) {
	return DefaultRegistry.GetService(ctx, name)
}

// ListServices list the services. Only returns service names
func ListServices(ctx context.Context) ([]*Service, error) {
	return DefaultRegistry.ListServices(ctx)
}

// Watch returns a watcher which allows you to track updates to the registry.
func Watch(ctx context.Context, opts ...WatchOption) (Watcher, error) {
	return DefaultRegistry.Watch(ctx, opts...)
}

func String() string {
	return DefaultRegistry.String()
}
