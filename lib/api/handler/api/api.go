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

// Package api provides an http-rpc handler which provides the entire http request over rpc
package api

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/lib/api/handler"
	"github.com/vine-io/vine/lib/errors"
	"github.com/vine-io/vine/lib/api"
	ctx "github.com/vine-io/vine/util/context"
)

const (
	Handler = "api"
)

type apiHandler struct {
	opts handler.Options
	s    *api.Service
}

// Handle API handler is the default handler which takes api.Request and returns api.Response
func (a *apiHandler) Handle(c *fiber.Ctx) error {
	bsize := handler.DefaultMaxRecvSize
	if a.opts.MaxRecvSize > 0 {
		bsize = a.opts.MaxRecvSize
	}

	c.Context().SetBodyStream(c.Context().RequestBodyStream(), int(bsize))

	request, err := requestToProto(c)
	if err != nil {
		er := errors.InternalServerError("go.vine.client", err.Error())
		c.Set("Content-Type", "application/json")
		return fiber.NewError(500, er.Error())
	}

	var service *api.Service

	// create the context from headers
	cx := ctx.FromRequest(c)
	r := ctx.NewRequestCtx(c, cx)
	if a.s != nil {
		// we were given the service
		service = a.s
	} else if a.opts.Router != nil {
		// try get service from router
		s, err := a.opts.Router.Route(r)
		if err != nil {
			er := errors.InternalServerError("go.vine.client", err.Error())
			c.Set("Content-Type", "application/json")
			return fiber.NewError(500, er.Error())
		}
		service = s
	} else {
		// we have no way of routing the request
		er := errors.InternalServerError("go.vine.client", "no route found")
		c.Set("Content-Type", "application/json")
		return fiber.NewError(500, er.Error())
	}

	// create request and response
	cc := a.opts.Client
	req := cc.NewRequest(service.Name, service.Endpoint.Name, request)
	rsp := &api.Response{}

	// create strategy
	so := selector.WithStrategy(strategy(service.Services))

	if err := cc.Call(cx, req, rsp, client.WithSelectOption(so)); err != nil {
		c.Set("Content-Type", "application/json")
		ce := errors.Parse(err.Error())
		if ce.Code == 0 {
			return fiber.NewError(500, ce.Error())
		}
		return fiber.NewError(int(ce.Code), ce.Error())
	} else if rsp.StatusCode == 0 {
		rsp.StatusCode = http.StatusOK
	}

	for _, header := range rsp.Header {
		for _, val := range header.Values {
			r.Request().Header.Add(header.Key, val)
		}
	}

	if len(r.Get("Content-Type")) == 0 {
		r.Set("Content-Type", "application/json")
	}

	r.SendStatus(int(rsp.StatusCode))
	r.SendString(rsp.Body)
	return nil
}

func (a *apiHandler) String() string {
	return "api"
}

func NewHandler(opts ...handler.Option) handler.Handler {
	options := handler.NewOptions(opts...)
	return &apiHandler{
		opts: options,
	}
}

func WithService(s *api.Service, opts ...handler.Option) handler.Handler {
	options := handler.NewOptions(opts...)
	return &apiHandler{
		opts: options,
		s:    s,
	}
}
