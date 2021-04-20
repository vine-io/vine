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

type testRuntime struct {
	createCount  int
	readCount    int
	updateCount  int
	deleteCount  int
	readServices []*runtime.Service
	events       chan *runtime.Service
	runtime.Runtime
}

func (r *testRuntime) Reset() {
	r.createCount = 0
	r.readCount = 0
	r.updateCount = 0
	r.deleteCount = 0
}

func (r *testRuntime) Create(svc *runtime.Service, opts ...runtime.CreateOption) error {
	r.createCount++
	if r.events != nil {
		r.events <- svc
	}
	return nil
}
func (r *testRuntime) Update(svc *runtime.Service, opts ...runtime.UpdateOption) error {
	r.updateCount++
	if r.events != nil {
		r.events <- svc
	}
	return nil
}
func (r *testRuntime) Delete(svc *runtime.Service, opts ...runtime.DeleteOption) error {
	r.deleteCount++
	if r.events != nil {
		r.events <- svc
	}
	return nil
}

func (r *testRuntime) Read(opts ...runtime.ReadOption) ([]*runtime.Service, error) {
	r.readCount++
	return r.readServices, nil
}

func TestStatus(t *testing.T) {
	testServices := []*runtime.Service{
		&runtime.Service{
			Name:     "go.vine.service.foo",
			Version:  "latest",
			Metadata: map[string]string{"status": "starting"},
		},
		&runtime.Service{
			Name:     "go.vine.service.bar",
			Version:  "2.0.0",
			Metadata: map[string]string{"status": "error", "error": "Crashed on L1"},
		},
	}

	rt := &testRuntime{readServices: testServices}
	m := New(rt, Store(memory.NewStore())).(*manager)

	// sync the status with the runtime, this should set the status for the testServices in the cache
	m.syncStatus()

	// get the statuses from the service
	statuses, err := m.listStatuses(namespace.DefaultNamespace)
	if err != nil {
		t.Fatalf("Unexpected error when listing statuses: %v", err)
	}

	// loop through the test services and check the status matches what was set in the metadata
	for _, svc := range testServices {
		s, ok := statuses[svc.Name+":"+svc.Version]
		if !ok {
			t.Errorf("Missing status for %v:%v", svc.Name, svc.Version)
			continue
		}
		if s.Status != svc.Metadata["status"] {
			t.Errorf("Incorrect status for %v:%v, expepcted %v but got %v", svc.Name, svc.Version, svc.Metadata["status"], s.Status)
		}
		if s.Error != svc.Metadata["error"] {
			t.Errorf("Incorrect error for %v:%v, expepcted %v but got %v", svc.Name, svc.Version, svc.Metadata["error"], s.Error)
		}
	}

	// update the status for a service and check it correctly updated
	svc := testServices[0]
	svc.Metadata["status"] = "running"
	if err := m.cacheStatus(namespace.DefaultNamespace, svc); err != nil {
		t.Fatalf("Unexpected error when caching status: %v", err)
	}

	// get the statuses from the service
	statuses, err = m.listStatuses(namespace.DefaultNamespace)
	if err != nil {
		t.Fatalf("Unexpected error when listing statuses: %v", err)
	}

	// check the new status matches the changed service
	s, ok := statuses[svc.Name+":"+svc.Version]
	if !ok {
		t.Errorf("Missing status for %v:%v", svc.Name, svc.Version)
	}
	if s.Status != svc.Metadata["status"] {
		t.Errorf("Incorrect status for %v:%v, expepcted %v but got %v", svc.Name, svc.Version, svc.Metadata["status"], s.Status)
	}
	if s.Error != svc.Metadata["error"] {
		t.Errorf("Incorrect error for %v:%v, expepcted %v but got %v", svc.Name, svc.Version, svc.Metadata["error"], s.Error)
	}
}
