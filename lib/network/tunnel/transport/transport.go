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

// Package transport provides a tunnel transport
package transport

import (
	"context"

	"github.com/lack-io/vine/core/transport"
	"github.com/lack-io/vine/lib/network/tunnel"
)

type tunTransport struct {
	options transport.Options

	tunnel tunnel.Tunnel
}

type tunnelKey struct{}

type transportKey struct{}

func (t *tunTransport) Init(opts ...transport.Option) error {
	for _, o := range opts {
		o(&t.options)
	}

	// close the existing tunnel
	if t.tunnel != nil {
		t.tunnel.Close()
	}

	// get the tunnel
	tun, ok := t.options.Context.Value(tunnelKey{}).(tunnel.Tunnel)
	if !ok {
		tun = tunnel.NewTunnel()
	}

	// get the transport
	tr, ok := t.options.Context.Value(transportKey{}).(transport.Transport)
	if ok {
		tun.Init(tunnel.Transport(tr))
	}

	// set the tunnel
	t.tunnel = tun

	return nil
}

func (t *tunTransport) Dial(addr string, opts ...transport.DialOption) (transport.Client, error) {
	if err := t.tunnel.Connect(); err != nil {
		return nil, err
	}

	c, err := t.tunnel.Dial(addr)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (t *tunTransport) Listen(addr string, opts ...transport.ListenOption) (transport.Listener, error) {
	if err := t.tunnel.Connect(); err != nil {
		return nil, err
	}

	l, err := t.tunnel.Listen(addr)
	if err != nil {
		return nil, err
	}

	return &tunListener{l}, nil
}

func (t *tunTransport) Options() transport.Options {
	return t.options
}

func (t *tunTransport) String() string {
	return "tunnel"
}

// NewTransport honours the initialiser used in
func NewTransport(opts ...transport.Option) transport.Transport {
	t := &tunTransport{
		options: transport.Options{},
	}

	// initialise
	t.Init(opts...)

	return t
}

// WithTransport sets the internal tunnel
func WithTunnel(t tunnel.Tunnel) transport.Option {
	return func(o *transport.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, tunnelKey{}, t)
	}
}

// WithTransport sets the internal transport
func WithTransport(t transport.Transport) transport.Option {
	return func(o *transport.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, transportKey{}, t)
	}
}
