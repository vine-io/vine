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

package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lack-io/vine"
	client2 "github.com/lack-io/vine/core/client"
	"github.com/lack-io/vine/lib/api/handler"
	aapi "github.com/lack-io/vine/lib/api/handler/api"
	"github.com/lack-io/vine/lib/api/handler/event"
	ahttp "github.com/lack-io/vine/lib/api/handler/http"
	arpc "github.com/lack-io/vine/lib/api/handler/rpc"
	aweb "github.com/lack-io/vine/lib/api/handler/web"
	"github.com/lack-io/vine/lib/api/router"
	"github.com/lack-io/vine/proto/apis/errors"
	ctx "github.com/lack-io/vine/util/context"
)

type metaHandler struct {
	c  client2.Client
	r  router.Router
	ns func(*fiber.Ctx) string
}

func (m *metaHandler) Handle(c *fiber.Ctx) error {

	r := ctx.NewRequestCtx(c, ctx.FromRequest(c))
	service, err := m.r.Route(r)
	if err != nil {
		err := errors.InternalServerError(m.ns(c), err.Error())
		c.Set("Content-Type", "application/json")
		return fiber.NewError(500, err.Error())
	}

	// TODO: don't do this ffs
	switch service.Endpoint.Handler {
	// web socket handler
	case aweb.Handler:
		return aweb.WithService(service, handler.WithClient(m.c)).Handle(c)
	// proxy handler
	case "proxy", ahttp.Handler:
		return ahttp.WithService(service, handler.WithClient(m.c)).Handle(c)
	// rpcx handler
	case arpc.Handler:
		return arpc.WithService(service, handler.WithClient(m.c)).Handle(c)
	// event handler
	case event.Handler:
		ev := event.NewHandler(
			handler.WithNamespace(m.ns(c)),
			handler.WithClient(m.c),
		)
		return ev.Handle(c)
	// api handler
	case aapi.Handler:
		return aapi.WithService(service, handler.WithClient(m.c)).Handle(c)
	// default handler: rpc
	default:
		return arpc.WithService(service, handler.WithClient(m.c)).Handle(c)
	}
}

func (m *metaHandler) String() string {
	return "meta"
}

// Meta is a http.Handler that routes based on endpoint metadata
func Meta(s vine.Service, r router.Router, ns func(ctx *fiber.Ctx) string) handler.Handler {
	return &metaHandler{
		c:  s.Client(),
		r:  r,
		ns: ns,
	}
}
