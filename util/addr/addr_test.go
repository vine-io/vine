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

package addr

import (
	"net"
	"testing"
)

func TestIsLocal(t *testing.T) {
	testData := []struct {
		addr   string
		expect bool
	}{
		{"localhost", true},
		{"localhost:8080", true},
		{"127.0.0.1", true},
		{"127.0.0.1:1001", true},
		{"80.1.1.1", false},
	}

	for _, d := range testData {
		res := IsLocal(d.addr)
		if res != d.expect {
			t.Fatalf("expected %t got %t", d.expect, res)
		}
	}
}

func TestExtractor(t *testing.T) {
	testData := []struct {
		addr   string
		expect string
		parse  bool
	}{
		{"127.0.0.1", "127.0.0.1", false},
		{"10.0.0.1", "10.0.0.1", false},
		{"", "", true},
		{"0.0.0.0", "", true},
		{"[::]", "", true},
	}

	for _, d := range testData {
		addr, err := Extract(d.addr)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		if d.parse {
			ip := net.ParseIP(addr)
			if ip == nil {
				t.Errorf("Unexpected nil IP")
			}

		} else if addr != d.expect {
			t.Errorf("Expected %v got %v", d.expect, addr)
		}
	}
}

func TestAppendPrivateBlocks(t *testing.T) {
	tests := []struct {
		addr   string
		expect bool
	}{
		{"9.134.71.34", true},
		{"8.222.33.110", false}, // not in private blocks
	}

	AppendPrivateBlocks("9.134.0.0/16")

	for _, test := range tests {
		t.Run(test.addr, func(t *testing.T) {
			res := isPrivateIP(test.addr)
			if res != test.expect {
				t.Fatalf("expected %v got %v", test.expect, res)
			}
		})
	}
}
