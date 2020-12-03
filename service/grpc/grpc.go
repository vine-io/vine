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
	"github.com/lack-io/vine/client"
	gclient "github.com/lack-io/vine/client/grpc"
	"github.com/lack-io/vine/server"
	gserver "github.com/lack-io/vine/server/grpc"
	"github.com/lack-io/vine/service"
)

type grpcService struct {
	opts service.Options
}

func newService(opts ...service.Option) service.Service {
	options := service.NewOptions(opts...)

	return &grpcService{
		opts: options,
	}
}

func (s *grpcService) Name() string {
	return s.opts.Server.Options().Name
}

// Init initialises options. Additionally it calls cmd.Init
// which parses command line flags. cmd.Init is only called
// on first Init.
func (s *grpcService) Init(opts ...service.Option) {
	// process options
	for _, o := range opts {
		o(&s.opts)
	}
}

func (s *grpcService) Options() service.Options {
	return s.opts
}

func (s *grpcService) Client() client.Client {
	return s.opts.Client
}

func (s *grpcService) Server() server.Server {
	return s.opts.Server
}

func (s *grpcService) String() string {
	return "grpc"
}

func (s *grpcService) Start() error {
	for _, fn := range s.opts.BeforeStart {
		if err := fn(); err != nil {
			return err
		}
	}

	if err := s.opts.Server.Start(); err != nil {
		return err
	}

	for _, fn := range s.opts.AfterStart {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (s *grpcService) Stop() error {
	var gerr error

	for _, fn := range s.opts.BeforeStop {
		if err := fn(); err != nil {
			gerr = err
		}
	}

	if err := s.opts.Server.Stop(); err != nil {
		return err
	}

	for _, fn := range s.opts.AfterStop {
		if err := fn(); err != nil {
			gerr = err
		}
	}

	return gerr
}

func (s *grpcService) Run() error {
	if err := s.Start(); err != nil {
		return err
	}

	// wait on context cancel
	<-s.opts.Context.Done()

	return s.Stop()
}

// NewService returns a grpc service compatible with vine.Service
func NewService(opts ...service.Option) service.Service {
	// our grpc client
	c := gclient.NewClient()
	// our grpc server
	s := gserver.NewServer()

	// create options with priority for our opts
	options := []service.Option{
		service.Client(c),
		service.Server(s),
	}

	// append passed in opts
	options = append(options, opts...)

	// generate and return a service
	return newService(options...)
}
