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

package api

import (
	"fmt"
	"mime"
	"net"
	"net/http"
	"strings"

	"github.com/oxtoacart/bpool"
	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/core/registry"
	"github.com/vine-io/vine/lib/api"
)

var (
	// need to calculate later to specify useful defaults
	bufferPool = bpool.NewSizedBufferPool(1024, 8)
)

func requestToProto(r *http.Request) (*api.Request, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("Error parsing form: %v", err)
	}

	req := &api.Request{
		Path:   r.URL.Path,
		Method: r.Method,
		Header: make(map[string]*api.Pair),
		Get:    make(map[string]*api.Pair),
		Post:   make(map[string]*api.Pair),
		Url:    r.URL.String(),
	}

	ct, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		ct = "text/plain; charset=UTF-8" //default CT is text/plain
		r.Header.Set("Content-Type", ct)
	}

	//set the body:
	if r.Body != nil {
		switch ct {
		case "application/x-www-form-urlencoded":
			// expect form vals in Post data
		default:
			buf := bufferPool.Get()
			defer bufferPool.Put(buf)
			if _, err = buf.ReadFrom(r.Body); err != nil {
				return nil, err
			}
			req.Body = buf.String()
		}
	}

	// Set X-Forwarded-For if it does not exist
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if prior, ok := r.Header["X-Forwarded-For"]; ok {
			ip = strings.Join(prior, ", ") + ", " + ip
		}

		// Set the header
		req.Header["X-Forwarded-For"] = &api.Pair{
			Key:    "X-Forwarded-For",
			Values: []string{ip},
		}
	}

	// Host is stripped from net/http Headers so let's add it
	req.Header["Host"] = &api.Pair{
		Key:    "Host",
		Values: []string{r.Host},
	}

	// Get data
	for key, vals := range r.URL.Query() {
		header, ok := req.Get[key]
		if !ok {
			header = &api.Pair{
				Key: key,
			}
			req.Get[key] = header
		}
		header.Values = vals
	}

	// Post data
	for key, vals := range r.PostForm {
		header, ok := req.Post[key]
		if !ok {
			header = &api.Pair{
				Key: key,
			}
			req.Post[key] = header
		}
		header.Values = vals
	}

	for key, vals := range r.Header {
		header, ok := req.Header[key]
		if !ok {
			header = &api.Pair{
				Key: key,
			}
			req.Header[key] = header
		}
		header.Values = vals
	}

	return req, nil
}

// strategy is a hack for selection
func strategy(services []*registry.Service) selector.Strategy {
	return func(_ []*registry.Service) selector.Next {
		// ignore input to this function, use services above
		return selector.Random(services)
	}
}
