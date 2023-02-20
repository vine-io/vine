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

package server

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/pflag"
	"github.com/vine-io/vine/core/codec"
	"github.com/vine-io/vine/core/registry"
	log "github.com/vine-io/vine/lib/logger"
	usignal "github.com/vine-io/vine/util/signal"
)

var (
	DefaultAddress          = ":0"
	DefaultName             = "go.vine.server"
	DefaultVersion          = "latest"
	DefaultId               = uuid.New().String()
	DefaultServer           Server
	DefaultRegisterCheck    = func(context.Context) error { return nil }
	DefaultRegisterInterval = time.Second * 20
	DefaultRegisterTTL      = time.Second * 30

	Flag = pflag.NewFlagSet("server", pflag.ExitOnError)
)

func init() {
	Flag.String("server.default", "", "Server for vine")
	Flag.String("server.name", "", "Name of the server")
	Flag.String("server.address", "", "Bind address for the server")
	Flag.String("server.id", "", "Id of the server")
	Flag.String("server.advertise", "", "Use instead of the server-address when registering with discovery")
	Flag.StringSlice("server.metadata", nil, "A list of key-value pairs defining metadata")
	Flag.Duration("server.register-interval", 0, "Register interval")
	Flag.Duration("server.register-ttl", 0, "Registry TTL")
}

// Server is a simple vine server abstraction
type Server interface {
	// Init initialise options
	Init(...Option) error
	// Options retrieve the options
	Options() Options
	// Handle register a handler
	Handle(Handler) error
	// NewHandler create a new handler
	NewHandler(interface{}, ...HandlerOption) Handler
	// NewSubscriber create a new subscriber
	NewSubscriber(string, interface{}, ...SubscriberOption) Subscriber
	// Subscribe register a subscriber
	Subscribe(Subscriber) error
	// Start the server
	Start() error
	// Stop the server
	Stop() error
	// String implementation
	String() string
}

// Router handle serving messages
type Router interface {
	// ProcessMessage processes a message
	ProcessMessage(context.Context, Message) error
	// ServeRequest processes a request to completion
	ServeRequest(context.Context, Request, Response) error
}

// Message is an async message interface
type Message interface {
	// Topic of the message
	Topic() string
	// Payload the decoded payload value
	Payload() interface{}
	// ContentType the content type of the payload
	ContentType() string
	// Header the raw headers of the message
	Header() map[string]string
	// Body the raw body of the message
	Body() []byte
	// Codec used tp decode the message
	Codec() codec.Reader
}

// Request is a synchronous request interface
type Request interface {
	// Service name requested
	Service() string
	// Method the action requested
	Method() string
	// Endpoint name requested
	Endpoint() string
	// ContentType Content Type provided
	ContentType() string
	// Header of the request
	Header() map[string]string
	// Body is the initial decoded value
	Body() interface{}
	// Read the undecoded request body
	Read() ([]byte, error)
	// Codec the encoded message body
	Codec() codec.Reader
	// Stream indicates whether it's a stream
	Stream() bool
}

// Response is the response write for unencoded messages
type Response interface {
	// Codec encoded writer
	Codec() codec.Writer
	// WriteHeader write the header
	WriteHeader(map[string]string)
	// Write a response directly to the client
	Write([]byte) error
}

// Stream represents a stream established with a client.
// A stream can be bidirectional which is indicated by the request.
// The last error will be left in Error().
// EOF indicates end of the stream.
type Stream interface {
	Context() context.Context
	Request() Request
	Send(interface{}) error
	Recv(interface{}) error
	Error() error
	Close() error
}

// Handler interface represents a request handler. It's generated
// by passing any type of public concrete object with endpoints into server.NewHandler.
// Most will pass in a struct.
//
// Example:
//
//	type Greeter struct{}
//
//	func (g *Greeter) Hello(context, request, response) error {
//		return nil
//	}
type Handler interface {
	Name() string
	Handler() interface{}
	Endpoints() []*registry.Endpoint
	Options() HandlerOptions
}

// Subscriber interface represents a subscription to a given topic using
// a specific subscriber function or object with endpoints. It mirrors
// the handler in its behaviour.
type Subscriber interface {
	Topic() string
	Subscriber() interface{}
	Endpoints() []*registry.Endpoint
	Options() SubscriberOptions
}

// DefaultOptions returns config options for the default service
func DefaultOptions() Options {
	return DefaultServer.Options()
}

// Init initialises the default server with options passed in
func Init(opts ...Option) {
	_ = DefaultServer.Init(opts...)
}

// NewSubscriber creates a new subscriber interface with the given topic
// and handler using the default server
func NewSubscriber(topic string, h interface{}, opts ...SubscriberOption) Subscriber {
	return DefaultServer.NewSubscriber(topic, h, opts...)
}

// NewHandler creates a new handler interface using the default server
// Handlers are required to be a public object with public
// endpoint. Call to a service endpoint such as Foo.Bar expects
// the type:
//
// type Foo struct{}
//
//	func (f *Foo) Bar(ctx, req, rsp) error {
//	    return nil
//	}
func NewHandler(h interface{}, opts ...HandlerOption) Handler {
	return DefaultServer.NewHandler(h, opts...)
}

// Handle registers a handler interface with the default server to
// handle inbound requests.
func Handle(h Handler) error {
	return DefaultServer.Handle(h)
}

// Subscribe registers a subscriber interface with the default server
// which subscribes to specified topic with the broker
func Subscribe(s Subscriber) error {
	return DefaultServer.Subscribe(s)
}

// Run starts the default server and waits for a kill
// signal before existing. Also registers/deregisters the server
func Run() error {
	if err := Start(); err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, usignal.Shutdown()...)

	log.Infof("Received signal %s", <-ch)

	return Stop()
}

// Start starts the default server
func Start() error {
	config := DefaultServer.Options()
	log.Infof("Starting server %s id %s", config.Name, config.Id)
	return DefaultServer.Start()
}

// Stop stops the default server
func Stop() error {
	log.Infof("Stopping server")
	return DefaultServer.Stop()
}

// String returns name of Server implementation
func String() string {
	return DefaultServer.String()
}
