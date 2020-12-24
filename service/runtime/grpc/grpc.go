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

package grpc

import (
	"context"
	"sync"

	pb "github.com/lack-io/vine/proto/runtime"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/runtime"
)

type gRPC struct {
	sync.RWMutex
	options runtime.Options
	runtime pb.RuntimeService
}

// Init initializes runtime with given options
func (s *gRPC) Init(opts ...runtime.Option) error {
	s.Lock()
	defer s.Unlock()

	for _, o := range opts {
		o(&s.options)
	}

	// reset the runtime as the client could have changed
	s.runtime = pb.NewRuntimeService("go.vine.runtime", s.options.Client)

	return nil
}

// Create registers a service in the runtime
func (s *gRPC) Create(svc *runtime.Service, opts ...runtime.CreateOption) error {
	var options runtime.CreateOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.Background()
	}

	// set the default source from VINE_RUNTIME_SOURCE
	if len(svc.Source) == 0 {
		svc.Source = s.options.Source
	}

	// runtime service create request
	req := &pb.CreateRequest{
		Service: &pb.Service{
			Name:     svc.Name,
			Version:  svc.Version,
			Source:   svc.Source,
			Metadata: svc.Metadata,
		},
		Options: &pb.CreateOptions{
			Command: options.Command,
			Args:    options.Args,
			Env:     options.Env,
			Type:    options.Type,
			Image:   options.Image,
		},
	}

	if _, err := s.runtime.Create(options.Context, req); err != nil {
		return err
	}

	return nil
}

func (s *gRPC) Logs(service *runtime.Service, opts ...runtime.LogsOption) (runtime.LogStream, error) {
	var options runtime.LogsOptions
	for _, o := range opts {
		o(&options)
	}

	if options.Context == nil {
		options.Context = context.Background()
	}

	ls, err := s.runtime.Logs(options.Context, &pb.LogsRequest{
		Service: service.Name,
		Stream:  options.Stream,
		Count:   options.Count,
	})
	if err != nil {
		return nil, err
	}
	logStream := &serviceLogStream{
		service: service.Name,
		stream:  make(chan runtime.LogRecord),
		stop:    make(chan bool),
	}

	go func() {
		for {
			select {
			// @todo this never seems to return, investigate
			case <-ls.Context().Done():
				logStream.Stop()
			}
		}
	}()

	go func() {
		for {
			select {
			// @todo this never seems to return, investigate
			case <-ls.Context().Done():
				return
			case _, ok := <-logStream.stream:
				if !ok {
					return
				}
			default:
				record := pb.LogRecord{}
				err := ls.RecvMsg(&record)
				if err != nil {
					logStream.Stop()
					return
				}
				logStream.stream <- runtime.LogRecord{
					Message:  record.GetMessage(),
					Metadata: record.GetMetadata(),
				}
			}
		}
	}()
	return logStream, nil
}

type serviceLogStream struct {
	service string
	stream  chan runtime.LogRecord
	sync.Mutex
	stop chan bool
	err  error
}

func (l *serviceLogStream) Error() error {
	return l.err
}

func (l *serviceLogStream) Chan() chan runtime.LogRecord {
	return l.stream
}

func (l *serviceLogStream) Stop() error {
	l.Lock()
	defer l.Unlock()
	select {
	case <-l.stop:
		return nil
	default:
		close(l.stream)
		close(l.stop)
	}
	return nil
}

// Read returns the service with the given name from the runtime
func (s *gRPC) Read(opts ...runtime.ReadOption) ([]*runtime.Service, error) {
	var options runtime.ReadOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.Background()
	}

	// runtime service create request
	req := &pb.ReadRequest{
		Options: &pb.ReadOptions{
			Service: options.Service,
			Version: options.Version,
			Type:    options.Type,
		},
	}

	resp, err := s.runtime.Read(options.Context, req)
	if err != nil {
		return nil, err
	}

	services := make([]*runtime.Service, 0, len(resp.Services))
	for _, service := range resp.Services {
		svc := &runtime.Service{
			Name:     service.Name,
			Version:  service.Version,
			Source:   service.Source,
			Metadata: service.Metadata,
		}
		services = append(services, svc)
	}

	return services, nil
}

// Update updates the running service
func (s *gRPC) Update(svc *runtime.Service, opts ...runtime.UpdateOption) error {
	var options runtime.UpdateOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.Background()
	}

	// runtime service create request
	req := &pb.UpdateRequest{
		Service: &pb.Service{
			Name:     svc.Name,
			Version:  svc.Version,
			Source:   svc.Source,
			Metadata: svc.Metadata,
		},
	}

	if _, err := s.runtime.Update(options.Context, req); err != nil {
		return err
	}

	return nil
}

// Delete stops and removes the service from the runtime
func (s *gRPC) Delete(svc *runtime.Service, opts ...runtime.DeleteOption) error {
	var options runtime.DeleteOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.Background()
	}

	// runtime service create request
	req := &pb.DeleteRequest{
		Service: &pb.Service{
			Name:     svc.Name,
			Version:  svc.Version,
			Source:   svc.Source,
			Metadata: svc.Metadata,
		},
	}

	if _, err := s.runtime.Delete(options.Context, req); err != nil {
		return err
	}

	return nil
}

// Start starts the runtime
func (s *gRPC) Start() error {
	// NOTE: nothing to be done here
	return nil
}

// Stop stops the runtime
func (s *gRPC) Stop() error {
	// NOTE: nothing to be done here
	return nil
}

// Returns the runtime service implementation
func (s *gRPC) String() string {
	return "grpc"
}

// NewRuntime creates new service runtime and returns it
func NewRuntime(opts ...runtime.Option) runtime.Runtime {
	var options runtime.Options

	for _, o := range opts {
		o(&options)
	}
	if options.Client == nil {
		options.Client = client.DefaultClient
	}

	return &gRPC{
		options: options,
		runtime: pb.NewRuntimeService("go.vine.runtime", options.Client),
	}
}
