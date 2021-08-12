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

// Package rpc is a vine rpc handler.
package rpc

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/gofiber/fiber/v2"
	json "github.com/json-iterator/go"
	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/core/codec"
	"github.com/vine-io/vine/core/codec/jsonrpc"
	"github.com/vine-io/vine/core/codec/protorpc"
	"github.com/vine-io/vine/lib/api/handler"
	"github.com/vine-io/vine/lib/logger"
	apipb "github.com/vine-io/vine/proto/apis/api"
	"github.com/vine-io/vine/proto/apis/errors"
	regpb "github.com/vine-io/vine/proto/apis/registry"
	ctx "github.com/vine-io/vine/util/context"
	"github.com/vine-io/vine/util/context/metadata"
	"github.com/vine-io/vine/util/qson"
	"github.com/oxtoacart/bpool"
)

const (
	Handler = "rpc"
)

var (
	// supported json codecs
	jsonCodecs = []string{
		"application/grpc+json",
		"application/json",
		"application/json-rpc",
	}

	// support proto codecs
	protoCodecs = []string{
		"application/grpc",
		"application/grpc+proto",
		"application/proto",
		"application/protobuf",
		"application/proto-rpc",
		"application/octet-stream",
	}

	bufferPool = bpool.NewSizedBufferPool(1024, 8)
)

type RawMessage json.RawMessage

// MarshalJSON returns m as the JSON encoding of m.
func (m RawMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return fmt.Errorf("RawMessage: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

type rpcHandler struct {
	opts handler.Options
	s    *apipb.Service
}

type buffer struct {
	io.ReadCloser
}

func (b *buffer) Write(_ []byte) (int, error) {
	return 0, nil
}

// strategy is a hack for selection
func strategy(services []*regpb.Service) selector.Strategy {
	return func(_ []*regpb.Service) selector.Next {
		// ignore input to this function, use services above
		return selector.Random(services)
	}
}

func (h *rpcHandler) Handle(c *fiber.Ctx) error {

	bsize := handler.DefaultMaxRecvSize
	if h.opts.MaxRecvSize > 0 {
		bsize = h.opts.MaxRecvSize
	}
	c.Context().SetBodyStream(c.Context().RequestBodyStream(), int(bsize))
	var service *apipb.Service

	ct := c.Get("Content-Type")
	if ct == "" {
		c.Request().Header.Set("Content-Type", "application/json")
		ct = "application/json"
	}

	// create context
	cx := ctx.FromRequest(c)
	// set merged context to request
	r := ctx.NewRequestCtx(c, cx)

	if h.s != nil {
		// we were given the service
		service = h.s
	} else if h.opts.Router != nil {
		// try get service from router
		s, err := h.opts.Router.Route(r)
		if err != nil {
			if err.Error() == "service not found" {
				return writeError(c, errors.NotFound("go.vine.api", "invalid url"))
			}
			return writeError(c, errors.BadGateway("go.vine.api", err.Error()))
		}
		service = s
	} else {
		// we have no way of routing the request
		return writeError(c, errors.BadGateway("go.vine.api", "no route found"))
	}

	// Strip charset from Content-Type (like `application/json; charset=UTF-8`)
	if idx := strings.IndexRune(ct, ';'); idx >= 0 {
		ct = ct[:idx]
	}

	// vine client
	cc := h.opts.Client

	// if stream we currently only support json
	if isStream(c, service) {
		// drop older context as it can have timeouts and create new
		//		md, _ := metadata.FromContext(cx)
		// serveWebSocket(context.TODO(), w, r, service, c)
		return serveWebSocket(r, service, cc)
	}

	// create strategy
	so := selector.WithStrategy(strategy(service.Services))

	// walk the standard call path
	// get payload
	br, err := requestPayload(r)
	if err != nil {
		return writeError(c, err)
	}

	var rsp []byte

	switch {
	// proto codecs
	case hasCodec(ct, protoCodecs):
		request := &Message{}
		// if the extracted payload isn't empty lets use it
		if len(br) > 0 {
			request = NewMessage(br)
		}

		// create request/response
		response := &Message{}

		req := cc.NewRequest(
			service.Name,
			service.Endpoint.Name,
			request,
			client.WithContentType(ct),
		)

		// make the call
		if err := cc.Call(cx, req, response, client.WithSelectOption(so)); err != nil {
			return writeError(c, err)
		}

		// marshall response
		rsp, err = response.Marshal()
		if err != nil {
			return writeError(c, err)
		}

	default:
		// if json codec is not present set to json
		if !hasCodec(ct, jsonCodecs) {
			ct = "application/json"
		}

		// default to trying json
		var request RawMessage
		// if the extracted payload isn't empty lets use it
		if len(br) > 0 {
			request = br
		}

		// create request/response
		var response RawMessage

		req := cc.NewRequest(
			service.Name,
			service.Endpoint.Name,
			&request,
			client.WithContentType(ct),
		)
		// make the call
		if err := cc.Call(cx, req, &response, client.WithSelectOption(so)); err != nil {
			return writeError(c, err)
		}

		// marshall response
		rsp, err = response.MarshalJSON()
		if err != nil {
			return writeError(c, err)
		}
	}

	// write the response
	writeResponse(c, rsp)

	return nil
}

func (h *rpcHandler) String() string {
	return "rpc"
}

func hasCodec(ct string, codecs []string) bool {
	for _, c := range codecs {
		if ct == c {
			return true
		}
	}
	return false
}

// requestPayload takes a *http.Request.
// If the request is a GET the query string parameters are extracted and marshaled to JSON and the raw bytes are returned.
// If the request method is a POST the request body is read and returned
func requestPayload(r *ctx.RequestCtx) ([]byte, error) {
	var err error

	// we have to decode json-rpc and proto-rpc because we suck
	// well actually because there's no proxy codec right now
	ct := r.Get("Content-Type")
	switch {
	case strings.Contains(ct, "application/json-rpc"):
		msg := codec.Message{
			Type:   codec.Request,
			Header: make(map[string]string),
		}

		c := jsonrpc.NewCodec(&buffer{io.NopCloser(bytes.NewBuffer(r.Body()))})
		if err = c.ReadHeader(&msg, codec.Request); err != nil {
			return nil, err
		}
		var raw RawMessage
		if err = c.ReadBody(&raw); err != nil {
			return nil, err
		}
		return raw, nil
	case strings.Contains(ct, "application/proto-rpc"), strings.Contains(ct, "application/octet-stream"):
		msg := codec.Message{
			Type:   codec.Request,
			Header: make(map[string]string),
		}
		c := protorpc.NewCodec(&buffer{io.NopCloser(bytes.NewBuffer(r.Body()))})
		if err = c.ReadHeader(&msg, codec.Request); err != nil {
			return nil, err
		}
		var raw Message
		if err = c.ReadBody(&raw); err != nil {
			return nil, err
		}
		return raw.Marshal()
	case strings.Contains(ct, "application/x-www-form-urlencoded"):
		// generate a new set of values from the form
		vals := make(map[string]string)
		r.Request().PostArgs().VisitAll(func(k, v []byte) {
			vals[string(k)] = string(v)
		})
		r.Ctx.Context().QueryArgs().VisitAll(func(k, v []byte) {
			key := string(k)
			vv, ok := vals[key]
			if !ok {
				vals[key] = string(v)
			} else {
				vals[key] = vv + "," + string(v)
			}
		})

		// marshal
		return json.Marshal(vals)
		// TODO: application/grpc
	}

	// otherwise as per usual
	cctx := r.Context()
	// don't use metadata.FromContext as it mangles names
	md, ok := metadata.FromContext(cctx)
	if !ok {
		md = make(map[string]string)
	}

	// allocate maximum
	matches := make(map[string]interface{}, len(md))
	bodydst := ""

	// get fields from url path
	for k, v := range md {
		k = strings.ToLower(k)
		// filter own keys
		if strings.HasPrefix(k, "x-api-field-") {
			matches[strings.TrimPrefix(k, "x-api-field-")] = v
			delete(md, k)
		} else if k == "x-api-body" {
			bodydst = v
			delete(md, k)
		}
	}

	// map of all fields
	req := make(map[string]interface{}, len(md))

	// get fields from url values
	query := string(r.Request().URI().QueryString())
	if len(query) > 0 {
		umd := make(map[string]interface{})
		err = qson.Unmarshal(&umd, query)
		if err != nil {
			return nil, err
		}
		for k, v := range umd {
			matches[k] = v
		}
	}

	// restore context without fields
	r = r.Clone(metadata.NewContext(cctx, md))

	for k, v := range matches {
		ps := strings.Split(k, ".")
		if len(ps) == 1 {
			req[k] = v
			continue
		}
		em := make(map[string]interface{})
		em[ps[len(ps)-1]] = v
		for i := len(ps) - 2; i > 0; i-- {
			nm := make(map[string]interface{})
			nm[ps[i]] = em
			em = nm
		}
		if vm, ok := req[ps[0]]; ok {
			// nested map
			nm := vm.(map[string]interface{})
			for vk, vv := range em {
				nm[vk] = vv
			}
			req[ps[0]] = nm
		} else {
			req[ps[0]] = em
		}
	}
	pathbuf := []byte("{}")
	if len(req) > 0 {
		pathbuf, err = json.Marshal(req)
		if err != nil {
			return nil, err
		}
	}

	urlbuf := []byte("{}")
	out, err := jsonpatch.MergeMergePatches(urlbuf, pathbuf)
	if err != nil {
		return nil, err
	}

	switch r.Method() {
	case "GET":
		// empty response
		if strings.Contains(ct, "application/json") && string(out) == "{}" {
			return out, nil
		} else if string(out) == "{}" && !strings.Contains(ct, "application/json") {
			return []byte{}, nil
		}
		return out, nil
	case "PATCH", "POST", "PUT", "DELETE":
		bodybuf := []byte("{}")
		buf := bufferPool.Get()
		defer bufferPool.Put(buf)
		if _, err := buf.ReadFrom(bytes.NewBuffer(r.Body())); err != nil {
			return nil, err
		}
		if b := buf.Bytes(); len(b) > 0 {
			bodybuf = b
		}
		if bodydst == "" || bodydst == "*" {
			if out, err = jsonpatch.MergeMergePatches(out, bodybuf); err == nil {
				return out, nil
			}
		}
		var jsonbody map[string]interface{}
		if json.Valid(bodybuf) {
			if err = json.Unmarshal(bodybuf, &jsonbody); err != nil {
				return nil, err
			}
		}
		dstmap := make(map[string]interface{})
		ps := strings.Split(bodydst, ".")
		if len(ps) == 1 {
			if jsonbody != nil {
				dstmap[ps[0]] = jsonbody
			} else {
				// old unexpected behaviour
				dstmap[ps[0]] = bodybuf
			}
		} else {
			em := make(map[string]interface{})
			if jsonbody != nil {
				em[ps[len(ps)-1]] = jsonbody
			} else {
				// old unexpected behaviour
				em[ps[len(ps)-1]] = bodybuf
			}
			for i := len(ps) - 2; i > 0; i-- {
				nm := make(map[string]interface{})
				nm[ps[i]] = em
				em = nm
			}
			dstmap[ps[0]] = em
		}

		bodyout, err := json.Marshal(dstmap)
		if err != nil {
			return nil, err
		}

		if out, err = jsonpatch.MergeMergePatches(out, bodyout); err == nil {
			return out, nil
		}

		//fallback to previous unknown behaviour
		return bodybuf, nil
	}

	return []byte{}, nil
}

func writeError(c *fiber.Ctx, err error) error {
	ce := errors.Parse(err.Error())

	switch ce.Code {
	case 0:
		// assuming it's totally screwed
		ce.Code = 500
		ce.Id = "go.vine.api"
		ce.Status = http.StatusText(500)
		ce.Detail = "error during request: " + ce.Detail
		c.Response().SetStatusCode(500)
	default:
		// handle unknown error code
		if ce.Code < 100 || ce.Code > 999 {
			ce.Code = 500
		}
		c.Response().SetStatusCode(int(ce.Code))
	}

	// response content type
	c.Set("Content-Type", "application/json")

	// Set trailers
	if strings.Contains(c.Get("Content-Type"), "application/grpc") {
		c.Set("Trailer", "grpc-status")
		c.Set("Trailer", "grpc-message")
		c.Set("grpc-status", "13")
		c.Set("grpc-message", ce.Detail)
	}

	logger.Errorf("code=%d [%s] %s | %s", ce.Code, c.Method(), c.Path(), ce.Detail)
	return c.JSON(ce)
}

func writeResponse(c *fiber.Ctx, rsp []byte) {
	c.Set("Content-Type", c.Get("Content-Type"))
	c.Set("Content-Length", strconv.Itoa(len(rsp)))

	// Set trailers
	if strings.Contains(c.Get("Content-Type"), "application/grpc") {
		c.Set("Trailer", "grpc-status")
		c.Set("Trailer", "grpc-message")
		c.Set("grpc-status", "0")
		c.Set("grpc-message", "")
	}

	code := http.StatusOK
	// write 204 status if rsp is nil
	if len(rsp) == 0 {
		code = http.StatusNoContent
	}

	// write response
	c.Response().SetStatusCode(code)
	_, err := c.Write(rsp)
	if err != nil {
		logger.Error(err)
	}
	logger.Infof("code=%d [%s] %s", code, c.Method(), c.Path())
}

func NewHandler(opts ...handler.Option) handler.Handler {
	options := handler.NewOptions(opts...)
	return &rpcHandler{
		opts: options,
	}
}

func WithService(s *apipb.Service, opts ...handler.Option) handler.Handler {
	options := handler.NewOptions(opts...)
	return &rpcHandler{
		opts: options,
		s:    s,
	}
}
