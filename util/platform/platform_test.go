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

package platform

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	gorun "github.com/lack-io/vine/service/runtime"
)

type mockScheduler struct {
	ch chan gorun.Event
}

func (s *mockScheduler) Notify() (<-chan gorun.Event, error) {
	return s.ch, nil
}

func (s *mockScheduler) Close() error {
	return nil
}

func TestNilOnNotify(t *testing.T) {
	msch := &mockScheduler{ch: make(chan gorun.Event, 32)}
	is := initScheduler{Scheduler: msch, services: Services}

	ch, err := is.Notify()
	assert.NoError(t, err)
	msch.ch <- gorun.Event{
		Type:      gorun.Update,
		Timestamp: time.Now(),
	}
	multiple := len(Services) * 2
	tick := time.NewTicker(time.Duration(multiple) * time.Second)
	defer tick.Stop()
	for i := 0; i < len(Services); i++ {
		select {
		case ev := <-ch:
			assert.NotNil(t, ev.Timestamp, "Event timestamp should not be nil")
		case <-tick.C:
			assert.FailNow(t, "Failed to get enough events")
		}
	}

	msch.ch <- gorun.Event{
		Type:      gorun.Update,
		Timestamp: time.Now(),
		Service: &gorun.Service{
			Name:    "foobar",
			Version: "123",
		},
	}
	multiple = len(Services) * 2
	tick = time.NewTicker(time.Duration(multiple) * time.Second)
	defer tick.Stop()
	for i := 0; i < len(Services); i++ {
		select {
		case ev := <-ch:
			assert.NotNil(t, ev.Timestamp, "Event timestamp should not be nil")
			assert.Equal(t, ev.Service.Version, "123")
		case <-tick.C:
			assert.FailNow(t, "Failed to get enough events")
		}
	}

}
