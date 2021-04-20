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

// Package http resolves names to network addresses using a http request
package http

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/lack-io/vine/lib/network/resolver"
)

// Resolver is a HTTP network resolver
type Resolver struct {
	// If not set, defaults to http
	Proto string

	// Path sets the path to lookup. Defaults to /network
	Path string

	// Host url to use for the query
	Host string
}

type Response struct {
	Nodes []*resolver.Record `json:"nodes,omitempty"`
}

// Resolve assumes ID is a domain which can be converted to a http://name/network request
func (r *Resolver) Resolve(name string) ([]*resolver.Record, error) {
	proto := "https"
	host := "go.vine.mu"
	path := "/network/nodes"

	if len(r.Proto) > 0 {
		proto = r.Proto
	}

	if len(r.Path) > 0 {
		path = r.Path
	}

	if len(r.Host) > 0 {
		host = r.Host
	}

	uri := &url.URL{
		Scheme: proto,
		Path:   path,
		Host:   host,
	}
	q := uri.Query()
	q.Set("name", name)
	uri.RawQuery = q.Encode()

	rsp, err := http.Get(uri.String())
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		return nil, errors.New("non 200 response")
	}
	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	// encoding format is assumed to be json
	var response *Response

	if err := json.Unmarshal(b, &response); err != nil {
		return nil, err
	}

	return response.Nodes, nil
}
