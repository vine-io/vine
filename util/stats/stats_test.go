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
