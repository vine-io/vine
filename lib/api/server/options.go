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

package server

import (
	"crypto/tls"
	"net/http"

	"github.com/vine-io/vine/lib/api/resolver"
	"github.com/vine-io/vine/lib/api/server/cors"
)

type Option func(o *Options)

type Options struct {
	EnableCORS bool
	CORSConfig *cors.Config
	EnableTLS  bool
	TLSConfig  *tls.Config
	Resolver   resolver.Resolver
	Wrappers   []Wrapper
}

type Wrapper func(handler http.Handler) http.Handler

func WrapHandler(w Wrapper) Option {
	return func(o *Options) {
		o.Wrappers = append(o.Wrappers, w)
	}
}

func EnableCORS(b bool) Option {
	return func(o *Options) {
		o.EnableCORS = b
	}
}

func CORSConfig(c *cors.Config) Option {
	return func(o *Options) {
		o.CORSConfig = c
	}
}

func EnableTLS(b bool) Option {
	return func(o *Options) {
		o.EnableTLS = b
	}
}

func TLSConfig(t *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}

func Resolver(r resolver.Resolver) Option {
	return func(o *Options) {
		o.Resolver = r
	}
}
