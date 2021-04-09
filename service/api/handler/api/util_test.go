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

package api

import (
	"net/http"
	"net/url"
	"testing"
)

func TestRequestToProto(t *testing.T) {
	testData := []*http.Request{
		{
			Method: "GET",
			Header: http.Header{
				"Header": []string{"test"},
			},
			URL: &url.URL{
				Scheme:   "http",
				Host:     "localhost",
				Path:     "/foo/bar",
				RawQuery: "param1=value1",
			},
		},
	}

	for _, d := range testData {
		p, err := requestToProto(d)
		if err != nil {
			t.Fatal(err)
		}
		if p.Path != d.URL.Path {
			t.Fatalf("Expected path %s got %s", d.URL.Path, p.Path)
		}
		if p.Method != d.Method {
			t.Fatalf("Expected method %s got %s", d.Method, p.Method)
		}
		for k, v := range d.Header {
			if val, ok := p.Header[k]; !ok {
				t.Fatalf("Expected header %s", k)
			} else {
				if val.Values[0] != v[0] {
					t.Fatalf("Expected val %s, got %s", val.Values[0], v[0])
				}
			}
		}
	}
}
