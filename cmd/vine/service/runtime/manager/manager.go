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
func (m *manager) Create(srv *runtime.Service, opts ...runtime.CreateOption) error {
	// parse the options
	var options runtime.CreateOptions
	for _, o := range opts {
		o(&options)
	}
	if len(options.Namespace) == 0 {
		options.Namespace = namespace.DefaultNamespace
	}

	// set defaults
	if srv.Metadata == nil {
		srv.Metadata = make(map[string]string)
	}
	if len(srv.Version) == 0 {
		srv.Version = "latest"
	}

	// write the object to the store
	if err := m.createService(srv, &options); err != nil {
		return err
	}

	// publish the event, this will apply it aysnc to the runtime
	return m.publishEvent(runtime.Create, srv, &options)
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

	// query the store. TODO: query by type? (it isn't an attr of srv)
	srvs, err := m.readServices(options.Namespace, &runtime.Service{
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
	for _, srv := range srvs {
		md, ok := statuses[srv.Name+":"+srv.Version]
		if !ok {
			continue
		}
		srv.Metadata["status"] = md.Status
		srv.Metadata["error"] = md.Error
	}

	return srvs, nil
}

// Update the service in place
func (m *manager) Update(srv *runtime.Service, opts ...runtime.UpdateOption) error {
	// parse the options
	var options runtime.UpdateOptions
	for _, o := range opts {
		o(&options)
	}
	if len(options.Namespace) == 0 {
		options.Namespace = namespace.DefaultNamespace
	}

	// set defaults
	if len(srv.Version) == 0 {
		srv.Version = "latest"
	}

	// publish the update event which will trigger an update in the runtime
	return m.publishEvent(runtime.Update, srv, &runtime.CreateOptions{Namespace: options.Namespace})
}

// Remove a service
func (m *manager) Delete(srv *runtime.Service, opts ...runtime.DeleteOption) error {
	// parse the options
	var options runtime.DeleteOptions
	for _, o := range opts {
		o(&options)
	}
	if len(options.Namespace) == 0 {
		options.Namespace = namespace.DefaultNamespace
	}

	// set defaults
	if len(srv.Version) == 0 {
		srv.Version = "latest"
	}

	// delete from the store
	if err := m.deleteService(options.Namespace, srv); err != nil {
		return err
	}

	// publish the event which will trigger a delete in the runtime
	return m.publishEvent(runtime.Delete, srv, &runtime.CreateOptions{Namespace: options.Namespace})
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
