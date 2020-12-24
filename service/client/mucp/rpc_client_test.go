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

package mucp

import (
	"context"
	"fmt"
	"testing"

	"github.com/lack-io/vine/proto/errors"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/client/selector"
	"github.com/lack-io/vine/service/registry"
	"github.com/lack-io/vine/service/registry/memory"
)

func newTestRegistry() registry.Registry {
	return memory.NewRegistry(memory.Services(testData))
}

func TestCallAddress(t *testing.T) {
	var called bool
	service := "test.service"
	endpoint := "Test.Endpoint"
	address := "10.1.10.1:8080"

	wrap := func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
			called = true

			if req.Service() != service {
				return fmt.Errorf("expected service: %s got %s", service, req.Service())
			}

			if req.Endpoint() != endpoint {
				return fmt.Errorf("expected service: %s got %s", endpoint, req.Endpoint())
			}

			if node.Address != address {
				return fmt.Errorf("expected address: %s got %s", address, node.Address)
			}

			// don't do the call
			return nil
		}
	}

	r := newTestRegistry()
	c := NewClient(
		client.Registry(r),
		client.WrapCall(wrap),
	)
	c.Options().Selector.Init(selector.Registry(r))

	req := c.NewRequest(service, endpoint, nil)

	// test calling remote address
	if err := c.Call(context.Background(), req, nil, client.WithAddress(address)); err != nil {
		t.Fatal("call with address error", err)
	}

	if !called {
		t.Fatal("wrapper not called")
	}

}

func TestCallRetry(t *testing.T) {
	service := "test.service"
	endpoint := "Test.Endpoint"
	address := "10.1.10.1"

	var called int

	wrap := func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
			called++
			if called == 1 {
				return errors.InternalServerError("test.error", "retry request")
			}

			// don't do the call
			return nil
		}
	}

	r := newTestRegistry()
	c := NewClient(
		client.Registry(r),
		client.WrapCall(wrap),
	)
	c.Options().Selector.Init(selector.Registry(r))

	req := c.NewRequest(service, endpoint, nil)

	// test calling remote address
	if err := c.Call(context.Background(), req, nil, client.WithAddress(address)); err != nil {
		t.Fatal("call with address error", err)
	}

	// num calls
	if called < c.Options().CallOptions.Retries+1 {
		t.Fatal("request not retried")
	}
}

func TestCallWrapper(t *testing.T) {
	var called bool
	id := "test.1"
	service := "test.service"
	endpoint := "Test.Endpoint"
	address := "10.1.10.1:8080"

	wrap := func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
			called = true

			if req.Service() != service {
				return fmt.Errorf("expected service: %s got %s", service, req.Service())
			}

			if req.Endpoint() != endpoint {
				return fmt.Errorf("expected service: %s got %s", endpoint, req.Endpoint())
			}

			if node.Address != address {
				return fmt.Errorf("expected address: %s got %s", address, node.Address)
			}

			// don't do the call
			return nil
		}
	}

	r := newTestRegistry()
	c := NewClient(
		client.Registry(r),
		client.WrapCall(wrap),
	)
	c.Options().Selector.Init(selector.Registry(r))

	r.Register(&registry.Service{
		Name:    service,
		Version: "latest",
		Nodes: []*registry.Node{
			{
				Id:      id,
				Address: address,
				Metadata: map[string]string{
					"protocol": "mucp",
				},
			},
		},
	})

	req := c.NewRequest(service, endpoint, nil)
	if err := c.Call(context.Background(), req, nil); err != nil {
		t.Fatal("call wrapper error", err)
	}

	if !called {
		t.Fatal("wrapper not called")
	}
}
