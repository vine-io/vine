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

// Package static provides a static resolver which returns the name/ip passed in without any change
package static

import (
	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/core/registry"
)

// staticSelector is a static selector
type staticSelector struct {
	opts selector.Options
}

func (s *staticSelector) Init(opts ...selector.Option) error {
	for _, o := range opts {
		o(&s.opts)
	}
	return nil
}

func (s *staticSelector) Options() selector.Options {
	return s.opts
}

func (s *staticSelector) Select(service string, opts ...selector.SelectOption) (selector.Next, error) {
	return func() (*registry.Node, error) {
		return &registry.Node{
			Id:      service,
			Address: service,
		}, nil
	}, nil
}

func (s *staticSelector) Mark(service string, node *registry.Node, err error) {
	return
}

func (s *staticSelector) Reset(service string) {
	return
}

func (s *staticSelector) Close() error {
	return nil
}

func (s *staticSelector) String() string {
	return "static"
}

func NewSelector(opts ...selector.Option) selector.Selector {
	var options selector.Options
	for _, o := range opts {
		o(&options)
	}
	return &staticSelector{
		opts: options,
	}
}
