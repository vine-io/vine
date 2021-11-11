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

package context

import (
	"context"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/vine-io/vine/util/context/metadata"
)

func FromRequest(r *http.Request) context.Context {
	md, ok := metadata.FromContext(r.Context())
	if !ok {
		md = make(metadata.Metadata)
	}
	for key, values := range r.Header {
		switch key {
		case "Connection", "Content-Length":
			continue
		}
		md.Set(textproto.CanonicalMIMEHeaderKey(key), strings.Join(values, ","))
	}
	if v, ok := md.Get("X-Forwarded-For"); ok {
		md["X-Forwarded-For"] = v + ", " + r.RemoteAddr
	} else {
		md["X-Forwarded-For"] = r.RemoteAddr
	}
	if _, ok = md.Get("Host"); !ok {
		md.Set("Host", r.Host)
	}
	md.Set("Method", r.Method)
	if _, ok = md.Get("Vine-Api-Path"); !ok {
		if r.URL != nil {
			md.Set("Vine-Api-Path", r.URL.Path)
		}
	}
	return metadata.NewContext(r.Context(), md)
}
