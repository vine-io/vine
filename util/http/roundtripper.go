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

package http

import (
	"errors"
	"net/http"

	"github.com/lack-io/vine/core/client/selector"
)

type roundTripper struct {
	rt   http.RoundTripper
	st   selector.Strategy
	opts Options
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	s, err := r.opts.Registry.GetService(req.URL.Host)
	if err != nil {
		return nil, err
	}

	next := r.st(s)

	// rudimentary retry 3 times
	for i := 0; i < 3; i++ {
		n, err := next()
		if err != nil {
			continue
		}
		req.URL.Host = n.Address
		w, err := r.rt.RoundTrip(req)
		if err != nil {
			continue
		}
		return w, nil
	}

	return nil, errors.New("failed request")
}
