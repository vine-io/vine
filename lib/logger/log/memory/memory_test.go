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
	"reflect"
	"testing"

	log2 "github.com/vine-io/vine/lib/logger/log"
)

func TestLogger(t *testing.T) {
	// set size to some value
	size := 100
	// override the global logger
	lg := NewLog(log2.Size(size))
	// make sure we have the right size of the logger ring buffer
	if lg.(*memoryLog).Size() != size {
		t.Errorf("expected buffer size: %d, got: %d", size, lg.(*memoryLog).Size())
	}

	// Log some cruft
	lg.Write(log2.Record{Message: "foobar"})
	lg.Write(log2.Record{Message: "foo bar"})

	// Check if the logs are stored in the logger ring buffer
	expected := []string{"foobar", "foo bar"}
	entries, _ := lg.Read(log2.Count(len(expected)))
	for i, entry := range entries {
		if !reflect.DeepEqual(entry.Message, expected[i]) {
			t.Errorf("expected %s, got %s", expected[i], entry.Message)
		}
	}
}
