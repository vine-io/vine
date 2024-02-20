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

package net

import (
	"net"
	"os"
	"testing"
)

func TestListen(t *testing.T) {
	fn := func(addr string) (net.Listener, error) {
		return net.Listen("tcp", addr)
	}

	// try to create a number of listeners
	for i := 0; i < 10; i++ {
		l, err := Listen("localhost:10000-11000", fn)
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()
	}

	// TODO nats case test
}

// TestProxyEnv checks whether we have proxy/network settings in env
func TestProxyEnv(t *testing.T) {
	service := "foo"
	address := []string{"bar"}

	s, a, ok := Proxy(service, address)
	if ok {
		t.Fatal("Should not have proxy", s, a, ok)
	}

	test := func(key, val, expectSrv, expectAddr string) {
		// set env
		os.Setenv(key, val)

		s, a, ok := Proxy(service, address)
		if !ok {
			t.Fatal("Expected proxy")
		}
		if len(expectSrv) > 0 && s != expectSrv {
			t.Fatal("Expected proxy service", expectSrv, "got", s)
		}
		if len(expectAddr) > 0 {
			if len(a) == 0 || a[0] != expectAddr {
				t.Fatal("Expected proxy address", expectAddr, "got", a)
			}
		}

		os.Unsetenv(key)
	}

	test("VINE_PROXY", "service", "go.vine.proxy", "")
	test("VINE_NETWORK", "service", "go.vine.network", "")
	test("VINE_NETWORK_ADDRESS", "10.0.0.1:8081", "", "10.0.0.1:8081")
}
