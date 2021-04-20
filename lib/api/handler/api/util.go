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

package api

import (
	"bytes"
	"mime"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/lack-io/vine/core/client/selector"
	apipb "github.com/lack-io/vine/proto/apis/api"
	regpb "github.com/lack-io/vine/proto/apis/registry"
	"github.com/oxtoacart/bpool"
)

var (
	// need to calculate later to specify useful defaults
	bufferPool = bpool.NewSizedBufferPool(1024, 8)
)

func requestToProto(c *fiber.Ctx) (*apipb.Request, error) {
	req := &apipb.Request{
		Path:   string(c.Request().URI().Path()),
		Method: c.Method(),
		Header: make(map[string]*apipb.Pair),
		Get:    make(map[string]*apipb.Pair),
		Post:   make(map[string]*apipb.Pair),
		Url:    c.Request().URI().String(),
	}

	ct, _, err := mime.ParseMediaType(c.Get("Content-Type"))
	if err != nil {
		ct = "text/plain; charset=UTF-8" //default CT is text/plain
		c.Set("Content-Type", ct)
	}

	//set the body:
	if len(c.Body()) != 0 {
		switch ct {
		case "application/x-www-form-urlencoded":
			// expect form vals in Post data
		default:
			buf := bufferPool.Get()
			defer bufferPool.Put(buf)
			if _, err = buf.ReadFrom(bytes.NewBuffer(c.Body())); err != nil {
				return nil, err
			}
			req.Body = buf.String()
		}
	}

	// Set X-Forwarded-For if it does not exist
	if ip, _, err := net.SplitHostPort(c.IP()); err == nil {
		if prior := c.Get("X-Forwarded-For"); prior != "" {
			ip = prior + ", " + ip
		}

		// Set the header
		req.Header["X-Forwarded-For"] = &apipb.Pair{
			Key:    "X-Forwarded-For",
			Values: []string{ip},
		}
	}

	// Host is stripped from net/http Headers so let's add it
	req.Header["Host"] = &apipb.Pair{
		Key:    "Host",
		Values: []string{string(c.Request().Host())},
	}

	// Get data
	c.Request().URI().QueryArgs().VisitAll(func(k, v []byte) {
		key, value := string(k), string(v)
		header, ok := req.Get[key]
		if !ok {
			header = &apipb.Pair{
				Key: key,
			}
			req.Get[key] = header
		}
		header.Values = append(header.Values, value)
	})

	// Post data
	c.Request().PostArgs().VisitAll(func(k, v []byte) {
		key, value := string(k), string(v)
		header, ok := req.Post[key]
		if !ok {
			header = &apipb.Pair{
				Key: key,
			}
			req.Get[key] = header
		}
		header.Values = append(header.Values, value)
	})

	c.Request().Header.VisitAll(func(k, v []byte) {
		key, value := string(k), string(v)
		header, ok := req.Header[key]
		if !ok {
			header = &apipb.Pair{
				Key: key,
			}
			req.Get[key] = header
		}
		header.Values = append(header.Values, value)
	})

	return req, nil
}

// strategy is a hack for selection
func strategy(services []*regpb.Service) selector.Strategy {
	return func(_ []*regpb.Service) selector.Next {
		// ignore input to this function, use services above
		return selector.Random(services)
	}
}
