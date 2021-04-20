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

package manager

import (
	"testing"

	"github.com/lack-io/vine/lib/runtime"
	"github.com/lack-io/vine/lib/store/memory"
	"github.com/lack-io/vine/util/namespace"
)

func TestServices(t *testing.T) {
	testServices := []*runtime.Service{
		&runtime.Service{Name: "go.vine.service.foo", Version: "2.0.0"},
		&runtime.Service{Name: "go.vine.service.foo", Version: "1.0.0"},
		&runtime.Service{Name: "go.vine.service.bar", Version: "latest"},
	}

	testNamespace := "foo"

	m := New(&testRuntime{}, Store(memory.NewStore())).(*manager)

	// listNamespaces shoud always return the default namespace
	t.Run("DefaultNamespace", func(t *testing.T) {
		if ns, err := m.listNamespaces(); err != nil {
			t.Errorf("Unexpected error when listing service: %v", err)
		} else if len(ns) != 1 {
			t.Errorf("Expected one namespace, acutually got %v", len(ns))
		} else if ns[0] != namespace.DefaultNamespace {
			t.Errorf("Expected the default namespace to be %v but was got %v", namespace.DefaultNamespace, ns[0])
		}
	})

	// creating a service should not error
	t.Run("CreateService", func(t *testing.T) {
		for _, svc := range testServices {
			if err := m.createService(svc, &runtime.CreateOptions{Namespace: testNamespace}); err != nil {
				t.Fatalf("Unexpected error when creating service %v:%v: %v", svc.Name, svc.Version, err)
			}
		}
	})

	// Calling readServices with a blank service should return all the services in the namespace
	t.Run("ReadServices", func(t *testing.T) {
		svcs, err := m.readServices(testNamespace, &runtime.Service{})
		if err != nil {
			t.Fatalf("Unexpected error when reading services%v", err)
		}
		if len(svcs) != 3 {
			t.Errorf("Expected 3 services, got %v", len(svcs))
		}
	})

	// Calling readServices with a name should return any service with that name
	t.Run("ReadServicesWithName", func(t *testing.T) {
		svcs, err := m.readServices(testNamespace, &runtime.Service{Name: "go.vine.service.foo"})
		if err != nil {
			t.Fatalf("Unexpected error when reading services%v", err)
		}
		if len(svcs) != 2 {
			t.Errorf("Expected 2 services, got %v", len(svcs))
		}
	})

	// Calling readServices with a name and version should only return the services with that name
	// and version
	t.Run("ReadServicesWithNameAndVersion", func(t *testing.T) {
		query := &runtime.Service{Name: "go.vine.service.foo", Version: "1.0.0"}
		svcs, err := m.readServices(testNamespace, query)
		if err != nil {
			t.Fatalf("Unexpected error when reading services%v", err)
		}
		if len(svcs) != 1 {
			t.Errorf("Expected 1 service, got %v", len(svcs))
		}
	})

	// Calling delete service should remove the service with that name and version
	t.Run("DeleteService", func(t *testing.T) {
		query := &runtime.Service{Name: "go.vine.service.foo", Version: "1.0.0"}
		if err := m.deleteService(testNamespace, query); err != nil {
			t.Fatalf("Unexpected error when reading services%v", err)
		}

		svcs, err := m.readServices(testNamespace, &runtime.Service{})
		if err != nil {
			t.Fatalf("Unexpected error when reading services%v", err)
		}
		if len(svcs) != 2 {
			t.Errorf("Expected 2 services, got %v", len(svcs))
		}
	})

	// a service created in one namespace shouldn't be returned when querying another
	t.Run("NamespaceScope", func(t *testing.T) {
		svc := &runtime.Service{Name: "go.vine.service.apple", Version: "latest"}

		if err := m.createService(svc, &runtime.CreateOptions{Namespace: "random"}); err != nil {
			t.Fatalf("Unexpected error when creating service %v", err)
		}

		if svcs, err := m.readServices(testNamespace, svc); err != nil {
			t.Fatalf("Unexpected error when listing services %v", err)
		} else if len(svcs) != 0 {
			t.Errorf("Expected 0 services, got %v", len(svcs))
		}
	})
}
