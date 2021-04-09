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
	"time"

	"github.com/lack-io/vine/service/runtime"
	"github.com/lack-io/vine/service/store/memory"
	"github.com/lack-io/vine/util/namespace"
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
		case svc := <-eventChan:
			if svc.Name != testSrv.Name || svc.Version != testSrv.Version {
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
		case svc := <-eventChan:
			if svc.Name != testSrv.Name || svc.Version != testSrv.Version {
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
		case svc := <-eventChan:
			if svc.Name != testSrv.Name || svc.Version != testSrv.Version {
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
