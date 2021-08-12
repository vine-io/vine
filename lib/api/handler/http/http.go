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

// Package http is a http reverse proxy handler
package http

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/lib/api/handler"
	apipb "github.com/vine-io/vine/proto/apis/api"
	ctx "github.com/vine-io/vine/util/context"
)

const (
	Handler = "http"
)

type httpHandler struct {
	options handler.Options

	// set with different initialiser
	s *apipb.Service
}

func (h *httpHandler) Handle(c *fiber.Ctx) error {
	service, err := h.getService(c)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	if len(service) == 0 {
		return fiber.NewError(404)
	}

	rp, err := url.Parse(service)
	if err != nil {
		return fiber.NewError(500)
	}

	//httputil.NewSingleHostReverseProxy(rp).ServeHTTP(w, r)
	return c.Redirect(rp.String())
}

// getService returns the service for this request from the selector
func (h *httpHandler) getService(c *fiber.Ctx) (string, error) {
	var service *apipb.Service

	r := ctx.NewRequestCtx(c, ctx.FromRequest(c))
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
func WithService(s *apipb.Service, opts ...handler.Option) handler.Handler {
	options := handler.NewOptions(opts...)

	return &httpHandler{
		options: options,
		s:       s,
	}
}
