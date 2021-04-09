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

package util

// download from https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/master/protoc-gen-grpc-gateway/httprule/types.go

import (
	"fmt"
	"strings"
)

type template struct {
	segments []segment
	verb     string
	template string
}

type segment interface {
	fmt.Stringer
	compile() (ops []op)
}

type wildcard struct{}

type deepWildcard struct{}

type literal string

type variable struct {
	path     string
	segments []segment
}

func (wildcard) String() string {
	return "*"
}

func (deepWildcard) String() string {
	return "**"
}

func (l literal) String() string {
	return string(l)
}

func (v variable) String() string {
	var segs []string
	for _, s := range v.segments {
		segs = append(segs, s.String())
	}
	return fmt.Sprintf("{%s=%s}", v.path, strings.Join(segs, "/"))
}

func (t template) String() string {
	var segs []string
	for _, s := range t.segments {
		segs = append(segs, s.String())
	}
	str := strings.Join(segs, "/")
	if t.verb != "" {
		str = fmt.Sprintf("%s:%s", str, t.verb)
	}
	return "/" + str
}
