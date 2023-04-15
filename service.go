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

package vine

import (
	"os"
	"os/signal"
	"sync"

	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/core/server"
	"github.com/vine-io/vine/lib/cache"
	"github.com/vine-io/vine/lib/cmd"
	"github.com/vine-io/vine/lib/logger"
	"github.com/vine-io/vine/lib/trace"
	usignal "github.com/vine-io/vine/util/signal"
	"github.com/vine-io/vine/util/wrapper"
)

type service struct {
	opts Options

	once sync.Once
}

func newService(opts ...Option) Service {
	sv := new(service)
	options := newOptions(opts...)

	// service name
	serviceName := options.Server.Options().Name

	// wrap client to inject From-Service header on any calls
	options.Client = wrapper.FromService(serviceName, options.Client)
	options.Client = wrapper.TraceCall(serviceName, trace.DefaultTracer, options.Client)

	// wrap the server to provided handler stats
	_ = options.Server.Init(
		server.WrapHandler(wrapper.TraceHandler(trace.DefaultTracer)),
	)

	// set opts
	sv.opts = options

	return sv
}

func (s *service) Name() string {
	return s.opts.Server.Options().Name
}

// Init initialises options. Additionally, it calls cmd.Init
// which parses command line flags. cmd.Init is only called
// on first Init.
func (s *service) Init(opts ...Option) error {
	// process options
	for _, o := range opts {
		o(&s.opts)
	}

	var err error
	s.once.Do(func() {

		if s.opts.Cmd != nil {

			options := []cmd.Option{
				cmd.Broker(&s.opts.Broker),
				cmd.Registry(&s.opts.Registry),
				cmd.Client(&s.opts.Client),
				cmd.Config(&s.opts.Config),
				cmd.Server(&s.opts.Server),
				cmd.Cache(&s.opts.Cache),
			}

			if len(s.opts.Cmd.Options().Name) == 0 {
				options = append(options, cmd.Name(s.opts.Server.Options().Name))
			}
			if len(s.opts.Cmd.Options().Version) == 0 {
				options = append(options, cmd.Version(s.opts.Server.Options().Version))
			}

			// Initialise the command flags, overriding new service
			if err = s.opts.Cmd.Init(options...); err != nil {
				return
			}
		}

		// Explicitly set the table name to the service name
		name := s.opts.Server.Options().Name
		if err = s.opts.Cache.Init(cache.Table(name)); err != nil {
			return
		}
	})

	return err
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

	select {
	case <-s.opts.Context.Done():
	default:
		s.opts.Cancel()
	}

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
	logger.Infof("Starting [service] %s", s.Name())
	logger.Infof("service [version] %s", s.Options().Server.Options().Version)

	if err := s.Start(); err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)
	if s.opts.Signal {
		signal.Notify(ch, usignal.Shutdown()...)
	}

	select {
	// wait on kill signal
	case <-ch:
		s.opts.Cancel()
	// wait on context cancel
	case <-s.opts.Context.Done():
	}

	return s.Stop()
}

func (s *service) String() string {
	return "vine"
}
