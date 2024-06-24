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

// Package http is a http reverse proxy handler
package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/lib/api"
	"github.com/vine-io/vine/lib/api/handler"
	ctx "github.com/vine-io/vine/util/context"
)

const (
	Handler = "http"
)

type httpHandler struct {
	options handler.Options

	// set with different initialiser
	s *api.Service
}

func (h *httpHandler) Handle(c *gin.Context) {
	service, err := h.getService(c)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	if len(service) == 0 {
		c.JSON(404, "")
		return
	}

	rp, err := url.Parse(service)
	if err != nil {
		c.JSON(500, "")
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(rp)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = rp.Host
		req.URL.Scheme = rp.Scheme
		req.URL.Host = rp.Host
		req.URL.Path = c.Request.URL.Path
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// getService returns the service for this request from the selector
func (h *httpHandler) getService(c *gin.Context) (string, error) {
	var service *api.Service

	r := c.Request.Clone(ctx.FromRequest(c.Request))
	if h.s != nil {
		// we were given the service
		service = h.s
	} else if h.options.Router != nil {
		// try get service from router
		s, err := h.options.Router.Route(r)
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

func (h *httpHandler) String() string {
	return "http"
}

// NewHandler returns a http proxy handler
func NewHandler(opts ...handler.Option) handler.Handler {
	options := handler.NewOptions(opts...)

	return &httpHandler{
		options: options,
	}
}

// WithService creates a handler with a service
func WithService(s *api.Service, opts ...handler.Option) handler.Handler {
	options := handler.NewOptions(opts...)

	return &httpHandler{
		options: options,
		s:       s,
	}
}
