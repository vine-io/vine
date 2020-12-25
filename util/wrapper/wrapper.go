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

package wrapper

import (
	"context"
	"reflect"
	"strings"

	"github.com/lack-io/vine/proto/errors"
	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/debug/stats"
	"github.com/lack-io/vine/service/debug/trace"
	"github.com/lack-io/vine/service/server"
	"github.com/lack-io/vine/util/context/metadata"
)

type fromServiceWrapper struct {
	client.Client

	// headers to inject
	headers metadata.Metadata
}

var (
	HeaderPrefix = "Vine-"
)

func (f *fromServiceWrapper) setHeaders(ctx context.Context) context.Context {
	// don't overwrite keys
	return metadata.MergeContext(ctx, f.headers, false)
}

func (f *fromServiceWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	ctx = f.setHeaders(ctx)
	return f.Client.Call(ctx, req, rsp, opts...)
}

func (f *fromServiceWrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	ctx = f.setHeaders(ctx)
	return f.Client.Stream(ctx, req, opts...)
}

func (f *fromServiceWrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
	ctx = f.setHeaders(ctx)
	return f.Client.Publish(ctx, p, opts...)
}

// FromService wraps a client to inject service and auth metadata
func FromService(name string, c client.Client) client.Client {
	return &fromServiceWrapper{
		Client: c,
		headers: metadata.Metadata{
			HeaderPrefix + "From-Service": name,
		},
	}
}

// HandlerStats wraps a server handler to generate request/error stats
func HandlerStats(stats stats.Stats) server.HandlerWrapper {
	// return a handler wrapper
	return func(h server.HandlerFunc) server.HandlerFunc {
		// return a function that returns a function
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			// execute the handler
			err := h(ctx, req, rsp)
			// record the stats
			stats.Record(err)
			// return the error
			return err
		}
	}
}

type traceWrapper struct {
	client.Client

	name  string
	trace trace.Tracer
}

func (c *traceWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	newCtx, s := c.trace.Start(ctx, req.Service()+"."+req.Endpoint())

	s.Type = trace.SpanTypeRequestOutbound
	err := c.Client.Call(newCtx, req, rsp, opts...)
	if err != nil {
		s.Metadata["error"] = err.Error()
	}

	// finish the trace
	c.trace.Finish(s)
	return err
}

// TraceCall is a call tracing wrapper
func TraceCall(name string, t trace.Tracer, c client.Client) client.Client {
	return &traceWrapper{
		name:   name,
		trace:  t,
		Client: c,
	}
}

// TraceHandler wraps a server handler to perform tracing
func TraceHandler(t trace.Tracer) server.HandlerWrapper {
	// return a handler wrapper
	return func(h server.HandlerFunc) server.HandlerFunc {
		// return a function that returns a function
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			// don't store traces for debug
			if strings.HasPrefix(req.Endpoint(), "Debug.") {
				return h(ctx, req, rsp)
			}

			// get the span
			newCtx, s := t.Start(ctx, req.Service()+"."+req.Endpoint())
			s.Type = trace.SpanTypeRequestInbound

			err := h(newCtx, req, rsp)
			if err != nil {
				s.Metadata["error"] = err.Error()
			}

			// finish
			t.Finish(s)

			return err
		}
	}
}

type authWrapper struct {
	client.Client
	auth func() auth.Auth
}

func (a *authWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	// parse the options
	var options client.CallOptions
	for _, o := range opts {
		o(&options)
	}

	// check to see if the authorization header has already been set.
	// We don't override the header unless the ServiceToken option has
	// been specified or the header wasn't provided
	if _, ok := metadata.Get(ctx, "Authorization"); ok && !options.ServiceToken {
		return a.Client.Call(ctx, req, rsp, opts...)
	}

	// if auth is nil we won't be able to get an access token, so we execute
	// the request without one.
	aa := a.auth()
	if aa == nil {
		return a.Client.Call(ctx, req, rsp, opts...)
	}

	// set the namespace header if it has not been set (e.g. on a service to service request)
	if _, ok := metadata.Get(ctx, "Vine-Namespace"); !ok {
		ctx = metadata.Set(ctx, "Vine-Namespace", aa.Options().Namespace)
	}

	// check to see if we have a valid access token
	aaOpts := aa.Options()
	if aaOpts.Token != nil && !aaOpts.Token.Expired() {
		ctx = metadata.Set(ctx, "Authorization", auth.BearerScheme+aaOpts.Token.AccessToken)
		return a.Client.Call(ctx, req, rsp, opts...)
	}

	// call without an auth token
	return a.Client.Call(ctx, req, rsp, opts...)
}

// AuthClient wraps requests with the auth header
func AuthClient(auth func() auth.Auth, c client.Client) client.Client {
	return &authWrapper{c, auth}
}

// AuthHandler wraps a server handler to perform auth
func AuthHandler(fn func() auth.Auth) server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			// get the auth.Auth interface
			a := fn()

			// Check for debug endpoints which should be excluded from auth
			if strings.HasPrefix(req.Endpoint(), "Debug.") {
				return h(ctx, req, rsp)
			}

			// Extract the token if present. Note: if noop is being used
			// then the token can be blank without erroring
			var account *auth.Account
			if header, ok := metadata.Get(ctx, "Authorization"); ok {
				// Ensure the correct scheme is being used
				if !strings.HasPrefix(header, auth.BearerScheme) {
					return errors.Unauthorized(req.Service(), "invalid authorization header. expected Bearer schema")
				}

				// Strip the prefix and inspect the resulting token
				account, _ = a.Inspect(strings.TrimPrefix(header, auth.BearerScheme))
			}

			// Extract the namespace header
			ns, ok := metadata.Get(ctx, "Vine-Namespace")
			if !ok {
				ns = a.Options().Namespace
				ctx = metadata.Set(ctx, "Vine-Namespace", ns)
			}

			// Check the issuer matches the services namespace. TODO: Stop allowing go.vine to access
			// any namespace and instead check for the server issuer.
			if account != nil && account.Issuer != ns && account.Issuer != "go.vine" {
				return errors.Forbidden(req.Service(), "Account was not issued by %v", ns)
			}

			// construct the resource
			res := &auth.Resource{
				Type:     "service",
				Name:     req.Service(),
				Endpoint: req.Endpoint(),
			}

			// Verify the caller has access to the resource
			err := a.Verify(account, res, auth.VerifyContext(ctx))
			if err != nil && account != nil {
				return errors.Forbidden(req.Service(), "Forbidden call mode to %v.%v by %v", req.Service(), req.Endpoint(), account.ID)
			} else if err != nil {
				return errors.Unauthorized(req.Service(), "Unauthorized call mode to %v.%v", req.Service(), req.Endpoint())
			}

			// There is an account, set it in the context
			if account != nil {
				ctx = auth.ContextWithAccount(ctx, account)
			}

			// The user is authorised, allow the call
			return h(ctx, req, rsp)
		}
	}
}

type cacheWrapper struct {
	cacheFn func() *client.Cache
	client.Client
}

// Call executes the request. If the CacheExpiry option was set, the response will be cached using
// a hash of the metadata and request as the key.
func (c *cacheWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	// parse the options
	var options client.CallOptions
	for _, o := range opts {
		o(&options)
	}

	// if the client doesn't have a cache setup don't continue
	cache := c.cacheFn()
	if cache == nil {
		return c.Client.Call(ctx, req, rsp, opts...)
	}

	// if the cache expiry is not set, execute the call without the cache
	if options.CacheExpiry == 0 {
		return c.Client.Call(ctx, req, rsp, opts...)
	}

	// if the response is nil don't call the cache since we can't assign the response
	if rsp == nil {
		return c.Client.Call(ctx, req, rsp, opts...)
	}

	// check to see if there is a response cached, if there is assign it
	if r, ok := cache.Get(ctx, &req); ok {
		val := reflect.ValueOf(rsp).Elem()
		val.Set(reflect.ValueOf(r).Elem())
		return nil
	}

	// don't cache the result if there was an error
	if err := c.Client.Call(ctx, req, rsp, opts...); err != nil {
		return err
	}

	// set the result in the cache
	cache.Set(ctx, &req, rsp, options.CacheExpiry)
	return nil
}

// CacheClient wraps requests with the cache wrapper
func CacheClient(cacheFn func() *client.Cache, c client.Client) client.Client {
	return &cacheWrapper{cacheFn, c}
}

type staticClient struct {
	address string
	client.Client
}

func (s *staticClient) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	return s.Client.Call(ctx, req, rsp, append(opts, client.WithAddress(s.address))...)
}

func (s *staticClient) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	return s.Client.Stream(ctx, req, append(opts, client.WithAddress(s.address))...)
}

// StaticClient sets an address on every call
func StaticClient(address string, c client.Client) client.Client {
	return &staticClient{address, c}
}
