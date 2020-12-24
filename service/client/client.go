// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

<<<<<<<< HEAD:service/client/client.go
import (
	"context"
	"time"

	"github.com/lack-io/vine/internal/codec"
)

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
	// The service to call
	Service() string
	// The action to take
	Method() string
	// The endpoint to invoke
	Endpoint() string
	// The content type
	ContentType() string
	// The unencoded request body
	Body() interface{}
	// Write to the encoded request writer. This is nil before a call is made
	Codec() codec.Writer
	// indicates whether the request will be a streaming one rather than unary
	Stream() bool
}

// Response is the response received from a service
type Response interface {
	// Read the response
	Codec() codec.Reader
	// read the header
	Header() map[string]string
	// Read the undecoded response
	Read() ([]byte, error)
}

// Stream is the inteface for a bidirectional synchronous stream
type Stream interface {
	// Context for the stream
	Context() context.Context
	// The request made
	Request() Request
	// The response read
	Response() Response
	// Send will encode and send a request
	Send(interface{}) error
	// Recv will decode and read a response
	Recv(interface{}) error
	// Error returns the stream error
	Error() error
	// Close closes the stream
	Close() error
}

// Option used by the Client
type Option func(*Options)

// CallOption used by Call or Stream
type CallOption func(*CallOptions)

// PublishOption used by Publish
type PublishOption func(*PublishOptions)

// MessageOption used by NewMessage
type MessageOption func(*MessageOptions)

// RequestOption used by NewRequest
type RequestOption func(*RequestOptions)

var (
	// DefaultClient is a default client to use out of the box
	DefaultClient Client
	// DefaultContentType is the default transport content type
	DefaultContentType = "application/protobuf"
	// DefaultBackoff is the default backoff function for retries
	DefaultBackoff = exponentialBackoff
	// DefaultRetry is the default check-for-retry function for retries
	DefaultRetry = RetryOnError
	// DefaultRetries is the default number of times a request is tried
	DefaultRetries = 1
	// DefaultRequestTimeout is the default request timeout
	DefaultRequestTimeout = time.Second * 5
	// DefaultPoolSize sets the connection pool size
	DefaultPoolSize = 100
	// DefaultPoolTTL sets the connection pool ttl
	DefaultPoolTTL = time.Minute
)

// Makes a synchronous call to a service using the default client
func Call(ctx context.Context, request Request, response interface{}, opts ...CallOption) error {
	return DefaultClient.Call(ctx, request, response, opts...)
}

// Publishes a publication using the default client. Using the underlying broker
// set within the options.
func Publish(ctx context.Context, msg Message, opts ...PublishOption) error {
	return DefaultClient.Publish(ctx, msg, opts...)
}

// Creates a new message using the default client
func NewMessage(topic string, payload interface{}, opts ...MessageOption) Message {
	return DefaultClient.NewMessage(topic, payload, opts...)
}

// Creates a new request using the default client. Content Type will
// be set to the default within options and use the appropriate codec
func NewRequest(service, endpoint string, request interface{}, reqOpts ...RequestOption) Request {
	return DefaultClient.NewRequest(service, endpoint, request, reqOpts...)
}

// Creates a streaming connection with a service and returns responses on the
// channel passed in. It's up to the user to close the streamer.
func NewStream(ctx context.Context, request Request, opts ...CallOption) (Stream, error) {
	return DefaultClient.Stream(ctx, request, opts...)
}

func String() string {
	return DefaultClient.String()
========
import "github.com/lack-io/cli"

// Flags common to all clients
var Flags = []cli.Flag{
	&cli.BoolFlag{
		Name:  "local",
		Usage: "Enable local only development: Defaults to true.",
	},
	&cli.BoolFlag{
		Name:    "enable-acme",
		Usage:   "Enables ACME support via Let's Encrypt. ACME hosts should also be specified.",
		EnvVars: []string{"VINE_ENABLE_ACME"},
	},
	&cli.StringFlag{
		Name:    "acme-hosts",
		Usage:   "Comma separated list of hostnames to manage ACME certs for",
		EnvVars: []string{"VINE_ACME_HOSTS"},
	},
	&cli.StringFlag{
		Name:    "acme-provider",
		Usage:   "The provider that will be used to communicate with Let's Encrypt. Valid options: autocert, certmagic",
		EnvVars: []string{"VINE_ACME_PROVIDER"},
	},
	&cli.BoolFlag{
		Name:    "enable-tls",
		Usage:   "Enable TLS support. Expects cert and key file to be specified",
		EnvVars: []string{"VINE_ENABLE_TLS"},
	},
	&cli.StringFlag{
		Name:    "tls-cert-file",
		Usage:   "Path to the TLS Certificate file",
		EnvVars: []string{"VINE_TLS_CERT_FILE"},
	},
	&cli.StringFlag{
		Name:    "tls-key-file",
		Usage:   "Path to the TLS Key file",
		EnvVars: []string{"VINE_TLS_KEY_FILE"},
	},
	&cli.StringFlag{
		Name:    "tls-client-ca-file",
		Usage:   "Path to the TLS CA file to verify clients against",
		EnvVars: []string{"VINE_TLS_CLIENT_CA_FILE"},
	},
>>>>>>>> e91be0e7314d353e26e2a6aad44a9990713f2d54:client/client.go
}
