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

// Package vpath resolves using http path and recognised versioned urls
package vpath

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/lack-io/vine/service/api/resolver"
)

func NewResolver(opts ...resolver.Option) resolver.Resolver {
	return &Resolver{opts: resolver.NewOptions(opts...)}
}

type Resolver struct {
	opts resolver.Options
}

var (
	re = regexp.MustCompile("^v[0-9]+$")
)

func (r *Resolver) Resolve(req *http.Request) (*resolver.Endpoint, error) {
	if req.URL.Path == "/" {
		return nil, errors.New("unknown name")
	}

	parts := strings.Split(req.URL.Path[1:], "/")
	if len(parts) == 1 {
		return &resolver.Endpoint{
			Name:   r.withNamespace(req, parts...),
			Host:   req.Host,
			Method: req.Method,
			Path:   req.URL.Path,
		}, nil
	}

	// /v1/foo
	if re.MatchString(parts[0]) {
		return &resolver.Endpoint{
			Name:   r.withNamespace(req, parts[0:2]...),
			Host:   req.Host,
			Method: req.Method,
			Path:   req.URL.Path,
		}, nil
	}

	return &resolver.Endpoint{
		Name:   r.withNamespace(req, parts[0]),
		Host:   req.Host,
		Method: req.Method,
		Path:   req.URL.Path,
	}, nil
}

func (r *Resolver) String() string {
	return "path"
}

func (r *Resolver) withNamespace(req *http.Request, parts ...string) string {
	ns := r.opts.Namespace(req)
	if len(ns) == 0 {
		return strings.Join(parts, ".")
	}

	return strings.Join(append([]string{ns}, parts...), ".")
}
