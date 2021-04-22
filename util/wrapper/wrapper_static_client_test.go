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

package wrapper_test

import (
	"context"
	"testing"

	"github.com/lack-io/vine/core/broker"
	brokerMem "github.com/lack-io/vine/core/broker/memory"
	 "github.com/lack-io/vine/core/client"
	clientMucp "github.com/lack-io/vine/core/client/mucp"
	regMem "github.com/lack-io/vine/core/registry/memory"
	"github.com/lack-io/vine/core/server"
	serverMucp "github.com/lack-io/vine/core/server/mucp"
	tmemory "github.com/lack-io/vine/core/transport/memory"
	"github.com/lack-io/vine/util/wrapper"
)

type TestFoo struct {
}

type TestReq struct{}

type TestRsp struct {
	Data string
}

func (h *TestFoo) Bar(ctx context.Context, req *TestReq, rsp *TestRsp) error {
	rsp.Data = "pass"
	return nil
}

func TestStaticClientWrapper(t *testing.T) {
	var err error

	reg := regMem.NewRegistry()
	brk := brokerMem.NewBroker(broker.Registry(reg))
	tr := tmemory.NewTransport()

	svc := serverMucp.NewServer(
		server.Broker(brk),
		server.Registry(reg),
		server.Name("go.vine.service.foo"),
		server.Address("127.0.0.1:0"),
		server.Transport(tr),
	)
	if err = svc.Handle(svc.NewHandler(&TestFoo{})); err != nil {
		t.Fatal(err)
	}

	if err = svc.Start(); err != nil {
		t.Fatal(err)
	}

	cli := clientMucp.NewClient(
		client.Registry(reg),
		client.Broker(brk),
		client.Transport(tr),
	)

	req := cli.NewRequest("go.vine.service.foo", "TestFoo.Bar", &TestReq{}, client.WithContentType("application/json"))
	rsp := &TestRsp{}

	w1 := wrapper.StaticClient("xxx_localhost:12345", cli)
	if err = w1.Call(context.TODO(), req, nil); err == nil {
		t.Fatal("address xxx_#localhost:12345 must not exists and call must be failed")
	}

	w2 := wrapper.StaticClient(svc.Options().Address, cli)
	if err = w2.Call(context.TODO(), req, rsp); err != nil {
		t.Fatal(err)
	} else if rsp.Data != "pass" {
		t.Fatalf("something wrong with response: %#+v", rsp)
	}
}
