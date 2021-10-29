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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vine-io/vine/lib/errors"

	"github.com/vine-io/vine"
	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/lib/api/handler"
	aapi "github.com/vine-io/vine/lib/api/handler/api"
	"github.com/vine-io/vine/lib/api/handler/event"
	ahttp "github.com/vine-io/vine/lib/api/handler/http"
	arpc "github.com/vine-io/vine/lib/api/handler/rpc"
	aweb "github.com/vine-io/vine/lib/api/handler/web"
	"github.com/vine-io/vine/lib/api/router"
	ctx "github.com/vine-io/vine/util/context"
)

type metaHandler struct {
	c  client.Client
	r  router.Router
	ns func(*http.Request) string
}

func (m *metaHandler) Handle(c *gin.Context) {

	r := c.Request.Clone(ctx.FromRequest(c.Request))
	service, err := m.r.Route(r)
	if err != nil {
		err := errors.InternalServerError(m.ns(c.Request), err.Error())
		c.JSON(500, err.Error())
		return
	}

	// TODO: don't do this ffs
	switch service.Endpoint.Handler {
	// web socket handler
	case aweb.Handler:
		aweb.WithService(service, handler.WithClient(m.c)).Handle(c)
		return
	// proxy handler
	case "proxy", ahttp.Handler:
		ahttp.WithService(service, handler.WithClient(m.c)).Handle(c)
		return
	// rpcx handler
	case arpc.Handler:
		arpc.WithService(service, handler.WithClient(m.c)).Handle(c)
		return
	// event handler
	case event.Handler:
		ev := event.NewHandler(
			handler.WithNamespace(m.ns(c.Request)),
			handler.WithClient(m.c),
		)
		ev.Handle(c)
		return
	// api handler
	case aapi.Handler:
		aapi.WithService(service, handler.WithClient(m.c)).Handle(c)
		return
	// default handler: rpc
	default:
		arpc.WithService(service, handler.WithClient(m.c)).Handle(c)
		return
	}
}

func (m *metaHandler) String() string {
	return "meta"
}

// Meta is a http.Handler that routes based on endpoint metadata
func Meta(s vine.Service, r router.Router, ns func(*http.Request) string) handler.Handler {
	return &metaHandler{
		c:  s.Client(),
		r:  r,
		ns: ns,
	}
}
