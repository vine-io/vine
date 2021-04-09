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
		return fn(addr)
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
