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
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/lib/api/handler"
	apipb "github.com/vine-io/vine/proto/apis/api"
	ctx "github.com/vine-io/vine/util/context"
)

const (
	Handler = "web"
)

type webHandler struct {
	opts handler.Options
	s    *apipb.Service
}

func (wh *webHandler) Handle(c *fiber.Ctx) error {
	service, err := wh.getService(c)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	if len(service) == 0 {
		return fiber.NewError(400)
	}

	rp, err := url.Parse(service)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	if isWebSocket(c) {
		return wh.serveWebSocket(rp.Host, c)
	}

	//c.Redirect()
	//return httputil.NewSingleHostReverseProxy(rp).Handle(c)
	return c.Redirect(rp.String())
}

// getService returns the service for this request from the selector
func (wh *webHandler) getService(c *fiber.Ctx) (string, error) {
	var service *apipb.Service

	r := ctx.NewRequestCtx(c, ctx.FromRequest(c))
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
func (wh *webHandler) serveWebSocket(host string, c *fiber.Ctx) error {
	req := new(fiber.Request)
	c.Request().CopyTo(req)

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

	c.Response()
	// hijack the connection
	hj, ok := c.Context().Conn().(http.Hijacker)
	if !ok {
		return fiber.NewError(500, "failed to connect")
	}

	nc, _, err := hj.Hijack()
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	defer nc.Close()
	defer conn.Close()

	if err = req.BodyWriteTo(conn); err != nil {
		return fiber.NewError(500, err.Error())
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

func (wh *webHandler) String() string {
	return "web"
}

func NewHandler(opts ...handler.Option) handler.Handler {
	return &webHandler{
		opts: handler.NewOptions(opts...),
	}
}

func WithService(s *apipb.Service, opts ...handler.Option) handler.Handler {
	options := handler.NewOptions(opts...)

	return &webHandler{
		opts: options,
		s:    s,
	}
}
