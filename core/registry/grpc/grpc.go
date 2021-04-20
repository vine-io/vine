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

// Package grpc service uses the registry service
package grpc

import (
	"context"
	"time"

	"github.com/lack-io/vine/core/client"
	"github.com/lack-io/vine/core/client/grpc"
	"github.com/lack-io/vine/core/registry"
	"github.com/lack-io/vine/proto/apis/errors"
	regpb "github.com/lack-io/vine/proto/apis/registry"
	regSvc "github.com/lack-io/vine/proto/services/registry"
)

var (
	// DefaultService the default service name
	DefaultService = "go.vine.registry"
)

type gRPCRegistry struct {
	opts registry.Options
	// name of the registry
	name string
	// address
	address []string
	// client to call registry
	client regSvc.RegistryService
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

	s.client = regSvc.NewRegistryService(DefaultService, cli)

	return nil
}

func (s *gRPCRegistry) Options() registry.Options {
	return s.opts
}

func (s *gRPCRegistry) Register(svc *regpb.Service, opts ...registry.RegisterOption) error {
	var options registry.RegisterOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	// encode svc into protobuf adn pack Register TTL into it
	svc.Options.Ttl = int64(options.TTL.Seconds())

	// register the service
	_, err := s.client.Register(options.Context, svc, s.callOpts()...)
	return err
}

func (s *gRPCRegistry) Deregister(svc *regpb.Service, opts ...registry.DeregisterOption) error {
	var options registry.DeregisterOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	// deregister the service
	_, err := s.client.Deregister(options.Context, svc, s.callOpts()...)
	return err
}

func (s *gRPCRegistry) GetService(name string, opts ...registry.GetOption) ([]*regpb.Service, error) {
	var options registry.GetOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	rsp, err := s.client.GetService(options.Context, &regSvc.GetRequest{Service: name}, s.callOpts()...)

	if verr, ok := err.(*errors.Error); ok && verr.Code == 404 {
		return nil, registry.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return rsp.Services, nil
}

func (s *gRPCRegistry) ListServices(opts ...registry.ListOption) ([]*regpb.Service, error) {
	var options registry.ListOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	rsp, err := s.client.ListServices(options.Context, &regSvc.ListRequest{}, s.callOpts()...)
	if err != nil {
		return nil, err
	}

	return rsp.Services, nil
}

func (s *gRPCRegistry) Watch(opts ...registry.WatchOption) (registry.Watcher, error) {
	var options registry.WatchOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	stream, err := s.client.Watch(options.Context, &regSvc.WatchRequest{Service: options.Service}, s.callOpts()...)
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
		client:  regSvc.NewRegistryService(name, cli),
	}
}
