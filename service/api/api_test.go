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
	"strings"
	"testing"

	apipb "github.com/lack-io/vine/proto/apis/api"
)

func TestEncoding(t *testing.T) {
	testData := []*apipb.Endpoint{
		nil,
		{
			Name:        "Foo.Bar",
			Description: "A test endpoint",
			Handler:     "meta",
			Host:        []string{"foo.com"},
			Method:      []string{"GET"},
			Path:        []string{"/test"},
		},
	}

	compare := func(expect, got []string) bool {
		// no data to compare, return true
		if len(expect) == 0 && len(got) == 0 {
			return true
		}
		// no data expected but got some return false
		if len(expect) == 0 && len(got) > 0 {
			return false
		}

		// compare expected with what we got
		for _, e := range expect {
			var seen bool
			for _, g := range got {
				if e == g {
					seen = true
					break
				}
			}
			if !seen {
				return false
			}
		}

		// we're done, return true
		return true
	}

	for _, d := range testData {
		// encode
		e := Encode(d)
		// decode
		de := Decode(e)

		// nil endpoint returns nil
		if d == nil {
			if e != nil {
				t.Fatalf("expected nil got %v", e)
			}
			if de != nil {
				t.Fatalf("expected nil got %v", de)
			}

			continue
		}

		// check encoded map
		name := e["endpoint"]
		desc := e["description"]
		method := strings.Split(e["method"], ",")
		path := strings.Split(e["path"], ",")
		host := strings.Split(e["host"], ",")
		handler := e["handler"]

		if name != d.Name {
			t.Fatalf("expected %v got %v", d.Name, name)
		}
		if desc != d.Description {
			t.Fatalf("expected %v got %v", d.Description, desc)
		}
		if handler != d.Handler {
			t.Fatalf("expected %v got %v", d.Handler, handler)
		}
		if ok := compare(d.Method, method); !ok {
			t.Fatalf("expected %v got %v", d.Method, method)
		}
		if ok := compare(d.Path, path); !ok {
			t.Fatalf("expected %v got %v", d.Path, path)
		}
		if ok := compare(d.Host, host); !ok {
			t.Fatalf("expected %v got %v", d.Host, host)
		}

		if de.Name != d.Name {
			t.Fatalf("expected %v got %v", d.Name, de.Name)
		}
		if de.Description != d.Description {
			t.Fatalf("expected %v got %v", d.Description, de.Description)
		}
		if de.Handler != d.Handler {
			t.Fatalf("expected %v got %v", d.Handler, de.Handler)
		}
		if ok := compare(d.Method, de.Method); !ok {
			t.Fatalf("expected %v got %v", d.Method, de.Method)
		}
		if ok := compare(d.Path, de.Path); !ok {
			t.Fatalf("expected %v got %v", d.Path, de.Path)
		}
		if ok := compare(d.Host, de.Host); !ok {
			t.Fatalf("expected %v got %v", d.Host, de.Host)
		}
	}
}

func TestValidate(t *testing.T) {
	epPcre := &apipb.Endpoint{
		Name:        "Foo.Bar",
		Description: "A test endpoint",
		Handler:     "meta",
		Host:        []string{"foo.com"},
		Method:      []string{"GET"},
		Path:        []string{"^/test/?$"},
	}
	if err := Validate(epPcre); err != nil {
		t.Fatal(err)
	}

	epGpath := &apipb.Endpoint{
		Name:        "Foo.Bar",
		Description: "A test endpoint",
		Handler:     "meta",
		Host:        []string{"foo.com"},
		Method:      []string{"GET"},
		Path:        []string{"/test/{id}"},
	}
	if err := Validate(epGpath); err != nil {
		t.Fatal(err)
	}

	epPcreInvalid := &apipb.Endpoint{
		Name:        "Foo.Bar",
		Description: "A test endpoint",
		Handler:     "meta",
		Host:        []string{"foo.com"},
		Method:      []string{"GET"},
		Path:        []string{"/test/?$"},
	}
	if err := Validate(epPcreInvalid); err == nil {
		t.Fatalf("invalid pcre %v", epPcreInvalid.Path[0])
	}

}
