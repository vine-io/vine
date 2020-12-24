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

package wrapper_test

import (
	"context"
	"testing"

	tmemory "github.com/lack-io/vine/internal/network/transport/memory"
	wrapper "github.com/lack-io/vine/internal/wrapper"
	"github.com/lack-io/vine/service/broker"
	bmemory "github.com/lack-io/vine/service/broker/memory"
	"github.com/lack-io/vine/service/client"
	rmemory "github.com/lack-io/vine/service/registry/memory"
	"github.com/lack-io/vine/service/server"
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

	req := client.NewRequest("go.vine.service.foo", "TestFoo.Bar", &TestReq{}, client.WithContentType("application/json"))
	rsp := &TestRsp{}

	reg := rmemory.NewRegistry()
	brk := bmemory.NewBroker(broker.Registry(reg))
	tr := tmemory.NewTransport()

	srv := server.NewServer(
		server.Broker(brk),
		server.Registry(reg),
		server.Name("go.vine.service.foo"),
		server.Address("127.0.0.1:0"),
		server.Transport(tr),
	)
	if err = srv.Handle(srv.NewHandler(&TestFoo{})); err != nil {
		t.Fatal(err)
	}

	if err = srv.Start(); err != nil {
		t.Fatal(err)
	}

	cli := client.NewClient(
		client.Registry(reg),
		client.Broker(brk),
		client.Transport(tr),
	)

	w1 := wrapper.StaticClient("xxx_localhost:12345", cli)
	if err = w1.Call(context.TODO(), req, nil); err == nil {
		t.Fatal("address xxx_#localhost:12345 must not exists and call must be failed")
	}

	w2 := wrapper.StaticClient(srv.Options().Address, cli)
	if err = w2.Call(context.TODO(), req, rsp); err != nil {
		t.Fatal(err)
	} else if rsp.Data != "pass" {
		t.Fatalf("something wrong with response: %#+v", rsp)
	}
}
