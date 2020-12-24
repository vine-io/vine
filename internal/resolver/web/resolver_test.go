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

package web

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/lack-io/vine/internal/api/resolver"
	"github.com/lack-io/vine/service/client/selector"
	dnsSelector "github.com/lack-io/vine/service/client/selector/dns"
	"github.com/lack-io/vine/service/registry"
	"github.com/lack-io/vine/service/registry/memory"
)

func TestWebResolver(t *testing.T) {
	r := memory.NewRegistry()

	selector := dnsSelector.NewSelector(
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

			u, err := url.Parse("https://" + service.Host + service.Path)
			if err != nil {
				t.Fatal(err)
			}

			req := &http.Request{
				Header: make(http.Header),
				URL:    u,
				Host:   u.Hostname(),
			}
			if endpoint, err := res.Resolve(req); err != nil {
				t.Fatalf("Failed to resolve %v: %v", service, err)
			} else if endpoint.Host != "127.0.0.1:8080" {
				t.Fatalf("Failed to resolve %v", service.Host)
			}

			r.Deregister(v)
		})
	}

}
