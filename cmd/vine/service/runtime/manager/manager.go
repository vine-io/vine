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

package manager

import (
	"github.com/lack-io/vine/service/config/cmd"
	"github.com/lack-io/vine/service/runtime"
	"github.com/lack-io/vine/service/store"
	"github.com/lack-io/vine/service/store/memory"
	"github.com/lack-io/vine/util/namespace"
)

// Init initializes the runtime
func (m *manager) Init(...runtime.Option) error {
	return nil
}

// Create registers a service
func (m *manager) Create(svc *runtime.Service, opts ...runtime.CreateOption) error {
	// parse the options
	var options runtime.CreateOptions
	for _, o := range opts {
		o(&options)
	}
	if len(options.Namespace) == 0 {
		options.Namespace = namespace.DefaultNamespace
	}

	// set defaults
	if svc.Metadata == nil {
		svc.Metadata = make(map[string]string)
	}
	if len(svc.Version) == 0 {
		svc.Version = "latest"
	}

	// write the object to the store
	if err := m.createService(svc, &options); err != nil {
		return err
	}

	// publish the event, this will apply it aysnc to the runtime
	return m.publishEvent(runtime.Create, svc, &options)
}

// Read returns the service which matches the criteria provided
func (m *manager) Read(opts ...runtime.ReadOption) ([]*runtime.Service, error) {
	// parse the options
	var options runtime.ReadOptions
	for _, o := range opts {
		o(&options)
	}
	if len(options.Namespace) == 0 {
		options.Namespace = namespace.DefaultNamespace
	}

	// query the store. TODO: query by type? (it isn't an attr of svc)
	svcs, err := m.readServices(options.Namespace, &runtime.Service{
		Name:    options.Service,
		Version: options.Version,
	})
	if err != nil {
		return nil, err
	}

	// add the metadata to the service from the local runtime (e.g. status, err)
	statuses, err := m.listStatuses(options.Namespace)
	if err != nil {
		return nil, err
	}
	for _, svc := range svcs {
		md, ok := statuses[svc.Name+":"+svc.Version]
		if !ok {
			continue
		}
		svc.Metadata["status"] = md.Status
		svc.Metadata["error"] = md.Error
	}

	return svcs, nil
}

// Update the service in place
func (m *manager) Update(svc *runtime.Service, opts ...runtime.UpdateOption) error {
	// parse the options
	var options runtime.UpdateOptions
	for _, o := range opts {
		o(&options)
	}
	if len(options.Namespace) == 0 {
		options.Namespace = namespace.DefaultNamespace
	}

	// set defaults
	if len(svc.Version) == 0 {
		svc.Version = "latest"
	}

	// publish the update event which will trigger an update in the runtime
	return m.publishEvent(runtime.Update, svc, &runtime.CreateOptions{Namespace: options.Namespace})
}

// Remove a service
func (m *manager) Delete(svc *runtime.Service, opts ...runtime.DeleteOption) error {
	// parse the options
	var options runtime.DeleteOptions
	for _, o := range opts {
		o(&options)
	}
	if len(options.Namespace) == 0 {
		options.Namespace = namespace.DefaultNamespace
	}

	// set defaults
	if len(svc.Version) == 0 {
		svc.Version = "latest"
	}

	// delete from the store
	if err := m.deleteService(options.Namespace, svc); err != nil {
		return err
	}

	// publish the event which will trigger a delete in the runtime
	return m.publishEvent(runtime.Delete, svc, &runtime.CreateOptions{Namespace: options.Namespace})
}

// Starts the manager
func (m *manager) Start() error {
	if m.running {
		return nil
	}
	m.running = true

	// start the runtime we're going to manage
	if err := m.Runtime.Start(); err != nil {
		return err
	}

	// watch events written to the store
	go m.watchEvents()

	// periodically load the status of services from the runtime
	go m.watchStatus()

	// todo: compare the store to the runtime incase we missed any events

	return nil
}

// Stop the manager
func (m *manager) Stop() error {
	if !m.running {
		return nil
	}
	m.running = false

	return m.Runtime.Stop()
}

// String describes runtime
func (m *manager) String() string {
	return "manager"
}

type manager struct {
	// runtime being managed
	runtime.Runtime
	// options passed by the caller
	options Options
	// running is true after Start is called
	running bool
	// cache is a memory store which is used to store any information we don't want to write to the
	// global store, e.g. events consumed, service status / errors (these will change depending on the
	// managed runtime and hence won't be the same globally).
	cache store.Store
}

// New returns a manager for the runtime
func New(r runtime.Runtime, opts ...Option) runtime.Runtime {
	// parse the options
	var options Options
	for _, o := range opts {
		o(&options)
	}

	// set the defaults
	if options.Store == nil {
		options.Store = *cmd.DefaultCmd.Options().Store
	}

	return &manager{
		Runtime: r,
		options: options,
		cache:   memory.NewStore(),
	}
}
