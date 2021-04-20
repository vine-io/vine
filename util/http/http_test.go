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
	"io/ioutil"
	"net"
	"net/http"
	"testing"

	memory2 "github.com/lack-io/vine/core/registry/memory"
	regpb "github.com/lack-io/vine/proto/apis/registry"
)

func TestRoundTripper(t *testing.T) {
	m := memory2.NewRegistry()

	rt := NewRoundTripper(
		WithRegistry(m),
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`hello world`))
	})

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go http.Serve(l, nil)

	m.Register(&regpb.Service{
		Name: "example.com",
		Nodes: []*regpb.Node{
			{
				Id:      "1",
				Address: l.Addr().String(),
			},
		},
	})

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	w, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	w.Body.Close()

	if string(b) != "hello world" {
		t.Fatal("response is", string(b))
	}

	// test http request
	c := &http.Client{
		Transport: rt,
	}

	rsp, err := c.Get("http://example.com")
	if err != nil {
		t.Fatal(err)
	}

	b, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		t.Fatal(err)
	}
	rsp.Body.Close()

	if string(b) != "hello world" {
		t.Fatal("response is", string(b))
	}

}
