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
	"time"

	"github.com/lack-io/vine/internal/namespace"
	"github.com/lack-io/vine/service/runtime"
	"github.com/lack-io/vine/service/store/memory"
)

func TestEvents(t *testing.T) {
	// an event is passed through this channel from the test runtime everytime a method is called,
	// this is done since events ae processed async
	eventChan := make(chan *runtime.Service)

	rt := &testRuntime{events: eventChan}
	m := New(rt, Store(memory.NewStore())).(*manager)

	// set the eventPollFrequency to 10ms so events are processed immediately
	eventPollFrequency = time.Millisecond * 10
	go m.watchEvents()

	// timeout async tests after 500ms
	timeout := time.NewTimer(time.Millisecond * 500)

	// the service that should be passed to the runtime
	testSrv := &runtime.Service{Name: "go.vine.service.foo", Version: "latest"}
	opts := &runtime.CreateOptions{Namespace: namespace.DefaultNamespace}

	t.Run("Create", func(t *testing.T) {
		defer rt.Reset()

		if err := m.publishEvent(runtime.Create, testSrv, opts); err != nil {
			t.Errorf("Unexpected error when publishing events: %v", err)
		}

		timeout.Reset(time.Millisecond * 500)

		select {
		case srv := <-eventChan:
			if srv.Name != testSrv.Name || srv.Version != testSrv.Version {
				t.Errorf("Incorrect service passed to the runtime")
			}
		case <-timeout.C:
			t.Fatalf("The runtime wasn't called")
		}

		if rt.createCount != 1 {
			t.Errorf("Expected runtime create to be called 1 time but was actually called %v times", rt.createCount)
		}
	})

	t.Run("Update", func(t *testing.T) {
		defer rt.Reset()

		if err := m.publishEvent(runtime.Update, testSrv, opts); err != nil {
			t.Errorf("Unexpected error when publishing events: %v", err)
		}

		timeout.Reset(time.Millisecond * 500)

		select {
		case srv := <-eventChan:
			if srv.Name != testSrv.Name || srv.Version != testSrv.Version {
				t.Errorf("Incorrect service passed to the runtime")
			}
		case <-timeout.C:
			t.Fatalf("The runtime wasn't called")
		}

		if rt.updateCount != 1 {
			t.Errorf("Expected runtime update to be called 1 time but was actually called %v times", rt.updateCount)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		defer rt.Reset()

		if err := m.publishEvent(runtime.Delete, testSrv, opts); err != nil {
			t.Errorf("Unexpected error when publishing events: %v", err)
		}

		timeout.Reset(time.Millisecond * 500)

		select {
		case srv := <-eventChan:
			if srv.Name != testSrv.Name || srv.Version != testSrv.Version {
				t.Errorf("Incorrect service passed to the runtime")
			}
		case <-timeout.C:
			t.Fatalf("The runtime wasn't called")
		}

		if rt.deleteCount != 1 {
			t.Errorf("Expected runtime delete to be called 1 time but was actually called %v times", rt.deleteCount)
		}
	})
}
