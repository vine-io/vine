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

package helper

import (
	"net/http"
	"testing"

	"github.com/lack-io/vine/util/context/metadata"
)

func TestRequestToContext(t *testing.T) {
	testData := []struct {
		request *http.Request
		expect  metadata.Metadata
	}{
		{
			&http.Request{
				Header: http.Header{
					"Foo1": []string{"bar"},
					"Foo2": []string{"bar", "baz"},
				},
			},
			metadata.Metadata{
				"Foo1": "bar",
				"Foo2": "bar,baz",
			},
		},
	}

	for _, d := range testData {
		ctx := RequestToContext(d.request)
		md, ok := metadata.FromContext(ctx)
		if !ok {
			t.Fatalf("Expected metadata for request %+v", d.request)
		}
		for k, v := range d.expect {
			if val := md[k]; val != v {
				t.Fatalf("Expected %s for key %s for expected md %+v, got md %+v", v, k, d.expect, md)
			}
		}
	}
}
