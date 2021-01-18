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
