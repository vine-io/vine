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

// Package event provides a handler which publishes an event
package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/oxtoacart/bpool"

	"github.com/vine-io/vine/lib/api"
	"github.com/vine-io/vine/lib/api/handler"
	ctx "github.com/vine-io/vine/util/context"
)

var (
	bufferPool = bpool.NewSizedBufferPool(1024, 8)
)

type event struct {
	opts handler.Options
}

var (
	Handler   = "event"
	versionRe = regexp.MustCompilePOSIX("^v[0-9]+$")
)

func eventName(parts []string) string {
	return strings.Join(parts, ".")
}

func evRoute(ns, p string) (string, string) {
	p = path.Clean(p)
	p = strings.TrimPrefix(p, "/")

	if len(p) == 0 {
		return ns, "event"
	}

	parts := strings.Split(p, "/")

	// no path
	if len(parts) == 0 {
		// topic: namespace
		// action: event
		return strings.Trim(ns, "."), "event"
	}

	// Treat /v[0-9]+ as versioning
	// /v1/foo/bar => topic: v1.foo action: bar
	if len(parts) >= 2 && versionRe.Match([]byte(parts[0])) {
		topic := ns + "." + strings.Join(parts[:2], ".")
		action := eventName(parts[1:])
		return topic, action
	}

	// /foo => topic: ns.foo action: foo
	// /foo/bar => topic: ns.foo action: bar
	topic := ns + "." + strings.Join(parts[:1], ".")
	action := eventName(parts[1:])

	return topic, action
}

func (e *event) Handle(c *fiber.Ctx) error {
	bsize := handler.DefaultMaxRecvSize
	if e.opts.MaxRecvSize > 0 {
		bsize = e.opts.MaxRecvSize
	}
	c.Context().SetBodyStream(c.Context().RequestBodyStream(), int(bsize))

	// request to topic:event
	// create event
	// publish to topic
	topic, action := evRoute(e.opts.Namespace, string(c.Request().URI().Path()))

	// create event
	ev := &api.Event{
		Name: action,
		// TODO: dedupe event
		Id:        fmt.Sprintf("%s-%s-%s", topic, action, uuid.New().String()),
		Header:    make(map[string]*api.Pair),
		Timestamp: time.Now().Unix(),
	}

	c.Request().Header.VisitAll(func(k, v []byte) {
		key, vals := string(k), string(v)
		header, ok := ev.Header[key]
		if !ok {
			header = &api.Pair{
				Key: key,
			}
			ev.Header[key] = header
		}
		header.Values = []string{vals}
	})

	// set body
	if c.Method() == "GET" {
		bytes, _ := json.Marshal(c.Request().URI().QueryString())
		ev.Data = string(bytes)
	} else {
		// Read body
		buf := bufferPool.Get()
		defer bufferPool.Put(buf)
		if _, err := buf.ReadFrom(bytes.NewBuffer(c.Body())); err != nil {
			return fiber.NewError(500, err.Error())
		}
		ev.Data = buf.String()
	}

	// get client
	cc := e.opts.Client

	// create publication
	p := cc.NewMessage(topic, ev)

	// publish event
	if err := cc.Publish(ctx.FromRequest(c), p); err != nil {
		return fiber.NewError(500, err.Error())
	}
	return nil
}

func (e *event) String() string {
	return "event"
}

func NewHandler(opts ...handler.Option) handler.Handler {
	return &event{
		opts: handler.NewOptions(opts...),
	}
}
