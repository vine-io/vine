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

package httprule

// download from https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/master/protoc-gen-grpc-gateway/httprule/types_test.go

import (
	"fmt"
	"testing"
)

func TestTemplateStringer(t *testing.T) {
	for _, spec := range []struct {
		segs []segment
		want string
	}{
		{
			segs: []segment{
				literal("v1"),
			},
			want: "/v1",
		},
		{
			segs: []segment{
				wildcard{},
			},
			want: "/*",
		},
		{
			segs: []segment{
				deepWildcard{},
			},
			want: "/**",
		},
		{
			segs: []segment{
				variable{
					path: "name",
					segments: []segment{
						literal("a"),
					},
				},
			},
			want: "/{name=a}",
		},
		{
			segs: []segment{
				variable{
					path: "name",
					segments: []segment{
						literal("a"),
						wildcard{},
						literal("b"),
					},
				},
			},
			want: "/{name=a/*/b}",
		},
		{
			segs: []segment{
				literal("v1"),
				variable{
					path: "name",
					segments: []segment{
						literal("a"),
						wildcard{},
						literal("b"),
					},
				},
				literal("c"),
				variable{
					path: "field.nested",
					segments: []segment{
						wildcard{},
						literal("d"),
					},
				},
				wildcard{},
				literal("e"),
				deepWildcard{},
			},
			want: "/v1/{name=a/*/b}/c/{field.nested=*/d}/*/e/**",
		},
	} {
		tmpl := template{segments: spec.segs}
		if got, want := tmpl.String(), spec.want; got != want {
			t.Errorf("%#v.String() = %q; want %q", tmpl, got, want)
		}

		tmpl.verb = "LOCK"
		if got, want := tmpl.String(), fmt.Sprintf("%s:LOCK", spec.want); got != want {
			t.Errorf("%#v.String() = %q; want %q", tmpl, got, want)
		}
	}
}
