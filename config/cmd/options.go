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

package cmd

import (
	"context"

	"github.com/lack-io/vine/auth"
	"github.com/lack-io/vine/broker"
	"github.com/lack-io/vine/client"
	"github.com/lack-io/vine/client/selector"
	"github.com/lack-io/vine/config"
	"github.com/lack-io/vine/debug/profile"
	"github.com/lack-io/vine/debug/trace"
	"github.com/lack-io/vine/registry"
	"github.com/lack-io/vine/runtime"
	"github.com/lack-io/vine/server"
	"github.com/lack-io/vine/store"
	"github.com/lack-io/vine/transport"
)

type Options struct {
	// For the Command Line itself
	Name        string
	Description string
	Version     string

	// We need pointers to things so we can swap them out if needed.
	Broker    *broker.Broker
	Registry  *registry.Registry
	Selector  *selector.Selector
	Transport *transport.Transport
	Config    *config.Config
	Client    *client.Client
	Server    *server.Server
	Runtime   *runtime.Runtime
	Store     *store.Store
	Tracer    *trace.Tracer
	Auth      *auth.Auth
	Profile   *profile.Profile

	Brokers    map[string]func(...broker.Option) broker.Broker
	Configs    map[string]func(...config.Option) (config.Config, error)
	Clients    map[string]func(...client.Option) client.Client
	Registries map[string]func(...registry.Option) registry.Registry
	Selectors  map[string]func(...selector.Option) selector.Selector
	Servers    map[string]func(...server.Option) server.Server
	Transports map[string]func(...transport.Option) transport.Transport
	Runtimes   map[string]func(...runtime.Option) runtime.Runtime
	Stores     map[string]func(...store.Option) store.Store
	Tracers    map[string]func(...trace.Option) trace.Tracer
	Auths      map[string]func(...auth.Option) auth.Auth
	Profiles   map[string]func(...profile.Option) profile.Profile

	// Other options for implementations of the interfaces
	// can be stored in a context
	Context context.Context
}

type Option func(o *Options)

// Command line Name
func Name(n string) Option {
	return func(o *Options) {
		o.Name = n
	}
}

// Command line Description
func Description(d string) Option {
	return func(o *Options) {
		o.Description = d
	}
}

// Command line Version
func Version(v string) Option {
	return func(o *Options) {
		o.Version = v
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

func Runtime(r *runtime.Runtime) Option {
	return func(o *Options) {
		o.Runtime = r
	}
}

func Transport(t *transport.Transport) Option {
	return func(o *Options) {
		o.Transport = t
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

func Store(s *store.Store) Option {
	return func(o *Options) {
		o.Store = s
	}
}

func Tracer(t *trace.Tracer) Option {
	return func(o *Options) {
		o.Tracer = t
	}
}

func Auth(a *auth.Auth) Option {
	return func(o *Options) {
		o.Auth = a
	}
}

func Profile(p *profile.Profile) Option {
	return func(o *Options) {
		o.Profile = p
	}
}

// New broker func
func NewBroker(name string, b func(...broker.Option) broker.Broker) Option {
	return func(o *Options) {
		o.Brokers[name] = b
	}
}

// New client func
func NewClient(name string, b func(...client.Option) client.Client) Option {
	return func(o *Options) {
		o.Clients[name] = b
	}
}

// New registry func
func NewRegistry(name string, r func(...registry.Option) registry.Registry) Option {
	return func(o *Options) {
		o.Registries[name] = r
	}
}

// New selector func
func NewSelector(name string, s func(...selector.Option) selector.Selector) Option {
	return func(o *Options) {
		o.Selectors[name] = s
	}
}

// New server func
func NewServer(name string, s func(...server.Option) server.Server) Option {
	return func(o *Options) {
		o.Servers[name] = s
	}
}

// New transport func
func NewTransport(name string, t func(...transport.Option) transport.Transport) Option {
	return func(o *Options) {
		o.Transports[name] = t
	}
}

// New runtime func
func NewRuntime(name string, r func(...runtime.Option) runtime.Runtime) Option {
	return func(o *Options) {
		o.Runtimes[name] = r
	}
}

// New tracer func
func NewTracer(name string, t func(...trace.Option) trace.Tracer) Option {
	return func(o *Options) {
		o.Tracers[name] = t
	}
}

// New auth func
func NewAuth(name string, t func(...auth.Option) auth.Auth) Option {
	return func(o *Options) {
		o.Auths[name] = t
	}
}
