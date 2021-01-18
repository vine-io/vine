// Copyright 2020 lack
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

package service

import (
	"os"
	"os/signal"
	gruntime "runtime"
	"strings"
	"sync"

	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/config/cmd"
	"github.com/lack-io/vine/service/debug/handler"
	"github.com/lack-io/vine/service/debug/stats"
	"github.com/lack-io/vine/service/debug/trace"
	"github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/plugin"
	"github.com/lack-io/vine/service/server"
	"github.com/lack-io/vine/service/store"
	signalutil "github.com/lack-io/vine/util/signal"
	"github.com/lack-io/vine/util/wrapper"
)

type service struct {
	opts Options

	once sync.Once
}

func newService(opts ...Option) Service {
	service := new(service)
	options := newOptions(opts...)

	// service name
	serviceName := options.Server.Options().Name

	// we pass functions to the wrappers since the values can change during initialisation
	authFn := func() auth.Auth { return options.Server.Options().Auth }
	cacheFn := func() *client.Cache { return options.Client.Options().Cache }

	// wrap client to inject From-Service header on any calls
	options.Client = wrapper.FromService(serviceName, options.Client)
	options.Client = wrapper.TraceCall(serviceName, trace.DefaultTracer, options.Client)
	options.Client = wrapper.CacheClient(cacheFn, options.Client)
	options.Client = wrapper.AuthClient(authFn, options.Client)

	// wrap the server to provided handler stats
	options.Server.Init(
		server.WrapHandler(wrapper.HandlerStats(stats.DefaultStats)),
		server.WrapHandler(wrapper.TraceHandler(trace.DefaultTracer)),
		server.WrapHandler(wrapper.AuthHandler(authFn)),
	)

	// set opts
	service.opts = options

	return service
}

func (s *service) Name() string {
	return s.opts.Server.Options().Name
}

// Init initialises options. Additionally it calls cmd.Init
// which parses command line flags. cmd.Init is only called
// on first Init.
func (s *service) Init(opts ...Option) {
	// process options
	for _, o := range opts {
		o(&s.opts)
	}

	s.once.Do(func() {
		// setup the plugins
		for _, p := range strings.Split(os.Getenv("VINE_PLUGIN"), ",") {
			if len(p) == 0 {
				continue
			}

			// load the plugin
			c, err := plugin.Load(p)
			if err != nil {
				logger.Fatal(err)
			}

			// initialise the plugin
			if err := plugin.Init(c); err != nil {
				logger.Fatal(err)
			}
		}

		// set cmd name
		if len(s.opts.Cmd.App().Name) == 0 {
			s.opts.Cmd.App().Name = s.Server().Options().Name
		}

		// Initialise the command flags, overriding new service
		if err := s.opts.Cmd.Init(
			cmd.Auth(&s.opts.Auth),
			cmd.Broker(&s.opts.Broker),
			cmd.Registry(&s.opts.Registry),
			cmd.Runtime(&s.opts.Runtime),
			cmd.Transport(&s.opts.Transport),
			cmd.Client(&s.opts.Client),
			cmd.Config(&s.opts.Config),
			cmd.Server(&s.opts.Server),
			cmd.Store(&s.opts.Store),
			cmd.Profile(&s.opts.Profile),
		); err != nil {
			logger.Fatal(err)
		}

		// Explicitly set the table name to the service name
		name := s.opts.Cmd.App().Name
		s.opts.Store.Init(store.Table(name))
	})
}

func (s *service) Options() Options {
	return s.opts
}

func (s *service) Client() client.Client {
	return s.opts.Client
}

func (s *service) Server() server.Server {
	return s.opts.Server
}

func (s *service) String() string {
	return "vine"
}

func (s *service) Start() error {
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

func (s *service) Stop() error {
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

func (s *service) Run() error {
	// register the debug handler
	s.opts.Server.Handle(
		s.opts.Server.NewHandler(
			handler.NewHandler(s.opts.Client),
			server.InternalHandler(true),
		),
	)

	// start the profiler
	if s.opts.Profile != nil {
		// to view mutex contention
		gruntime.SetMutexProfileFraction(5)
		// to view blocking profile
		gruntime.SetBlockProfileRate(1)

		if err := s.opts.Profile.Start(); err != nil {
			return err
		}
		defer s.opts.Profile.Stop()
	}

	// start the profiler
	logger.Infof("Starting [service] %s", s.Name())

	if err := s.Start(); err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)
	if s.opts.Signal {
		signal.Notify(ch, signalutil.Shutdown()...)
	}

	select {
	// wait on kill signal
	case <-ch:
	// wait on context cancel
	case <-s.opts.Context.Done():
	}

	return s.Stop()
}
