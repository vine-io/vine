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
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/lib/api/server/cors"
	"github.com/vine-io/vine/lib/cmd"
	"github.com/vine-io/vine/lib/errors"
	"github.com/vine-io/vine/util/helper"
)

type rpcRequest struct {
	Service  string
	Endpoint string
	Method   string
	Address  string
	Request  interface{}
}

// RPC Handler passes on a JSON or form encoded RPC request to
// a service.
func RPC(c *fiber.Ctx) error {

	if c.Method() == "OPTIONS" {
		return cors.SetHeaders(c)
	}

	if c.Method() != "POST" {
		return fiber.NewError(405, "Method not allowed")
	}

	badRequest := func(description string) error {
		e := errors.BadRequest("go.vine.rpc", description)
		return fiber.NewError(400, e.Error())
	}

	var service, endpoint, address string
	var request interface{}

	// response content type
	c.Set("Content-Type", "application/json")

	ct := c.Get("Content-Type")

	// Strip charset from Content-Type (like `application/json; charset=UTF-8`)
	if idx := strings.IndexRune(ct, ';'); idx >= 0 {
		ct = ct[:idx]
	}

	switch ct {
	case "application/json":
		var rpcReq rpcRequest

		d := json.NewDecoder(bytes.NewBuffer(c.Body()))
		d.UseNumber()

		if err := d.Decode(&rpcReq); err != nil {
			return badRequest(err.Error())
		}

		service = rpcReq.Service
		endpoint = rpcReq.Endpoint
		address = rpcReq.Address
		request = rpcReq.Request
		if len(endpoint) == 0 {
			endpoint = rpcReq.Method
		}

		// JSON as string
		if req, ok := rpcReq.Request.(string); ok {
			d := json.NewDecoder(strings.NewReader(req))
			d.UseNumber()

			if err := d.Decode(&request); err != nil {
				return badRequest("error decoding request string: " + err.Error())

			}
		}
	default:
		service = c.FormValue("service")
		endpoint = c.FormValue("endpoint")
		address = c.FormValue("address")
		if len(endpoint) == 0 {
			endpoint = c.FormValue("method")
		}

		d := json.NewDecoder(strings.NewReader(c.FormValue("request")))
		d.UseNumber()

		if err := d.Decode(&request); err != nil {
			return badRequest("error decoding request string: " + err.Error())
		}
	}

	if len(service) == 0 {
		return badRequest("invalid service")

	}

	if len(endpoint) == 0 {
		return badRequest("invalid endpoint")
	}

	// create request/response
	var response json.RawMessage
	var err error
	req := (*cmd.DefaultOptions().Client).NewRequest(service, endpoint, request, client.WithContentType("application/json"))

	// create context
	ctx := helper.RequestToContext(c)

	var opts []client.CallOption

	timeout, _ := strconv.Atoi(c.Get("Timeout"))
	// set timeout
	if timeout > 0 {
		opts = append(opts, client.WithRequestTimeout(time.Duration(timeout)*time.Second))
	}

	// remote call
	if len(address) > 0 {
		opts = append(opts, client.WithAddress(address))
	}

	// remote call
	err = (*cmd.DefaultOptions().Client).Call(ctx, req, &response, opts...)
	if err != nil {
		ce := errors.Parse(err.Error())
		switch ce.Code {
		case 0:
			// assuming it's totally screwed
			ce.Code = 500
			ce.Id = "go.vine.rpc"
			ce.Status = http.StatusText(500)
			ce.Detail = "error during request: " + ce.Detail
			c.Status(500)
		default:
			c.Status(int(ce.Code))
		}
		_, err = c.Write([]byte(ce.Error()))
		return err
	}

	b, _ := response.MarshalJSON()
	c.Set("Content-Length", strconv.Itoa(len(b)))
	_, err = c.Write(b)
	return err
}
