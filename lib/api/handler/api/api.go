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

	"github.com/gin-gonic/gin"
	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/lib/api"
	"github.com/vine-io/vine/lib/api/handler"
	"github.com/vine-io/vine/lib/errors"
	ctx "github.com/vine-io/vine/util/context"
	"github.com/vine-io/vine/util/context/metadata"
)

const (
	Handler = "api"
)

type apiHandler struct {
	opts handler.Options
	s    *api.Service
}

// Handle API handler is the default handler which takes api.Request and returns api.Response
func (a *apiHandler) Handle(c *gin.Context) {
	bsize := handler.DefaultMaxRecvSize
	if a.opts.MaxRecvSize > 0 {
		bsize = a.opts.MaxRecvSize
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, bsize)

	request, err := requestToProto(c.Request)
	if err != nil {
		er := errors.InternalServerError("go.vine.client", err.Error())
		c.JSON(200, er.Error())
		return
	}

	var service *api.Service

	// create the context from headers
	cx := ctx.FromRequest(c.Request)
	for k, v := range a.opts.Metadata {
		cx = metadata.Set(cx, k, v)
	}

	r := c.Request.Clone(cx)
	if a.s != nil {
		// we were given the service
		service = a.s
	} else if a.opts.Router != nil {
		// try get service from router
		s, err := a.opts.Router.Route(r)
		if err != nil {
			er := errors.InternalServerError("go.vine.client", err.Error())
			c.JSON(500, er)
			return
		}
		service = s
	} else {
		// we have no way of routing the request
		er := errors.InternalServerError("go.vine.client", "no route found")
		c.JSON(500, er)
		return
	}

	// create request and response
	cc := a.opts.Client
	req := cc.NewRequest(service.Name, service.Endpoint.Name, request)
	rsp := &api.Response{}

	// create strategy
	so := selector.WithStrategy(strategy(service.Services))

	if err := cc.Call(cx, req, rsp, client.WithSelectOption(so)); err != nil {
		ce := errors.Parse(err.Error())
		if ce.Code == 0 {
			c.JSON(500, ce)
			return
		}
		c.JSON(500, ce)
		return
	} else if rsp.StatusCode == 0 {
		rsp.StatusCode = http.StatusOK
	}

	for _, header := range rsp.Header {
		for _, val := range header.Values {
			c.Request.Header.Add(header.Key, val)
		}
	}

	if len(c.GetHeader("Content-Type")) == 0 {
		c.Request.Header.Set("Content-Type", "application/json")
	}

	c.JSON(int(rsp.StatusCode), rsp.Body)
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
