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

package grpc

import (
	"context"
	"crypto/tls"
	"sync"
	"testing"
	"time"

	"github.com/lack-io/vine/registry/memory"
	"github.com/lack-io/vine/service"
	hello "github.com/lack-io/vine/service/grpc/proto"
	mls "github.com/lack-io/vine/util/tls"
)

type testHandler struct{}

func (t *testHandler) Call(ctx context.Context, req *hello.Request, rsp *hello.Response) error {
	rsp.Msg = "Hello " + req.Name
	return nil
}

func TestGRPCService(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create memory registry
	r := memory.NewRegistry()

	// create GRPC service
	service := NewService(
		service.Name("test.service"),
		service.Registry(r),
		service.AfterStart(func() error {
			wg.Done()
			return nil
		}),
		service.Context(ctx),
	)

	// register test handler
	hello.RegisterTestHandler(service.Server(), &testHandler{})

	// run service
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		errCh <- service.Run()
	}()

	// wait for start
	wg.Wait()

	// create client
	test := hello.NewTestService("test.service", service.Client())

	// call service
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Duration(time.Second))
	defer cancel2()
	rsp, err := test.Call(ctx2, &hello.Request{
		Name: "John",
	})
	if err != nil {
		t.Fatal(err)
	}

	// check server
	select {
	case err := <-errCh:
		t.Fatal(err)
	case <-time.After(time.Second):
		break
	}

	// check message
	if rsp.Msg != "Hello John" {
		t.Fatalf("unexpected response %s", rsp.Msg)
	}
}

func TestGRPCTLSService(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create memory registry
	r := memory.NewRegistry()

	// create cert
	cert, err := mls.Certificate("test.service")
	if err != nil {
		t.Fatal(err)
	}
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}

	// create GRPC service
	service := NewService(
		service.Name("test.service"),
		service.Registry(r),
		service.AfterStart(func() error {
			wg.Done()
			return nil
		}),
		service.Context(ctx),
		// set TLS config
		WithTLS(config),
	)

	// register test handler
	hello.RegisterTestHandler(service.Server(), &testHandler{})

	// run service
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		errCh <- service.Run()
	}()

	// wait for start
	wg.Wait()

	// create client
	test := hello.NewTestService("test.service", service.Client())

	// call service
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Duration(time.Second))
	defer cancel2()
	rsp, err := test.Call(ctx2, &hello.Request{
		Name: "John",
	})
	if err != nil {
		t.Fatal(err)
	}

	// check server
	select {
	case err := <-errCh:
		t.Fatal(err)
	case <-time.After(time.Second):
		break
	}

	// check message
	if rsp.Msg != "Hello John" {
		t.Fatalf("unexpected response %s", rsp.Msg)
	}
}
