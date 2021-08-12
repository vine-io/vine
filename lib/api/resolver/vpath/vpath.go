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

// Package vpath resolves using http path and recognised versioned urls
package vpath

import (
	"errors"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/vine-io/vine/lib/api/resolver"
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

func (r *Resolver) Resolve(c *fiber.Ctx) (*resolver.Endpoint, error) {
	if c.Path() == "/" {
		return nil, errors.New("unknown name")
	}

	parts := strings.Split(c.Path()[1:], "/")
	if len(parts) == 1 {
		return &resolver.Endpoint{
			Name:   r.withNamespace(c, parts...),
			Host:   string(c.Request().Host()),
			Method: c.Method(),
			Path:   c.Path(),
		}, nil
	}

	// /v1/foo
	if re.MatchString(parts[0]) {
		return &resolver.Endpoint{
			Name:   r.withNamespace(c, parts[0:2]...),
			Host:   string(c.Request().Host()),
			Method: c.Method(),
			Path:   c.Path(),
		}, nil
	}

	return &resolver.Endpoint{
		Name:   r.withNamespace(c, parts[0]),
		Host:   string(c.Request().Host()),
		Method: c.Method(),
		Path:   c.Path(),
	}, nil
}

func (r *Resolver) String() string {
	return "path"
}

func (r *Resolver) withNamespace(c *fiber.Ctx, parts ...string) string {
	ns := r.opts.Namespace(c)
	if len(ns) == 0 {
		return strings.Join(parts, ".")
	}

	return strings.Join(append([]string{ns}, parts...), ".")
}
