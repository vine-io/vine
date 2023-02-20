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

package client

import (
	"context"
	"time"

	"github.com/spf13/pflag"
	"github.com/vine-io/vine/core/codec"
)

var (
	// DefaultClient is a default client to use out of the box
	DefaultClient Client
	// DefaultContentType is the default content type for client
	DefaultContentType = "application/protobuf"
	// DefaultBackoff is the default backoff function for retries
	DefaultBackoff = exponentialBackoff
	// DefaultRetry is the default check-for-retry function for retries
	DefaultRetry = RetryOnError
	// DefaultRetries is the default number of times a request is tried
	DefaultRetries = 1
	// DefaultDialTimeout is the default dial timeout
	DefaultDialTimeout = time.Second * 30
	// DefaultRequestTimeout is the default request timeout
	DefaultRequestTimeout = time.Second * 30
	// DefaultPoolSize sets the connection pool size
	DefaultPoolSize = 100
	// DefaultPoolTTL sets the connection pool ttl
	DefaultPoolTTL = time.Minute

	Flag = pflag.NewFlagSet("client", pflag.ExitOnError)
)

func init() {
	Flag.String("client.default", "", "Client for vine")
	Flag.String("client.content-type", "", "Sets the content type for client")
	Flag.Int("client.retries", 0, "Sets the retries")
	Flag.Duration("client.dial-timeout", 0, "Sets the client dial timeout")
	Flag.Duration("client.request-timeout", 0, "Sets the client request timeout")
	Flag.Int("client.pool-size", 0, "Sets the client connection pool size")
	Flag.Duration("client.pool-ttl", 0, "Sets the client connection pool ttl")
}

// Client is the interface used to make requests to services.
// It supports Request/Response via Transport and Publishing via the Broker.
// It also supports bidirectional streaming of requests.
type Client interface {
	Init(...Option) error
	Options() Options
	NewMessage(topic string, msg interface{}, opts ...MessageOption) Message
	NewRequest(service, endpoint string, req interface{}, reqOpts ...RequestOption) Request
	Call(ctx context.Context, req Request, rsp interface{}, opts ...CallOption) error
	Stream(ctx context.Context, req Request, opts ...CallOption) (Stream, error)
	Publish(ctx context.Context, msg Message, opts ...PublishOption) error
	String() string
}

// Router manages request routing
type Router interface {
	SendRequest(context.Context, Request) (Response, error)
}

// Message is the interface for publishing asynchronously
type Message interface {
	Topic() string
	Payload() interface{}
	ContentType() string
}

// Request is the interface for a synchronous request used by Call or Stream
type Request interface {
	// Service the service to call
	Service() string
	// Method the action to take
	Method() string
	// Endpoint the endpoint to invoke
	Endpoint() string
	// ContentType the content type
	ContentType() string
	// Body the unencoded request body
	Body() interface{}
	// Codec writes to the encoded request writer. This is nil before a call is made
	Codec() codec.Writer
	// Stream indicates whether the request will be a streaming one rather than unary
	Stream() bool
}

// Response is the response received from a service
type Response interface {
	// Codec reads the response
	Codec() codec.Reader
	// Header reads the header
	Header() map[string]string
	// Read the undecoded response
	Read() ([]byte, error)
}

// Stream is the interface for a bidirectional synchronous stream
type Stream interface {
	// Context for the stream
	Context() context.Context
	// Request the request made
	Request() Request
	// Response the response read
	Response() Response
	// Send will encode and send a request
	Send(interface{}) error
	// Recv will decode and read a response
	Recv(interface{}) error
	// Error returns the stream error
	Error() error
	// Close closes the stream and close Conn
	Close() error
	// CloseSend closes the stream
	CloseSend() error
}

// Call makes a synchronous call to a service using the default client
func Call(ctx context.Context, request Request, response interface{}, opts ...CallOption) error {
	return DefaultClient.Call(ctx, request, response, opts...)
}

// Publish publishes a publication using the default client. Using the underlying broker
// set within the options.
func Publish(ctx context.Context, msg Message, opts ...PublishOption) error {
	return DefaultClient.Publish(ctx, msg, opts...)
}

// NewMessage creates a new message using the default client
func NewMessage(topic string, payload interface{}, opts ...MessageOption) Message {
	return DefaultClient.NewMessage(topic, payload, opts...)
}

// NewRequest creates a new request using the default client. Content Type will
// be set to the default within options and use the appropriate codec
func NewRequest(service, endpoint string, request interface{}, reqOpts ...RequestOption) Request {
	return DefaultClient.NewRequest(service, endpoint, request, reqOpts...)
}

// NewStream creates a streaming connection with a service and returns responses on the
// channel passed in. It's up to the user to close the streamer.
func NewStream(ctx context.Context, request Request, opts ...CallOption) (Stream, error) {
	return DefaultClient.Stream(ctx, request, opts...)
}

func String() string {
	return DefaultClient.String()
}
