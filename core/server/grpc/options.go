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

package grpc

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	"github.com/vine-io/vine/core/server"
)

type codecsKey struct{}
type grpcOptions struct{}
type netListener struct{}
type maxMsgSizeKey struct{}
type maxConnKey struct{}
type tlsAuth struct{}
type grpcWithHttp struct{}

// Codec gRPC Codec to be used to encode/decode requests for a given content type
func Codec(contentType string, c encoding.Codec) server.Option {
	return func(o *server.Options) {
		codecs := make(map[string]encoding.Codec)
		if o.Context == nil {
			o.Context = context.Background()
		}
		if v, ok := o.Context.Value(codecsKey{}).(map[string]encoding.Codec); ok && v != nil {
			codecs = v
		}
		codecs[contentType] = c
		o.Context = context.WithValue(o.Context, codecsKey{}, codecs)
	}
}

// AuthTLS should be used to setup a secure authentication using TLS
func AuthTLS(t *tls.Config) server.Option {
	return setServerOption(tlsAuth{}, t)
}

// HttpHandler specifies http and grpc service
func HttpHandler(hh http.Handler) server.Option {
	return setServerOption(grpcWithHttp{}, hh)
}

// MaxConn specifies maximum number of max simultaneous connections to server
func MaxConn(n int) server.Option {
	return setServerOption(maxConnKey{}, n)
}

// Listener specifies the net.Listener to use instead of the default
func Listener(l net.Listener) server.Option {
	return setServerOption(netListener{}, l)
}

// Options to be used to configure gRPC options
func Options(opts ...grpc.ServerOption) server.Option {
	return setServerOption(grpcOptions{}, opts)
}

// MaxMsgSize set the maximum message in bytes the server can receive and
// send. Default maximum message size is 4 MB
func MaxMsgSize(s int) server.Option {
	return setServerOption(maxMsgSizeKey{}, s)
}
