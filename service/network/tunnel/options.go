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

package tunnel

import (
	"time"

	"github.com/google/uuid"

	"github.com/lack-io/vine/service/transport"
	"github.com/lack-io/vine/service/transport/quic"
)

var (
	// DefaultAddress is default tunnel bind address
	DefaultAddress = ":0"
	// The shared default token
	DefaultToken = "go.vine.tunnel"
)

type Option func(*Options)

// Options provides network configuration options
type Options struct {
	// Id is tunnel id
	Id string
	// Address is tunnel address
	Address string
	// Nodes are remote nodes
	Nodes []string
	// The shared auth token
	Token string
	// Transport listens to incoming connections
	Transport transport.Transport
}

type DialOption func(*DialOptions)

type DialOptions struct {
	// Link specific the link of use
	Link string
	// specify mode of the session
	Mode Mode
	// Wait for connection to be accepted
	Wait bool
	// the dial timeout
	Timeout time.Duration
}

type ListenOption func(*ListenOptions)

type ListenOptions struct {
	// specify mode of the session
	Mode Mode
	// The read timeout
	Timeout time.Duration
}

// The Tunnel id
func Id(id string) Option {
	return func(o *Options) {
		o.Id = id
	}
}

// The tunnel address
func Address(a string) Option {
	return func(o *Options) {
		o.Address = a
	}
}

// Nodes specify remote network nodes
func Nodes(n ...string) Option {
	return func(o *Options) {
		o.Nodes = n
	}
}

// Token sets the shared token for auth
func Token(t string) Option {
	return func(o *Options) {
		o.Token = t
	}
}

// Transport listens for incoming connections
func Transport(t transport.Transport) Option {
	return func(o *Options) {
		o.Transport = t
	}
}

// Listen options
func ListenMode(m Mode) ListenOption {
	return func(o *ListenOptions) {
		o.Mode = m
	}
}

// Timeout for reads and writes on the listener session
func ListenTimeout(t time.Duration) ListenOption {
	return func(o *ListenOptions) {
		o.Timeout = t
	}
}

// Dial options

// Dial multicast sets the multicast option to send only to those mapped
func DialMode(m Mode) DialOption {
	return func(o *DialOptions) {
		o.Mode = m
	}
}

// DialTimeout sets the dial timeout of the connection
func DialTimeout(t time.Duration) DialOption {
	return func(o *DialOptions) {
		o.Timeout = t
	}
}

// DialLink specifies the link to pin this connection to.
// This is not applicable if the multicast option is set.
func DialLink(id string) DialOption {
	return func(o *DialOptions) {
		o.Link = id
	}
}

// DialWait specifies whether to wait for the connection
// to be accepted before returning the session
func DialWait(b bool) DialOption {
	return func(o *DialOptions) {
		o.Wait = b
	}
}

// DefaultOptions returns router default options
func DefaultOptions() Options {
	return Options{
		Id:        uuid.New().String(),
		Address:   DefaultAddress,
		Token:     DefaultToken,
		Transport: quic.NewTransport(),
	}
}
