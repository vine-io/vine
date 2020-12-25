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

// Package resolver resolves a http request to an endpoint
package resolver

import (
	"errors"
	"net/http"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrInvalidPath = errors.New("invalid path")
)

// Resolver resolves requests to endpoints
type Resolver interface {
	Resolve(r *http.Request) (*Endpoint, error)
	String() string
}

// Endpoint is the endpoint for a http request
type Endpoint struct {
	// e.g greeter
	Name string
	// HTTP Host e.g example.com
	Host string
	// HTTP Methods e.g GET, POST
	Method string
	// HTTP Path e.g /greeter.
	Path string
}

type Options struct {
	Handler   string
	Namespace func(*http.Request) string
}

type Option func(o *Options)

// StaticNamespace returns the same namespace for each request
func StaticNamespace(ns string) func(*http.Request) string {
	return func(*http.Request) string {
		return ns
	}
}
