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

// Package tunnel provides gre network tunnelling
package tunnel

import (
	"errors"
	"time"

	"github.com/lack-io/vine/service/transport"
)

const (
	// send over one link
	Unicast Mode = iota
	// send to all channel listeners
	Multicast
	// send to all links
	Broadcast
)

var (
	// DefaultDialTimeout is the dial timeout if none specifies
	DefaultDialTimeout = time.Second * 5
	// ErrDialTimeout is returned by a call to Dial where the timeout occurs
	ErrDialTimeout = errors.New("dial timeout")
	// ErrDiscoverChan is returned when we failed to receive the "announce" back from a discovery
	ErrDiscoverChan = errors.New("failed to discover channel")
	// ErrLinkNotFound is returned when a link is specified at dial time and does not exists
	ErrLinkNotFound = errors.New("link not found")
	// ErrLinkDisconnected is returned when a link we attempt to send to is disconnected
	ErrLinkDisconnected = errors.New("link not connected")
	// ErrLinkLoppback is returned when attempting to send an outbound message over loopback link
	ErrLinkLoopback = errors.New("link is loopback")
	// ErrLinkRemote is returned when attempting to send a loopback message over remote link
	ErrLinkRemote = errors.New("link is remote")
	// ErrReadTimeout is a timeout on session.Recv
	ErrReadTimeout = errors.New("read timeout")
	// ErrDecryptingData is for when theres a nonce error
	ErrDecryptingData = errors.New("error decrypting data")
)

// Mode of the session
type Mode uint8

// Tunnel creates a gre tunnel on top of vine-transport.
// It establishes multiple streams using the Vine-Tunnel-Channel header
// and Vine-Tunnel-Session header. The tunnel id is a hash of
// the address being requested.
type Tunnel interface {
	// Init initializes tunnel with options
	Init(opts ...Option) error
	// Address returns the address the tunnel is listening on
	Address() string
	// Connect connects the tunnel
	Connect() error
	// Close closes the tunnel
	Close() error
	// Links returns all the links the tunnel is connected to
	Links() []Link
	// Dial allows a client to connect to a channel
	Dial(channel string, opts ...DialOption) (Session, error)
	// Listen allows to accept connections on a channel
	Listen(channel string, opts ...ListenOption) (Listener, error)
	// String returns the name of the tunnel implementation
	String() string
}

// Link represents internal links to the tunnel
type Link interface {
	// Id returns the link unique Id
	Id() string
	// Delay is the current load on the link (lower is better)
	Delay() int64
	// Length returns the roundtrip time as nanoseconds (lower is better)
	Length() int64
	// Current transfer rate as bits per second (lower is better)
	Rate() float64
	// Is this a loopback link
	Loopback() bool
	// State of the link: connected/closed/error
	State() string
	// honours transport socket
	transport.Socket
}

// The listener provides similar constructs to the transport.Listener
type Listener interface {
	Accept() (Session, error)
	Channel() string
	Close() error
}

// Session is a unique session created dialling or accepting connections on the tunnel
type Session interface {
	// The unique session id
	Id() string
	// The channel name
	Channel() string
	// The link the session is on
	Link() string
	// a transport socket
	transport.Socket
}

// NewTunnel creates a new tunnel
func NewTunnel(opts ...Option) Tunnel {
	return newTunnel(opts...)
}
