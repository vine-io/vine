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

package http

import (
	"github.com/lack-io/vine"
	"github.com/lack-io/vine/service/server"
)

// WithBackend provides an option to set the http backend url
func WithBackend(url string) vine.Option {
	return func(o *vine.Options) {
		// get the router
		r := o.Server.Options().Router

		// not set
		if r == nil {
			r = DefaultRouter
			o.Server.Init(server.WithRouter(r))
		}

		// check its a http router
		if httpRouter, ok := r.(*Router); ok {
			httpRouter.Backend = url
		}
	}
}

// WithRouter provides an option to set the http router
func WithRouter(r server.Router) vine.Option {
	return func(o *vine.Options) {
		o.Server.Init(server.WithRouter(r))
	}
}
