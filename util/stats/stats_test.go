// Copyright 2020 lack
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

package stats

import (
	"testing"
)

func TestStats(t *testing.T) {
	testCounters := []struct {
		c string
		i []int
	}{
		{
			c: "test",
			i: []int{1, 10, 100},
		},
	}

	s := New()

	if err := s.Start(); err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCounters {
		for _, i := range tc.i {
			s.Record(tc.c, i)
		}
	}

	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}

	if len(s.Counters) == 0 {
		t.Fatalf("stats not recorded, counters are %+v", s.Counters)
	}

	for _, tc := range testCounters {
		if _, ok := s.Counters[0].Status[tc.c]; !ok {
			t.Fatalf("%s counter not found", tc.c)
		}
	}
}
