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

package manager

import (
	"testing"

	"github.com/lack-io/vine/service/runtime"
	"github.com/lack-io/vine/service/store/memory"
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

func (r *testRuntime) Create(srv *runtime.Service, opts ...runtime.CreateOption) error {
	r.createCount++
	if r.events != nil {
		r.events <- srv
	}
	return nil
}
func (r *testRuntime) Update(srv *runtime.Service, opts ...runtime.UpdateOption) error {
	r.updateCount++
	if r.events != nil {
		r.events <- srv
	}
	return nil
}
func (r *testRuntime) Delete(srv *runtime.Service, opts ...runtime.DeleteOption) error {
	r.deleteCount++
	if r.events != nil {
		r.events <- srv
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
	for _, srv := range testServices {
		s, ok := statuses[srv.Name+":"+srv.Version]
		if !ok {
			t.Errorf("Missing status for %v:%v", srv.Name, srv.Version)
			continue
		}
		if s.Status != srv.Metadata["status"] {
			t.Errorf("Incorrect status for %v:%v, expepcted %v but got %v", srv.Name, srv.Version, srv.Metadata["status"], s.Status)
		}
		if s.Error != srv.Metadata["error"] {
			t.Errorf("Incorrect error for %v:%v, expepcted %v but got %v", srv.Name, srv.Version, srv.Metadata["error"], s.Error)
		}
	}

	// update the status for a service and check it correctly updated
	srv := testServices[0]
	srv.Metadata["status"] = "running"
	if err := m.cacheStatus(namespace.DefaultNamespace, srv); err != nil {
		t.Fatalf("Unexpected error when caching status: %v", err)
	}

	// get the statuses from the service
	statuses, err = m.listStatuses(namespace.DefaultNamespace)
	if err != nil {
		t.Fatalf("Unexpected error when listing statuses: %v", err)
	}

	// check the new status matches the changed service
	s, ok := statuses[srv.Name+":"+srv.Version]
	if !ok {
		t.Errorf("Missing status for %v:%v", srv.Name, srv.Version)
	}
	if s.Status != srv.Metadata["status"] {
		t.Errorf("Incorrect status for %v:%v, expepcted %v but got %v", srv.Name, srv.Version, srv.Metadata["status"], s.Status)
	}
	if s.Error != srv.Metadata["error"] {
		t.Errorf("Incorrect error for %v:%v, expepcted %v but got %v", srv.Name, srv.Version, srv.Metadata["error"], s.Error)
	}
}
