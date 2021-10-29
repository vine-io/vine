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

// Package web contains the web handler including websocket support
package web

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/lib/api"
	"github.com/vine-io/vine/lib/api/handler"
	ctx "github.com/vine-io/vine/util/context"
)

const (
	Handler = "web"
)

type webHandler struct {
	opts handler.Options
	s    *api.Service
}

func (wh *webHandler) Handle(ctx *gin.Context) {
	service, err := wh.getService(ctx)
	if err != nil {
		ctx.JSON(500, err)
		return
	}

	if len(service) == 0 {
		ctx.JSON(500, err)
		return
	}

	rp, err := url.Parse(service)
	if err != nil {
		ctx.JSON(500, err)
		return
	}

	if isWebSocket(ctx) {
		wh.serveWebSocket(rp.Host, ctx)
		return
	}

	httputil.NewSingleHostReverseProxy(rp).ServeHTTP(ctx.Writer, ctx.Request)
}

// getService returns the service for this request from the selector
func (wh *webHandler) getService(c *gin.Context) (string, error) {
	var service *api.Service

	r := c.Request.Clone(ctx.FromRequest(c.Request))
	if wh.s != nil {
		// we were given the service
		service = wh.s
	} else if wh.opts.Router != nil {
		// try get service from router
		s, err := wh.opts.Router.Route(r)
		if err != nil {
			return "", err
		}
		service = s
	} else {
		// we have no way of routing the request
		return "", errors.New("no route found")
	}

	// create a random selector
	next := selector.Random(service.Services)

	// get the next node
	s, err := next()
	if err != nil {
		return "", nil
	}

	return fmt.Sprintf("http://%s", s.Address), nil
}

// serveWebSocket used to serve a web socket proxied connection
func (wh *webHandler) serveWebSocket(host string, c *gin.Context) {
	ctx := c.Copy()

	if len(host) == 0 {
		ctx.JSON(500, "invalid host")
		return
	}

	// set x-forward-for
	if clientIP, _, err := net.SplitHostPort(c.ClientIP()); err == nil {
		if ips := ctx.GetHeader("X-Forwarded-For"); ips != "" {
			clientIP = ips + ", " + clientIP
		}
		c.Header("X-Forwarded-For", clientIP)
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
		ctx.JSON(500, err.Error())
		return
	}

	defer nc.Close()
	defer conn.Close()

	if err = ctx.Request.Write(conn); err != nil {
		ctx.JSON(500, err.Error())
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

func isWebSocket(ctx *gin.Context) bool {
	contains := func(key, val string) bool {
		vv := strings.Split(ctx.GetHeader(key), ",")
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

func (wh *webHandler) String() string {
	return "web"
}

func NewHandler(opts ...handler.Option) handler.Handler {
	return &webHandler{
		opts: handler.NewOptions(opts...),
	}
}

func WithService(s *api.Service, opts ...handler.Option) handler.Handler {
	options := handler.NewOptions(opts...)

	return &webHandler{
		opts: options,
		s:    s,
	}
}
