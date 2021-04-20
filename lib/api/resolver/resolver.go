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

// Package resolver resolves a http request to an endpoint
package resolver

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrInvalidPath = errors.New("invalid path")
)

// Resolver resolves requests to endpoints
type Resolver interface {
	Resolve(c *fiber.Ctx) (*Endpoint, error)
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
	Namespace func(ctx *fiber.Ctx) string
}

type Option func(o *Options)

// StaticNamespace returns the same namespace for each request
func StaticNamespace(ns string) func(*fiber.Ctx) string {
	return func(*fiber.Ctx) string {
		return ns
	}
}
