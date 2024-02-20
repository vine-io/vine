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

package memory

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/vine-io/vine/core/registry"
)

var (
	testData = map[string][]*registry.Service{
		"foo": {
			{
				Name:    "foo",
				Version: "1.0.0",
				Nodes: []*registry.Node{
					{
						Id:      "foo-1.0.0-123",
						Address: "localhost:9999",
					},
					{
						Id:      "foo-1.0.0-321",
						Address: "localhost:9999",
					},
				},
			},
			{
				Name:    "foo",
				Version: "1.0.1",
				Nodes: []*registry.Node{
					{
						Id:      "foo-1.0.1-321",
						Address: "localhost:6666",
					},
				},
			},
			{
				Name:    "foo",
				Version: "1.0.3",
				Nodes: []*registry.Node{
					{
						Id:      "foo-1.0.3-345",
						Address: "localhost:8888",
					},
				},
			},
		},
		"bar": {
			{
				Name:    "bar",
				Version: "default",
				Nodes: []*registry.Node{
					{
						Id:      "bar-1.0.0-123",
						Address: "localhost:9999",
					},
					{
						Id:      "bar-1.0.0-321",
						Address: "localhost:9999",
					},
				},
			},
			{
				Name:    "bar",
				Version: "latest",
				Nodes: []*registry.Node{
					{
						Id:      "bar-1.0.1-321",
						Address: "localhost:6666",
					},
				},
			},
		},
	}
)

func TestMemoryRegistry(t *testing.T) {
	m := NewRegistry()

	ctx := context.TODO()
	fn := func(k string, v []*registry.Service) {
		services, err := m.GetService(ctx, k)
		if err != nil {
			t.Errorf("Unexpected error getting service %s: %v", k, err)
		}

		if len(services) != len(v) {
			t.Errorf("Expected %d services for %s, got %d", len(v), k, len(services))
		}

		for _, service := range v {
			var seen bool
			for _, s := range services {
				if s.Version == service.Version {
					seen = true
					break
				}
			}
			if !seen {
				t.Errorf("expected to find version %s", service.Version)
			}
		}
	}

	// register data
	for _, v := range testData {
		serviceCount := 0
		for _, service := range v {
			if err := m.Register(ctx, service); err != nil {
				t.Errorf("Unexpected register error: %v", err)
			}
			serviceCount++
			// after the service has been registered we should be able to query it
			services, err := m.GetService(ctx, service.Name)
			if err != nil {
				t.Errorf("Unexpected error getting service %s: %v", service.Name, err)
			}
			if len(services) != serviceCount {
				t.Errorf("Expected %d services for %s, got %d", serviceCount, service.Name, len(services))
			}
		}
	}

	// using test data
	for k, v := range testData {
		fn(k, v)
	}

	services, err := m.ListServices(ctx)
	if err != nil {
		t.Errorf("Unexpected error when listing services: %v", err)
	}

	totalServiceCount := 0
	for _, testSvc := range testData {
		for range testSvc {
			totalServiceCount++
		}
	}

	if len(services) != totalServiceCount {
		t.Errorf("Expected total service count: %d, got: %d", totalServiceCount, len(services))
	}

	// deregister
	for _, v := range testData {
		for _, service := range v {
			if err := m.Deregister(ctx, service); err != nil {
				t.Errorf("Unexpected deregister error: %v", err)
			}
		}
	}

	// after all the service nodes have been deregistered we should not get any results
	for _, v := range testData {
		for _, service := range v {
			services, err := m.GetService(ctx, service.Name)
			if err != registry.ErrNotFound {
				t.Errorf("Expected error: %v, got: %v", registry.ErrNotFound, err)
			}
			if len(services) != 0 {
				t.Errorf("Expected %d services for %s, got %d", 0, service.Name, len(services))
			}
		}
	}
}

func TestMemoryRegisterTTL(t *testing.T) {
	m := NewRegistry()

	ctx := context.TODO()
	for _, v := range testData {
		for _, service := range v {
			if err := m.Register(ctx, service, registry.RegisterTTL(time.Millisecond)); err != nil {
				t.Fatal(err)
			}
		}
	}

	time.Sleep(ttlPruneTime * 2)

	for name := range testData {
		svcs, err := m.GetService(ctx, name)
		if err != nil {
			t.Fatal(err)
		}

		for _, svc := range svcs {
			if len(svc.Nodes) > 0 {
				t.Fatalf("Service %q still has nodes registered", name)
			}
		}
	}
}

func TestMemoryRegisterTTLConcurrent(t *testing.T) {
	concurrency := 1000
	waitTime := ttlPruneTime * 2
	m := NewRegistry()
	ctx := context.TODO()

	for _, v := range testData {
		for _, service := range v {
			if err := m.Register(ctx, service, registry.RegisterTTL(waitTime/2)); err != nil {
				t.Fatal(err)
			}
		}
	}

	if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
		t.Logf("test will wait %v, then check TTL timeouts", waitTime)
	}

	errChan := make(chan error, concurrency)
	syncChan := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		go func() {
			<-syncChan
			for name := range testData {
				svcs, err := m.GetService(ctx, name)
				if err != nil {
					errChan <- err
					return
				}

				for _, svc := range svcs {
					if len(svc.Nodes) > 0 {
						errChan <- fmt.Errorf("Service %q still has nodes registered", name)
						return
					}
				}
			}

			errChan <- nil
		}()
	}

	time.Sleep(waitTime)
	close(syncChan)

	for i := 0; i < concurrency; i++ {
		if err := <-errChan; err != nil {
			t.Fatal(err)
		}
	}
}
