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

package web

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lack-io/cli"
	"github.com/lack-io/vine/core/registry"

	"github.com/lack-io/vine"
)

//Options for web
type Options struct {
	Name      string
	Version   string
	Id        string
	Metadata  map[string]string
	Address   string
	Advertise string

	Action func(*cli.Context)
	Flags  []cli.Flag

	RegisterTTL      time.Duration
	RegisterInterval time.Duration

	// RegisterCheck runs a check function before registering the service
	RegisterCheck func(context.Context) error

	App    *fiber.App

	// Alternative Options
	Context context.Context

	Registry registry.Registry
	Service  vine.Service

	Secure      bool
	TLSConfig   *tls.Config
	BeforeStart []func() error
	BeforeStop  []func() error
	AfterStart  []func() error
	AfterStop   []func() error

	// Static directory
	StaticDir string

	Signal bool
}

func newOptions(opts ...Option) Options {
	opt := Options{
		Name:             DefaultName,
		Version:          DefaultVersion,
		Id:               DefaultId,
		Address:          DefaultAddress,
		RegisterTTL:      DefaultRegisterTTL,
		RegisterInterval: DefaultRegisterInterval,
		StaticDir:        DefaultStaticDir,
		Service:          vine.NewService(),
		Context:          context.TODO(),
		Signal:           true,
	}

	for _, o := range opts {
		o(&opt)
	}

	if opt.RegisterCheck == nil {
		opt.RegisterCheck = DefaultRegisterCheck
	}

	return opt
}

// Name of Web
func Name(n string) Option {
	return func(o *Options) {
		o.Name = n
	}
}

// Icon specifies an icon url to load in the UI
func Icon(ico string) Option {
	return func(o *Options) {
		if o.Metadata == nil {
			o.Metadata = make(map[string]string)
		}
		o.Metadata["icon"] = ico
	}
}

//Id for Unique server id
func Id(id string) Option {
	return func(o *Options) {
		o.Id = id
	}
}

// Version of the service
func Version(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}

// Metadata associated with the service
func Metadata(md map[string]string) Option {
	return func(o *Options) {
		o.Metadata = md
	}
}

// Address to bind to - host:port
func Address(a string) Option {
	return func(o *Options) {
		o.Address = a
	}
}

//Advertise The address to advertise for discovery - host:port
func Advertise(a string) Option {
	return func(o *Options) {
		o.Advertise = a
	}
}

// Context specifies a context for the service.
// Can be used to signal shutdown of the service.
// Can be used for extra option values.
func Context(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
	}
}

// Registry used for discovery
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

//RegisterTTL Register the service with a TTL
func RegisterTTL(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterTTL = t
	}
}

//RegisterInterval Register the service with at interval
func RegisterInterval(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterInterval = t
	}
}

// App for custom fiber.App
func App(app *fiber.App) Option {
	return func(o *Options) {
		o.App = app
	}
}

// VineService sets the vine.Service used internally
func VineService(s vine.Service) Option {
	return func(o *Options) {
		o.Service = s
	}
}

// Flags sets the command flags.
func Flags(flags ...cli.Flag) Option {
	return func(o *Options) {
		o.Flags = append(o.Flags, flags...)
	}
}

// Action sets the command action.
func Action(a func(*cli.Context)) Option {
	return func(o *Options) {
		o.Action = a
	}
}

// BeforeStart is executed before the server starts.
func BeforeStart(fn func() error) Option {
	return func(o *Options) {
		o.BeforeStart = append(o.BeforeStart, fn)
	}
}

// BeforeStop is executed before the server stops.
func BeforeStop(fn func() error) Option {
	return func(o *Options) {
		o.BeforeStop = append(o.BeforeStop, fn)
	}
}

// AfterStart is executed after server start.
func AfterStart(fn func() error) Option {
	return func(o *Options) {
		o.AfterStart = append(o.AfterStart, fn)
	}
}

// AfterStop is executed after server stop.
func AfterStop(fn func() error) Option {
	return func(o *Options) {
		o.AfterStop = append(o.AfterStop, fn)
	}
}

// Secure Use secure communication. If TLSConfig is not specified we use InsecureSkipVerify and generate a self signed cert
func Secure(b bool) Option {
	return func(o *Options) {
		o.Secure = b
	}
}

// TLSConfig to be used for the transport.
func TLSConfig(t *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}

// StaticDir sets the static file directory. This defaults to ./html
func StaticDir(d string) Option {
	return func(o *Options) {
		o.StaticDir = d
	}
}

// RegisterCheck run func before registry service
func RegisterCheck(fn func(context.Context) error) Option {
	return func(o *Options) {
		o.RegisterCheck = fn
	}
}

// HandleSignal toggles automatic installation of the signal handler that
// traps TERM, INT, and QUIT.  Users of this feature to disable the signal
// handler, should control liveness of the service through the context.
func HandleSignal(b bool) Option {
	return func(o *Options) {
		o.Signal = b
	}
}
