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

// Package path resolves using http path
package path

import (
	"net/http"
	"strings"

	"github.com/lack-io/vine/util/api/resolver"
)

type Resolver struct {
	opts resolver.Options
}

func (r *Resolver) Resolve(req *http.Request) (*resolver.Endpoint, error) {
	if req.URL.Path == "/" {
		return nil, resolver.ErrNotFound
	}

	parts := strings.Split(req.URL.Path[1:], "/")
	ns := r.opts.Namespace(req)

	return &resolver.Endpoint{
		Name:   ns + "." + parts[0],
		Host:   req.Host,
		Method: req.Method,
		Path:   req.URL.Path,
	}, nil
}

func (r *Resolver) String() string {
	return "path"
}

func NewResolver(opts ...resolver.Option) resolver.Resolver {
	return &Resolver{opts: resolver.NewOptions(opts...)}
}
