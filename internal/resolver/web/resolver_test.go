// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
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
	"time"

	// load the cmd package to load defaults since we're using a test profile without importing
	// vine or service
	_ "github.com/lack-io/vine/cmd"

	"github.com/lack-io/vine/internal/api/resolver"
	"github.com/lack-io/vine/profile"
	"github.com/lack-io/vine/service/registry"
	"github.com/lack-io/vine/service/router"
)

type testCase struct {
	Host    string
	Path    string
	Service string
}

func TestWebResolver(t *testing.T) {
	profile.Test.Setup(nil)

	t.Run("WithServicePrefix", func(t *testing.T) {
		res := &Resolver{
			Options: resolver.NewOptions(
				resolver.WithServicePrefix("web"),
			),
			Router: router.DefaultRouter,
		}

		testCases := []testCase{
			{"localhost:8082", "/foobar", "web.foobar"},
			{"web.vine.mu", "/foobar", "web.foobar"},
			{"127.0.0.1:8082", "/hello", "web.hello"},
			{"demo.m3o.app", "/bar", "web.bar"},
		}

		runTests(t, res, testCases)
	})

	t.Run("WithoutServicePrefix", func(t *testing.T) {
		res := &Resolver{
			Options: resolver.NewOptions(),
			Router:  router.DefaultRouter,
		}

		testCases := []testCase{
			{"localhost:8082", "/foobar", "foobar"},
			{"web.vine.mu", "/foobar", "foobar"},
			{"127.0.0.1:8082", "/hello", "hello"},
			{"demo.m3o.app", "/bar", "bar"},
		}

		runTests(t, res, testCases)
	})
}

func runTests(t *testing.T, res *Resolver, testCases []testCase) {
	for _, service := range testCases {
		t.Run(service.Host+service.Path, func(t *testing.T) {
			v := &registry.Service{
				Name:    service.Service,
				Version: "latest",
				Nodes: []*registry.Node{
					{Id: "1", Address: "127.0.0.1:8080"},
				},
			}

			registry.DefaultRegistry.Register(v)

			// registry events are published to the router async (although if we don't wait the fallback should still kick in)
			time.Sleep(time.Millisecond * 10)

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

			registry.DefaultRegistry.Deregister(v)
		})
	}
}
