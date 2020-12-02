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

// Package mucp initialises a mucp service
package mucp

import (
	"github.com/lack-io/vine/client"
	cmucp "github.com/lack-io/vine/client/mucp"
	"github.com/lack-io/vine/server"
	smucp "github.com/lack-io/vine/server/mucp"
	"github.com/lack-io/vine/service"
)

type mucpService struct {
	opts service.Options
}

func newService(opts ...service.Option) service.Service {
	options := service.NewOptions(opts...)

	return &mucpService{
		opts: options,
	}
}

func (s *mucpService) Name() string {
	return s.opts.Server.Options().Name
}

// Init initialises options. Additionally it calls cmd.Init
// which parses command line flags. cmd.Init is only called
// on first Init.
func (s *mucpService) Init(opts ...service.Option) {
	// process options
	for _, o := range opts {
		o(&s.opts)
	}
}

func (s *mucpService) Options() service.Options {
	return s.opts
}

func (s *mucpService) Client() client.Client {
	return s.opts.Client
}

func (s *mucpService) Server() server.Server {
	return s.opts.Server
}

func (s *mucpService) String() string {
	return "mucp"
}

func (s *mucpService) Start() error {
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

func (s *mucpService) Stop() error {
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

func (s *mucpService) Run() error {
	if err := s.Start(); err != nil {
		return err
	}

	// wait on context cancel
	<-s.opts.Context.Done()

	return s.Stop()
}

// NewService returns a new mucp service
func NewService(opts ...service.Option) service.Service {
	options := []service.Option{
		service.Client(cmucp.NewClient()),
		service.Server(smucp.NewServer()),
	}

	options = append(options, opts...)

	return newService(options...)
}
