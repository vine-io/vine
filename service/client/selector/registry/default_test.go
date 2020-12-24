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

package registry

import (
	"os"
	"testing"

	"github.com/lack-io/vine/service/client/selector"
	"github.com/lack-io/vine/service/registry/memory"
)

func TestRegistrySelector(t *testing.T) {
	counts := map[string]int{}

	r := memory.NewRegistry(memory.Services(testData))
	cache := NewSelector(selector.Registry(r))

	next, err := cache.Select("foo")
	if err != nil {
		t.Errorf("Unexpected error calling cache select: %v", err)
	}

	for i := 0; i < 100; i++ {
		node, err := next()
		if err != nil {
			t.Errorf("Expected node err, got err: %v", err)
		}
		counts[node.Id]++
	}

	if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
		t.Logf("Selector Counts %v", counts)
	}
}
