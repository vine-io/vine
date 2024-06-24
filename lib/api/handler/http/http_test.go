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

package http

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vine-io/vine/core/registry"
	"github.com/vine-io/vine/core/registry/memory"
	"github.com/vine-io/vine/lib/api/handler"
	"github.com/vine-io/vine/lib/api/resolver"
	"github.com/vine-io/vine/lib/api/resolver/vpath"
	"github.com/vine-io/vine/lib/api/router"
	regRouter "github.com/vine-io/vine/lib/api/router/registry"
)

type ClosedRecorder struct {
	httptest.ResponseRecorder
}

func (r *ClosedRecorder) CloseNotify() <-chan bool {
	ch := make(chan bool, 1)
	return ch
}

func testHttp(t *testing.T, path, service, ns string) {
	r := memory.NewRegistry()

	port := 10000 + rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(1000)
	l, err := net.Listen("tcp", "127.0.0.1:"+fmt.Sprintf("%d", port))
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	s := &registry.Service{
		Name: service,
		Nodes: []*registry.Node{
			{
				Id:      service + "-1",
				Address: l.Addr().String(),
			},
		},
	}

	ctx := context.Background()
	r.Register(ctx, s)
	defer r.Deregister(ctx, s)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	// setup the test handler
	engine.POST(path, func(ctx *gin.Context) {
		ctx.String(http.StatusOK, `you got served`)
	})

	// start http test serve
	go http.Serve(l, engine)

	// create new request and writer
	w := &ClosedRecorder{ResponseRecorder: *httptest.NewRecorder()}
	req, err := http.NewRequest("POST", path, nil)
	if err != nil {
		t.Fatal(err)
	}

	// initialise the handler
	rt := regRouter.NewRouter(
		router.WithHandler("http"),
		router.WithRegistry(r),
		router.WithResolver(vpath.NewResolver(
			resolver.WithNamespace(resolver.StaticNamespace(ns)),
		)),
	)

	p := NewHandler(handler.WithRouter(rt))

	// execute the handler
	c := gin.CreateTestContextOnly(w, engine)
	c.Request = req
	p.Handle(c)

	if w.Code != 200 {
		t.Fatalf("Expected 200 response got %d %s", w.Code, w.Body.String())
	}

	if w.Body.String() != "you got served" {
		t.Fatalf("Expected body: you got served. Got: %s", w.Body.String())
	}
}

func TestHttpHandler(t *testing.T) {
	testData := []struct {
		path      string
		service   string
		namespace string
	}{
		{
			"/test/foo",
			"go.vine.api.test",
			"go.vine.api",
		},
		{
			"/test/foo/baz",
			"go.vine.api.test",
			"go.vine.api",
		},
		{
			"/v1/foo",
			"go.vine.api.v1.foo",
			"go.vine.api",
		},
		{
			"/v1/foo/bar",
			"go.vine.api.v1.foo",
			"go.vine.api",
		},
		{
			"/v2/baz",
			"go.vine.api.v2.baz",
			"go.vine.api",
		},
		{
			"/v2/baz/bar",
			"go.vine.api.v2.baz",
			"go.vine.api",
		},
		{
			"/v2/baz/bar",
			"v2.baz",
			"",
		},
	}

	for _, d := range testData {
		t.Run(d.service, func(t *testing.T) {
			testHttp(t, d.path, d.service, d.namespace)
			time.Sleep(time.Second)
		})
	}
}
