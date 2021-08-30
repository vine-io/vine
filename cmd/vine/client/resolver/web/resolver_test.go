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

package web

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/core/client/selector/dns"
	"github.com/vine-io/vine/core/registry/memory"
	"github.com/vine-io/vine/lib/api/resolver"
	regpb "github.com/vine-io/vine/proto/apis/registry"
)

func TestWebResolver(t *testing.T) {
	r := memory.NewRegistry()

	selector := dns.NewSelector(
		selector.Registry(r),
	)

	res := &Resolver{
		Namespace: resolver.StaticNamespace("go.vine.web"),
		Selector:  selector,
	}

	testCases := []struct {
		Host    string
		Path    string
		Service string
		Type    string
	}{
		{"web.vine.mu", "/home", "go.vine.web.home", "domain"},
		{"localhost:8082", "/foobar", "go.vine.web.foobar", "path"},
		{"web.vine.mu", "/foobar", "go.vine.web.foobar", "path"},
		{"127.0.0.1:8082", "/hello", "go.vine.web.hello", "path"},
		{"account.vine.mu", "/", "go.vine.web.account", "domain"},
		{"foo.m3o.app", "/bar", "foo.web.bar", "domain"},
		{"demo.m3o.app", "/bar", "go.vine.web.bar", "path"},
	}

	for _, service := range testCases {
		t.Run(service.Host+service.Path, func(t *testing.T) {
			// set resolver type
			res.Type = service.Type

			v := &registry.Service{
				Name:    service.Service,
				Version: "latest",
				Nodes: []*registry.Node{
					{Id: "1", Address: "127.0.0.1:8080"},
				},
			}

			r.Register(v)

			//u, err := url.Parse("https://" + service.Host + service.Path)
			//if err != nil {
			//	t.Fatal(err)
			//}

			req := fiber.Ctx{}
			if endpoint, err := res.Resolve(&req); err != nil {
				t.Fatalf("Failed to resolve %v: %v", service, err)
			} else if endpoint.Host != "127.0.0.1:8080" {
				t.Fatalf("Failed to resolve %v", service.Host)
			}

			r.Deregister(v)
		})
	}

}
