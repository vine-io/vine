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

package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/vine-io/vine/core/broker"
	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/core/registry"
	"github.com/vine-io/vine/core/server"
	"github.com/vine-io/vine/lib/cache"
	"github.com/vine-io/vine/lib/config"
	"github.com/vine-io/vine/lib/trace"
)

type Options struct {
	// For the Command Line itself
	Name        string
	Description string
	Version     string

	root bool // app is root?
	app  *cobra.Command

	// We need pointers to things, so we can swap them out if needed.
	Broker   *broker.Broker
	Registry *registry.Registry
	Selector *selector.Selector
	Config   *config.Config
	Client   *client.Client
	Server   *server.Server
	Cache    *cache.Cache
	Tracer   *trace.Tracer

	Brokers    map[string]func(...broker.Option) broker.Broker
	Configs    map[string]func(...config.Option) config.Config
	Clients    map[string]func(...client.Option) client.Client
	Registries map[string]func(...registry.Option) registry.Registry
	Selectors  map[string]func(...selector.Option) selector.Selector
	Servers    map[string]func(...server.Option) server.Server
	Caches     map[string]func(...cache.Option) cache.Cache
	Tracers    map[string]func(...trace.Option) trace.Tracer

	// Other options for implementations of the interfaces
	// can be stored in a context
	Context context.Context
}

type Option func(o *Options)

// NewApp sets the root of command line
func NewApp(app *cobra.Command) Option {
	return func(o *Options) {
		o.root = false
		o.app = app
	}
}

// Name command line Name
func Name(n string) Option {
	return func(o *Options) {
		o.Name = n
	}
}

// Description command line Description
func Description(d string) Option {
	return func(o *Options) {
		o.Description = d
	}
}

// Version command line Version
func Version(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}

func Command(app *cobra.Command) Option {
	return func(o *Options) {
		o.app = app
	}
}

func Broker(b *broker.Broker) Option {
	return func(o *Options) {
		o.Broker = b
	}
}

func Config(c *config.Config) Option {
	return func(o *Options) {
		o.Config = c
	}
}

func Selector(s *selector.Selector) Option {
	return func(o *Options) {
		o.Selector = s
	}
}

func Registry(r *registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

func Client(c *client.Client) Option {
	return func(o *Options) {
		o.Client = c
	}
}

func Server(s *server.Server) Option {
	return func(o *Options) {
		o.Server = s
	}
}

func Cache(c *cache.Cache) Option {
	return func(o *Options) {
		o.Cache = c
	}
}

func Tracer(t *trace.Tracer) Option {
	return func(o *Options) {
		o.Tracer = t
	}
}

// NewBroker new broker func
func NewBroker(name string, b func(...broker.Option) broker.Broker) Option {
	return func(o *Options) {
		o.Brokers[name] = b
	}
}

// NewClient new client func
func NewClient(name string, b func(...client.Option) client.Client) Option {
	return func(o *Options) {
		o.Clients[name] = b
	}
}

// NewRegistry new registry func
func NewRegistry(name string, r func(...registry.Option) registry.Registry) Option {
	return func(o *Options) {
		o.Registries[name] = r
	}
}

// NewSelector new selector func
func NewSelector(name string, s func(...selector.Option) selector.Selector) Option {
	return func(o *Options) {
		o.Selectors[name] = s
	}
}

// NewServer new server func
func NewServer(name string, s func(...server.Option) server.Server) Option {
	return func(o *Options) {
		o.Servers[name] = s
	}
}

// NewTracer new tracer func
func NewTracer(name string, t func(...trace.Option) trace.Tracer) Option {
	return func(o *Options) {
		o.Tracers[name] = t
	}
}
