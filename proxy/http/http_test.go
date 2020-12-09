// Copyright 2020 The vine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"

	"github.com/lack-io/vine"
	"github.com/lack-io/vine/client"
	"github.com/lack-io/vine/registry/memory"
	"github.com/lack-io/vine/server"
)

type testHandler struct{}

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"hello": "world"}`))
}

func TestHTTPProxy(t *testing.T) {
	c, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	addr := c.Addr().String()

	url := fmt.Sprintf("http://%s", addr)

	testCases := []struct {
		// http endpoint to call e.g /foo/bar
		httpEp string
		// rpc endpoint called e.g Foo.Bar
		rpcEp string
		// should be an error
		err bool
	}{
		{"/", "Foo.Bar", false},
		{"/", "Foo.Baz", false},
		{"/helloworld", "Hello.World", true},
	}

	// handler
	http.Handle("/", new(testHandler))

	// new proxy
	p := NewSingleHostProxy(url)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	// new vine service
	service := vine.NewService(
		vine.Context(ctx),
		vine.Name("foobar"),
		vine.Registry(memory.NewRegistry()),
		vine.AfterStart(func() error {
			wg.Done()
			return nil
		}),
	)

	// set router
	service.Server().Init(
		server.WithRouter(p),
	)

	// run service
	// server
	go http.Serve(c, nil)
	go service.Run()

	// wait till service is started
	wg.Wait()

	for _, test := range testCases {
		req := service.Client().NewRequest("foobar", test.rpcEp, map[string]string{"foo": "bar"}, client.WithContentType("application/json"))
		var rsp map[string]string
		err := service.Client().Call(ctx, req, &rsp)
		if err != nil && test.err == false {
			t.Fatal(err)
		}
		if v := rsp["hello"]; v != "world" {
			t.Fatalf("Expected hello world got %s from %s", v, test.rpcEp)
		}
	}
}
