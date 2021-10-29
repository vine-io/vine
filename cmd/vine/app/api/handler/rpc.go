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
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vine-io/vine/core/client"
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
func RPC(c *gin.Context) {
	c.ClientIP()

	if c.Request.Method == "OPTIONS" {
		return
	}

	if c.Request.Method != "POST" {
		c.JSON(405, "Method not allowed")
		return
	}

	badRequest := func(description string) {
		e := errors.BadRequest("go.vine.rpc", description)
		c.JSON(400, e.Error())
	}

	var service, endpoint, address string
	var request interface{}

	// response content type
	c.Set("Content-Type", "application/json")

	ct := c.GetHeader("Content-Type")

	// Strip charset from Content-Type (like `application/json; charset=UTF-8`)
	if idx := strings.IndexRune(ct, ';'); idx >= 0 {
		ct = ct[:idx]
	}

	switch ct {
	case "application/json":
		var rpcReq rpcRequest

		d := json.NewDecoder(c.Request.Body)
		d.UseNumber()

		if err := d.Decode(&rpcReq); err != nil {
			badRequest(err.Error())
			return
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
				badRequest("error decoding request string: " + err.Error())
				return
			}
		}
	default:
		service = c.PostForm("service")
		endpoint = c.PostForm("endpoint")
		address = c.PostForm("address")
		if len(endpoint) == 0 {
			endpoint = c.PostForm("method")
		}

		d := json.NewDecoder(strings.NewReader(c.PostForm("request")))
		d.UseNumber()

		if err := d.Decode(&request); err != nil {
			badRequest("error decoding request string: " + err.Error())
			return
		}
	}

	if len(service) == 0 {
		badRequest("invalid service")
		return
	}

	if len(endpoint) == 0 {
		badRequest("invalid endpoint")
		return
	}

	// create request/response
	var response json.RawMessage
	var err error
	req := (*cmd.DefaultOptions().Client).NewRequest(service, endpoint, request, client.WithContentType("application/json"))

	// create context
	ctx := helper.RequestToContext(c.Request)

	var opts []client.CallOption

	timeout, _ := strconv.Atoi(c.GetHeader("Timeout"))
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
		c.Writer.Write([]byte(ce.Error()))
	}

	b, _ := response.MarshalJSON()
	c.Set("Content-Length", strconv.Itoa(len(b)))
	c.Writer.Write(b)
}
