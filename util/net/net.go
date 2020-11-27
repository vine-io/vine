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

package net

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// HostPort format addr and port suitable for dial
func HostPort(addr string, port interface{}) string {
	host := addr
	if strings.Count(addr, ":") > 0 {
		host = fmt.Sprintf("[%s]", addr)
	}
	// when port is blank or 0, host is q queue name
	if v, ok := port.(string); ok && v == "" {
		return host
	} else if v, ok := port.(int); ok && v == 0 && net.ParseIP(host) == nil {
		return host
	}

	return fmt.Sprintf("%s:%v", host, port)
}

// Listen takes addr:portmin-portmax and binds to the first avaiable port
// Example: Listen("localhost:5000-6000", fn)
func Listen(addr string, fn func(string) (net.Listener, error)) (net.Listener, error) {

	if strings.Count(addr, ":") == 1 && strings.Count(addr, "-") == 0 {
		fn(addr)
	}

	// host:port || host:min-max
	host, ports, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	// try to extract port range
	prange := strings.Split(ports, "-")

	// single port
	if len(prange) < 2 {
		return fn(addr)
	}

	// we have a port range

	// extract min port
	min, err := strconv.Atoi(prange[0])
	if err != nil {
		return nil, fmt.Errorf("unable to extract port range")
	}

	// extract max port
	max, err := strconv.Atoi(prange[1])
	if err != nil {
		return nil, fmt.Errorf("unable to extract port prange")
	}

	// range the ports
	for port := min; port <= max; port++ {
		// try bind to host:port
		ln, err := fn(HostPort(host, port))
		if err == nil {
			return ln, nil
		}

		// hit max port
		if port == max {
			return nil, err
		}
	}

	// why are we here?
	return nil, fmt.Errorf("unable to bind %v", addr)
}

// Proxy returns the proxy and the address if it exists
func Proxy(service string, address []string) (string, []string, bool) {
	var hasProxy bool

	// get proxy. we parse out address if present
	if prx := os.Getenv("VINE_PROXY"); len(prx) > 0 {
		// default name
		if prx == "service" {
			prx = "go.vine.proxy"
			address = nil
		}

		// check if its an address
		if v := strings.Split(prx, ":"); len(v) > 1 {
			address = []string{prx}
		}

		service = prx
		hasProxy = true

		return service, address, hasProxy
	}

	if prx := os.Getenv("VINE_NETWORK"); len(prx) > 0 {
		// default name
		if prx == "service" {
			prx = "go.vine.network"
		}
		service = prx
		hasProxy = true
	}

	if prx := os.Getenv("VINE_NETWORK_ADDRESS"); len(prx) > 0 {
		address = []string{prx}
		hasProxy = true
	}

	return service, address, hasProxy
}