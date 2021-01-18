// Copyright 2020 lack
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

package resolver

import (
	"net/http"
)

// NewOptions returns new initialised options
func NewOptions(opts ...Option) Options {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.Namespace == nil {
		options.Namespace = StaticNamespace("go.vine")
	}

	return options
}

// WithHandler sets the handler being used
func WithHandler(h string) Option {
	return func(o *Options) {
		o.Handler = h
	}
}

// WithNamespace sets the function which determines the namespace for a request
func WithNamespace(n func(*http.Request) string) Option {
	return func(o *Options) {
		o.Namespace = n
	}
}
