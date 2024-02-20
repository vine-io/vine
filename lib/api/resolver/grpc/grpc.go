// MIT License
//
// Copyright (c) 2020 The vine Authors
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

// Package grpc resolves a grpc service like /greeter.Say/Hello to greeter service
package grpc

import (
	"errors"
	"net/http"
	"strings"

	"github.com/vine-io/vine/lib/api/resolver"
)

type Resolver struct{}

func (r *Resolver) Resolve(req *http.Request) (*resolver.Endpoint, error) {
	// /foo.Bar/Service
	if req.URL.Path == "/" {
		return nil, errors.New("unknown name")
	}
	// [foo.Bar, Service]
	parts := strings.Split(req.URL.Path[1:], "/")
	// [foo, Bar]
	name := strings.Split(parts[0], ".")
	// foo
	return &resolver.Endpoint{
		Name:   strings.Join(name[:len(name)-1], "."),
		Host:   req.Host,
		Method: req.Method,
		Path:   req.URL.Path,
	}, nil
}

func (r *Resolver) String() string {
	return "grpc"
}

func NewResolver(opts ...resolver.Option) resolver.Resolver {
	return &Resolver{}
}
