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
	"context"
	"time"

	"github.com/vine-io/cli"
	"github.com/vine-io/gscheduler"

	"github.com/vine-io/vine/core/broker"
	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/core/registry"
	"github.com/vine-io/vine/core/server"
	"github.com/vine-io/vine/lib/cache"
	"github.com/vine-io/vine/lib/cmd"
	"github.com/vine-io/vine/lib/config"
	"github.com/vine-io/vine/lib/dao"
	"github.com/vine-io/vine/lib/scheduler"
	"github.com/vine-io/vine/lib/trace"
)

// Options for vine service
type Options struct {
	Broker    broker.Broker
	Cmd       cmd.Cmd
	Client    client.Client
	Config    config.Config
	Server    server.Server
	Trace     trace.Tracer
	Dialect   dao.Dialect
	Cache     cache.Cache
	Registry  registry.Registry
	Scheduler gscheduler.Scheduler

	// Before and After functions
	BeforeStart []func() error
	BeforeStop  []func() error
	AfterStart  []func() error
	AfterStop   []func() error

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context

	Signal bool
}

func newOptions(opts ...Option) Options {
	opt := Options{
		Broker:    broker.DefaultBroker,
		Cmd:       cmd.DefaultCmd,
		Client:    client.DefaultClient,
		Config:    config.DefaultConfig,
		Server:    server.DefaultServer,
		Trace:     trace.DefaultTracer,
		Dialect:   dao.DefaultDialect,
		Cache:     cache.DefaultCache,
		Registry:  registry.DefaultRegistry,
		Scheduler: scheduler.DefaultScheduler,
		Context:   context.Background(),
		Signal:    true,
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// Broker to be used for service
func Broker(b broker.Broker) Option {
	return func(o *Options) {
		o.Broker = b
		// Update Client and Service
		_ = o.Client.Init(client.Broker(b))
		_ = o.Server.Init(server.Broker(b))
	}
}

// Cmd to be use for service
func Cmd(c cmd.Cmd) Option {
	return func(o *Options) {
		o.Cmd = c
	}
}

// Client to be used for service
func Client(c client.Client) Option {
	return func(o *Options) {
		o.Client = c
	}
}

// PutContext specifies a context for the service.
// Can be used to signal shutdown of the service and for extra option values.
func PutContext(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
	}
}

// HandleSignal toggle automatic installation of the signal handler that
// traps TERM, INT, and QUIT. Users of this future to disable the signal
// handler, should control liveness of the service through the context
func HandleSignal(b bool) Option {
	return func(o *Options) {
		o.Signal = b
	}
}

// Server to be used for service
func Server(s server.Server) Option {
	return func(o *Options) {
		o.Server = s
	}
}

// Dialect sets the dialect to use
func Dialect(d dao.Dialect) Option {
	return func(o *Options) {
		o.Dialect = d
	}
}

// Cache sets the cache to use
func Cache(c cache.Cache) Option {
	return func(o *Options) {
		o.Cache = c
	}
}

// Registry sets the registry for the service
// and the underlying components
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
		// Update Client and Server
		_ = o.Client.Init(client.Registry(r))
		_ = o.Server.Init(server.Registry(r))
		// Update Broker
		_ = o.Broker.Init(broker.Registry(r))
	}
}

// Tracer sets the tracer for the service
func Tracer(t trace.Tracer) Option {
	return func(o *Options) {
		o.Trace = t
	}
}

// Config sets the config for the service
func Config(c config.Config) Option {
	return func(o *Options) {
		o.Config = c
	}
}

// Selector sets the selector for the service client
func Selector(s selector.Selector) Option {
	return func(o *Options) {
		_ = o.Client.Init(client.Selector(s))
	}
}

// Address sets the address of the server
func Address(addr string) Option {
	return func(o *Options) {
		_ = o.Server.Init(server.Address(addr))
	}
}

// Name of the service
func Name(n string) Option {
	return func(o *Options) {
		_ = o.Server.Init(server.Name(n))
	}
}

// ID of the service
func ID(id string) Option {
	return func(o *Options) {
		_ = o.Server.Init(server.Id(id))
	}
}

// Version of the service
func Version(v string) Option {
	return func(o *Options) {
		_ = o.Server.Init(server.Version(v))
	}
}

// Metadata associated with the service
func Metadata(md map[string]string) Option {
	return func(o *Options) {
		_ = o.Server.Init(server.Metadata(md))
	}
}

// Flags that can be passed to service
func Flags(flags ...cli.Flag) Option {
	return func(o *Options) {
		o.Cmd.App().Flags = append(o.Cmd.App().Flags, flags...)
	}
}

// Action can be used to parse user provided cli options
func Action(a func(*cli.Context) error) Option {
	return func(o *Options) {
		o.Cmd.App().Action = a
	}
}

// RegisterTTL specifies the TTL to use when registering the service
func RegisterTTL(t time.Duration) Option {
	return func(o *Options) {
		_ = o.Server.Init(server.RegisterTTL(t))
	}
}

// RegisterInterval specifies the interval on which to re-register
func RegisterInterval(t time.Duration) Option {
	return func(o *Options) {
		_ = o.Server.Init(server.RegisterInterval(t))
	}
}

// WrapClient is a convenience method for wrapping a Client with
// some middleware component. A list of wrappers can be provided.
// Wrappers are applied in reverse order so the last is executed first.
func WrapClient(w ...client.Wrapper) Option {
	return func(o *Options) {
		// apply in reverse
		for i := len(w); i > 0; i-- {
			o.Client = w[i-1](o.Client)
		}
	}
}

// WrapCall is a convenience method for wrapping a Client CallFunc
func WrapCall(w ...client.CallWrapper) Option {
	return func(o *Options) {
		_ = o.Client.Init(client.WrapCall(w...))
	}
}

// WrapHandler adds a handler Wrapper to a list of options passed into the server
func WrapHandler(w ...server.HandlerWrapper) Option {
	return func(o *Options) {
		var wrappers []server.Option

		for _, wrap := range w {
			wrappers = append(wrappers, server.WrapHandler(wrap))
		}

		// Init Once
		_ = o.Server.Init(wrappers...)
	}
}

// WrapSubscriber adds subscriber Wrapper to a list of options passed into the server
func WrapSubscriber(w ...server.SubscriberWrapper) Option {
	return func(o *Options) {
		var wrappers []server.Option

		for _, wrap := range w {
			wrappers = append(wrappers, server.WrapSubscriber(wrap))
		}

		// Init once
		_ = o.Server.Init(wrappers...)
	}
}

// Before and Afters

// BeforeStart run functions before service starts
func BeforeStart(fn func() error) Option {
	return func(o *Options) {
		o.BeforeStart = append(o.BeforeStart, fn)
	}
}

// BeforeStop run functions before service stops
func BeforeStop(fn func() error) Option {
	return func(o *Options) {
		o.BeforeStop = append(o.BeforeStop, fn)
	}
}

// AfterStart run functions after service starts
func AfterStart(fn func() error) Option {
	return func(o *Options) {
		o.AfterStart = append(o.AfterStart, fn)
	}
}

// AfterStop run functions after service stops
func AfterStop(fn func() error) Option {
	return func(o *Options) {
		o.AfterStop = append(o.AfterStop, fn)
	}
}
