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

package http

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	proxyAuthHeader = "Proxy-Authorization"
)

func getURL(addr string) (*url.URL, error) {
	r := &http.Request{
		URL: &url.URL{
			Scheme: "https",
			Host:   addr,
		},
	}
	return http.ProxyFromEnvironment(r)
}

type pbuffer struct {
	net.Conn
	r io.Reader
}

func (p *pbuffer) Read(b []byte) (int, error) {
	return p.r.Read(b)
}

func proxyDial(conn net.Conn, addr string, proxyURL *url.URL) (_ net.Conn, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	r := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Host: addr},
		Header: map[string][]string{"User-Agent": {"vine/latest"}},
	}

	if user := proxyURL.User; user != nil {
		u := user.Username()
		p, _ := user.Password()
		auth := []byte(u + ":" + p)
		basicAuth := base64.StdEncoding.EncodeToString(auth)
		r.Header.Add(proxyAuthHeader, "Basic "+basicAuth)
	}

	if err := r.Write(conn); err != nil {
		return nil, fmt.Errorf("failed to write the HTTP request: %v", err)
	}

	br := bufio.NewReader(conn)
	rsp, err := http.ReadResponse(br, r)
	if err != nil {
		return nil, fmt.Errorf("reading server HTTP resposne: %v", err)
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		dump, err := httputil.DumpResponse(rsp, true)
		if err != nil {
			return nil, fmt.Errorf("failed to do connect handshake, status code: %s", rsp.Status)
		}
		return nil, fmt.Errorf("failed to do connect handshake, response: %v", dump)
	}

	return &pbuffer{Conn: conn, r: br}, nil
}

// Create a new connection
func newConn(dial func(string) (net.Conn, error)) func(string) (net.Conn, error) {
	return func(addr string) (net.Conn, error) {
		// get the proxy url
		proxyURL, err := getURL(addr)
		if err != nil {
			return nil, err
		}

		// set to addr
		callAddr := addr

		// got proxy
		if proxyURL != nil {
			callAddr = proxyURL.Host
		}

		// dial the addr
		c, err := dial(callAddr)
		if err != nil {
			return nil, err
		}

		// do proxy connect if we have proxy url
		if proxyURL != nil {
			c, err = proxyDial(c, addr, proxyURL)
		}

		return c, err
	}
}
