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

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	 "github.com/lack-io/vine/core/client"
	 "github.com/lack-io/vine/core/client/selector"
	 "github.com/lack-io/vine/core/registry/memory"
	"github.com/lack-io/vine/core/server"
	"github.com/lack-io/vine/lib/config/cmd"
	"github.com/lack-io/vine/util/context/metadata"
)

type TestHandler struct {
	t      *testing.T
	expect metadata.Metadata
}

type TestRequest struct{}
type TestResponse struct{}

func (t *TestHandler) Exec(ctx context.Context, req *TestRequest, rsp *TestResponse) error {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return fmt.Errorf("Expected metadata got %t", ok)
	}

	for k, v := range t.expect {
		if val := md[k]; val != v {
			return fmt.Errorf("Expected %s for key %s got %s", v, k, val)
		}
	}

	t.t.Logf("Received request %+v", req)
	t.t.Logf("Received metadata %+v", md)

	return nil
}

func TestRPCHandler(t *testing.T) {
	r := memory.NewRegistry()

	(*cmd.DefaultOptions().Client).Init(
		client.Registry(r),
		client.Selector(selector.NewSelector(selector.Registry(r))),
	)

	(*cmd.DefaultOptions().Server).Init(
		server.Name("test"),
		server.Registry(r),
	)

	(*cmd.DefaultOptions().Server).Handle(
		(*cmd.DefaultOptions().Server).NewHandler(&TestHandler{t, metadata.Metadata{"Foo": "Bar"}}),
	)

	if err := server.Start(); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	request := map[string]string{
		"service":  "test",
		"endpoint": "TestHandler.Exec",
		"request":  "{}",
	}

	rb, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBuffer(rb)

	req, err := http.NewRequest("POST", "/rpc", b)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Foo", "Bar")

	//RPC(w, req)

	if err := server.Stop(); err != nil {
		t.Fatal(err)
	}

	if w.Code != 200 {
		t.Fatalf("Expected 200 response got %d %s", w.Code, w.Body.String())
	}

}
