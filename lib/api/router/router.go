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

// Package router provides api service routing
package router

import (
	"net/http"

	"github.com/vine-io/vine/lib/api"
)

// Router is used to determine an endpoint for a request
type Router interface {
	// Options returns options
	Options() Options
	// Close stop the router
	Close() error
	// Endpoint returns an api.Service endpoint or an error if it does not exist
	Endpoint(r *http.Request) (*api.Service, error)
	// Register endpoint in router
	Register(ep *api.Endpoint) error
	// Deregister endpoint from router
	Deregister(ep *api.Endpoint) error
	// Route returns an api.Service route
	Route(r *http.Request) (*api.Service, error)
}
