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

package context

import (
	"context"
	"net/textproto"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lack-io/vine/util/context/metadata"
)

func FromRequest(c *fiber.Ctx) context.Context {
	ctx := c.Context()
	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = make(metadata.Metadata)
	}
	c.Request().Header.VisitAll(func(k, value []byte) {
		key := string(k)
		switch key {
		case "Connection", "Content-Length":
			return
		}
		md[textproto.CanonicalMIMEHeaderKey(key)] = string(value)
	})
	if v, ok := md.Get("X-Forwarded-For"); ok {
		md["X-Forwarded-For"] = v + ", " + c.Context().RemoteAddr().String()
	} else {
		md["X-Forwarded-For"] = c.Context().RemoteAddr().String()
	}
	if _, ok = md.Get("Host"); !ok {
		md["Host"] = string(c.Request().Host())
	}
	// pass http method
	md["Method"] = c.Method()
	return metadata.NewContext(ctx, md)
}

type RequestCtx struct {
	*fiber.Ctx
	ctx context.Context
}

var _ context.Context = (*RequestCtx)(nil)

func (c *RequestCtx) Clone(ctx context.Context) *RequestCtx {
	c.ctx = ctx
	return c
}

func (c *RequestCtx) Context() context.Context {
	return c.ctx
}

func (c *RequestCtx) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c *RequestCtx) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *RequestCtx) Err() error {
	return c.ctx.Err()
}

func (c *RequestCtx) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

func NewRequestCtx(c *fiber.Ctx, ctx context.Context) *RequestCtx {
	return &RequestCtx{
		Ctx: c,
		ctx: ctx,
	}
}
