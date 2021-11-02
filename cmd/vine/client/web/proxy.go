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

	"github.com/gin-gonic/gin"
)

type proxy struct {
	// The default http reverse proxy
	//Router *httputil.ReverseProxy
	// The director which picks the route
	Director func(ctx *gin.Context)
}

func (p *proxy) Handler(ctx *gin.Context) {
	//if !isWebSocket(c) {
	//	// the usual path
	//	p.Router.ServeHTTP(w, r)
	//	return
	//}

	// the websocket path
	p.Director(ctx)
	host := ctx.Request.Host

	if len(host) == 0 {
		ctx.JSON(500, "invalid host")
		return
	}

	// set x-forward-for
	if clientIP, _, err := net.SplitHostPort(ctx.ClientIP()); err == nil {
		if ips := ctx.GetHeader("X-Forwarded-For"); ips != "" {
			clientIP = ips + ", " + clientIP
		}
		ctx.Request.Header.Set("X-Forwarded-For", clientIP)
	}

	// connect to the backend host
	conn, err := net.Dial("tcp", host)
	if err != nil {
		ctx.JSON(500, err)
		return
	}

	// hijack the connection
	hj, ok := ctx.Request.Body.(http.Hijacker)
	if !ok {
		ctx.JSON(500, "failed to connect")
		return
	}

	nc, _, err := hj.Hijack()
	if err != nil {
		ctx.JSON(500, err)
		return
	}

	defer nc.Close()
	defer conn.Close()

	if err = ctx.Request.Write(conn); err != nil {
		ctx.JSON(500, err)
		return
	}

	errCh := make(chan error, 2)

	cp := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errCh <- err
	}

	go cp(conn, nc)
	go cp(nc, conn)

	<-errCh
}

func isWebSocket(c *gin.Context) bool {
	contains := func(key, val string) bool {
		vv := strings.Split(c.GetHeader(key), ",")
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
