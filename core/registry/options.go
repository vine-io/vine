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

package registry

import (
	"context"
	"crypto/tls"
	"time"
)

type Options struct {
	Addrs     []string
	Namespace string
	Timeout   time.Duration
	Secure    bool
	TLSConfig *tls.Config
	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

func NewOptions(opts ...Option) Options {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.Context == nil {
		options.Context = context.Background()
	}

	if options.Timeout == 0 {
		options.Timeout = DefaultRegistryTimeout
	}

	if options.Namespace == "" {
		options.Namespace = DefaultNamespace
	}

	return options
}

type RegisterOptions struct {
	TTL       time.Duration
	Namespace string
}

type WatchOptions struct {
	// Specify a service to watch
	// If blank, the watch is for all services
	Service   string
	Namespace string
}

type DeregisterOptions struct {
	Namespace string
}

type GetOptions struct {
	Namespace string
}

type ListOptions struct {
	Namespace string
}

type OpenAPIOptions struct {
}

type Option func(*Options)

type RegisterOption func(*RegisterOptions)

type WatchOption func(*WatchOptions)

type DeregisterOption func(*DeregisterOptions)

type GetOption func(*GetOptions)

type ListOption func(*ListOptions)

type OpenAPIOption func(*OpenAPIOptions)

// Addrs is the registry addresses to use
func Addrs(addrs ...string) Option {
	return func(o *Options) {
		o.Addrs = addrs
	}
}

func Namespace(ns string) Option {
	return func(o *Options) {
		o.Namespace = ns
	}
}

func Timeout(t time.Duration) Option {
	return func(o *Options) {
		o.Timeout = t
	}
}

// Secure communication with the registry
func Secure(b bool) Option {
	return func(o *Options) {
		o.Secure = b
	}
}

// TLSConfig Specify TLS Config
func TLSConfig(t *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}

func RegisterTTL(t time.Duration) RegisterOption {
	return func(o *RegisterOptions) {
		o.TTL = t
	}
}

func RegisterNamespace(ns string) RegisterOption {
	return func(o *RegisterOptions) {
		o.Namespace = ns
	}
}

func DeregisterNamespace(ns string) DeregisterOption {
	return func(o *DeregisterOptions) {
		o.Namespace = ns
	}
}

// WatchService watches a service
func WatchService(name string) WatchOption {
	return func(o *WatchOptions) {
		o.Service = name
	}
}

func WatchNamespace(ns string) WatchOption {
	return func(o *WatchOptions) {
		o.Namespace = ns
	}
}

func GetNamespace(ns string) GetOption {
	return func(o *GetOptions) {
		o.Namespace = ns
	}
}

func ListNamespace(ns string) ListOption {
	return func(o *ListOptions) {
		o.Namespace = ns
	}
}
