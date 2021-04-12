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

package web

import (
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type proxy struct {
	// The default http reverse proxy
	//Router *httputil.ReverseProxy
	// The director which picks the route
	Director func(c *fiber.Ctx)
}

func (p *proxy) Handler(c *fiber.Ctx) error {
	//if !isWebSocket(c) {
	//	// the usual path
	//	p.Router.ServeHTTP(w, r)
	//	return
	//}

	// the websocket path
	p.Director(c)
	host := string(c.Request().Host())

	if len(host) == 0 {
		return fiber.NewError(500, "invalid host")
	}

	// set x-forward-for
	if clientIP, _, err := net.SplitHostPort(c.IP()); err == nil {
		if ips := c.Get("X-Forwarded-For"); ips != "" {
			clientIP = ips + ", " + clientIP
		}
		c.Set("X-Forwarded-For", clientIP)
	}

	// connect to the backend host
	conn, err := net.Dial("tcp", host)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	// hijack the connection
	hj, ok := c.Context().Conn().(http.Hijacker)
	if !ok {
		return fiber.NewError(500, "failed to connect")
	}

	nc, _, err := hj.Hijack()
	if err != nil {
		return err
	}

	defer nc.Close()
	defer conn.Close()

	if _, err = c.Request().WriteTo(conn); err != nil {
		return err
	}

	errCh := make(chan error, 2)

	cp := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errCh <- err
	}

	go cp(conn, nc)
	go cp(nc, conn)

	<-errCh

	return nil
}

func isWebSocket(c *fiber.Ctx) bool {
	contains := func(key, val string) bool {
		vv := strings.Split(c.Get(key), ",")
		for _, v := range vv {
			if val == strings.ToLower(strings.TrimSpace(v)) {
				return true
			}
		}
		return false
	}

	if contains("Connection", "upgrade") && contains("Upgrade", "websocket") {
		return true
	}

	return false
}
