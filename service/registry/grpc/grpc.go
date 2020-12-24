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

// Package service uses the registry service
package grpc

import (
	"context"
	"time"

	"github.com/lack-io/vine/proto/errors"
	pb "github.com/lack-io/vine/proto/registry"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/client/grpc"
	"github.com/lack-io/vine/service/registry"
)

var (
	// The default service name
	DefaultService = "go.vine.registry"
)

type gRPCRegistry struct {
	opts registry.Options
	// name of the registry
	name string
	// address
	address []string
	// client to call registry
	client pb.RegistryService
}

func (s *gRPCRegistry) callOpts() []client.CallOption {
	var opts []client.CallOption

	// set registry address
	if len(s.address) > 0 {
		opts = append(opts, client.WithAddress(s.address...))
	}

	// set timeout
	if s.opts.Timeout > time.Duration(0) {
		opts = append(opts, client.WithRequestTimeout(s.opts.Timeout))
	}

	return opts
}

func (s *gRPCRegistry) Init(opts ...registry.Option) error {
	for _, o := range opts {
		o(&s.opts)
	}

	if len(s.opts.Addrs) > 0 {
		s.address = s.opts.Addrs
	}

	// extract the client from the context, fallback to grpc
	var cli client.Client
	if c, ok := s.opts.Context.Value(clientKey{}).(client.Client); ok {
		cli = c
	} else {
		cli = grpc.NewClient()
	}

	s.client = pb.NewRegistryService(DefaultService, cli)

	return nil
}

func (s *gRPCRegistry) Options() registry.Options {
	return s.opts
}

func (s *gRPCRegistry) Register(srv *registry.Service, opts ...registry.RegisterOption) error {
	var options registry.RegisterOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	// encode srv into protobuf adn pack Register TTL into it
	pbSrv := ToProto(srv)
	pbSrv.Options.Ttl = int64(options.TTL.Seconds())

	// register the service
	_, err := s.client.Register(options.Context, pbSrv, s.callOpts()...)
	return err
}

func (s *gRPCRegistry) Deregister(srv *registry.Service, opts ...registry.DeregisterOption) error {
	var options registry.DeregisterOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	// deregister the service
	_, err := s.client.Deregister(options.Context, ToProto(srv), s.callOpts()...)
	return err
}

func (s *gRPCRegistry) GetService(name string, opts ...registry.GetOption) ([]*registry.Service, error) {
	var options registry.GetOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	rsp, err := s.client.GetService(options.Context, &pb.GetRequest{Service: name}, s.callOpts()...)

	if verr, ok := err.(*errors.Error); ok && verr.Code == 404 {
		return nil, registry.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	services := make([]*registry.Service, 0, len(rsp.Services))
	for _, service := range rsp.Services {
		services = append(services, ToService(service))
	}

	return services, nil
}

func (s *gRPCRegistry) ListServices(opts ...registry.ListOption) ([]*registry.Service, error) {
	var options registry.ListOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	rsp, err := s.client.ListServices(options.Context, &pb.ListRequest{}, s.callOpts()...)
	if err != nil {
		return nil, err
	}

	services := make([]*registry.Service, 0, len(rsp.Services))
	for _, service := range rsp.Services {
		services = append(services, ToService(service))
	}

	return services, nil
}

func (s *gRPCRegistry) Watch(opts ...registry.WatchOption) (registry.Watcher, error) {
	var options registry.WatchOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	stream, err := s.client.Watch(options.Context, &pb.WatchRequest{Service: options.Service}, s.callOpts()...)
	if err != nil {
		return nil, err
	}

	return newWatcher(stream), nil
}

func (s *gRPCRegistry) String() string {
	return "service"
}

// NewRegistry returns a new registry service client
func NewRegistry(opts ...registry.Option) registry.Registry {
	var options registry.Options
	for _, o := range opts {
		o(&options)
	}

	// the registry address
	addrs := options.Addrs
	if len(addrs) == 0 {
		addrs = []string{"127.0.0.1:8000"}
	}

	if options.Context == nil {
		options.Context = context.TODO()
	}

	// extract the client from the context, fallback to grpc
	var cli client.Client
	if c, ok := options.Context.Value(clientKey{}).(client.Client); ok {
		cli = c
	} else {
		cli = grpc.NewClient()
	}

	// service name. TODO: accept option
	name := DefaultService

	return &gRPCRegistry{
		opts:    options,
		name:    name,
		address: addrs,
		client:  pb.NewRegistryService(name, cli),
	}
}
