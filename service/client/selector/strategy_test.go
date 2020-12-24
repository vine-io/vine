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

package selector

import (
	"os"
	"testing"

	"github.com/lack-io/vine/service/registry"
)

func TestStrategies(t *testing.T) {
	testData := []*registry.Service{
		{
			Name:    "test1",
			Version: "latest",
			Nodes: []*registry.Node{
				{
					Id:      "test1-1",
					Address: "10.0.0.1:1001",
				},
				{
					Id:      "test1-2",
					Address: "10.0.0.2:1002",
				},
			},
		},
		{
			Name:    "test1",
			Version: "default",
			Nodes: []*registry.Node{
				{
					Id:      "test1-3",
					Address: "10.0.0.3:1003",
				},
				{
					Id:      "test1-4",
					Address: "10.0.0.4:1004",
				},
			},
		},
	}

	for name, strategy := range map[string]Strategy{"random": Random, "roundrobin": RoundRobin} {
		next := strategy(testData)
		counts := make(map[string]int)

		for i := 0; i < 100; i++ {
			node, err := next()
			if err != nil {
				t.Fatal(err)
			}
			counts[node.Id]++
		}

		if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
			t.Logf("%s: %+v\n", name, counts)
		}
	}
}
