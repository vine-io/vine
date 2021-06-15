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

package rpc

import (
	b "bytes"
	"net/http"
	"strings"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	json "github.com/json-iterator/go"
	"github.com/lack-io/vine/core/client"
	"github.com/lack-io/vine/core/client/selector"
	"github.com/lack-io/vine/core/codec/bytes"
	"github.com/lack-io/vine/lib/logger"
	apipb "github.com/lack-io/vine/proto/apis/api"
	ctx "github.com/lack-io/vine/util/context"
	"github.com/valyala/fasthttp"
)

// serveWebSocket will stream rpc back over websockets assuming json
func serveWebSocket(r *ctx.RequestCtx, service *apipb.Service, c client.Client) error {
	var op int

	ct := r.Get("Content-Type")
	// Strip charset from Content-Type (like `application/json; charset=UTF-8`)
	if idx := strings.IndexRune(ct, ';'); idx >= 0 {
		ct = ct[:idx]
	}

	// check proto from request
	switch ct {
	case "application/json":
		op = websocket.TextMessage // TextMessage
	default:
		op = websocket.BinaryMessage // BinaryMessage
	}

	hdr := make(http.Header)
	if proto := r.Get("Set-WebSocket-Protocol"); proto != "" {
		for _, p := range strings.Split(proto, ",") {
			switch p {
			case "binary":
				hdr["Set-WebSocket-Protocol"] = []string{"binary"}
				op = websocket.BinaryMessage
			}
		}
	}
	payload, err := requestPayload(r)
	if err != nil {
		logger.Error(err)
		return err
	}

	upgrader := websocket.FastHTTPUpgrader{
		HandshakeTimeout: 5 * time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		//Subprotocols:      nil,
		//Error:             nil,
		CheckOrigin: func(c *fasthttp.RequestCtx) bool {
			return true
		},
		EnableCompression: false,
	}

	return upgrader.Upgrade(r.Ctx.Context(), func(conn *websocket.Conn) {

		defer func() {
			if err := conn.Close(); err != nil {
				logger.Error(err)
				return
			}
		}()

		var request interface{}
		if !b.Equal(payload, []byte(`{}`)) {
			switch ct {
			case "application/json", "":
				m := json.RawMessage(payload)
				request = &m
			default:
				request = &bytes.Frame{Data: payload}
			}
		}

		// we always need to set content type for message
		if ct == "" {
			ct = "application/json"
		}
		req := c.NewRequest(
			service.Name,
			service.Endpoint.Name,
			request,
			client.WithContentType(ct),
			client.StreamingRequest(),
		)

		so := selector.WithStrategy(strategy(service.Services))
		// create a new stream
		stream, err := c.Stream(r, req, client.WithSelectOption(so))
		if err != nil {
			logger.Error(err)
			return
		}

		if request != nil {
			if err = stream.Send(request); err != nil {
				logger.Error(err)
				return
			}
		}

		go writeLoop(conn, stream)

		rsp := stream.Response()

		// receive from stream and send to client
		for {
			select {
			case <-r.Context().Done():
				return
			case <-stream.Context().Done():
				return
			default:
				// read backend response body
				buf, err := rsp.Read()
				if err != nil {
					// wants to avoid import  grpc/status.Status
					if strings.Contains(err.Error(), "context canceled") {
						return
					}
					logger.Error(err)
					return
				}

				// write the response
				if err := conn.WriteMessage(op, buf); err != nil {
					logger.Error(err)
					return
				}
			}
		}
	})
}

// writeLoop
func writeLoop(conn *websocket.Conn, stream client.Stream) {
	// close stream when done
	defer stream.Close()

	for {
		select {
		case <-stream.Context().Done():
			return
		default:
			op, buf, err := conn.ReadMessage()
			if err != nil {
				if wserr, ok := err.(*websocket.CloseError); ok {
					switch wserr.Code {
					case websocket.CloseGoingAway:
						// this happens when user leave the page
						return
					case websocket.CloseNormalClosure, websocket.CloseNoStatusReceived:
						// this happens when user close ws connection, or we don't get any status
						return
					}
				}
				logger.Error(err)
				return
			}
			switch op {
			default:
				// not relevant
				continue
			case websocket.TextMessage, websocket.BinaryMessage:
				break
			}
			// send to backend
			// default to trying json
			// if the extracted payload isn't empty lets use it
			request := &bytes.Frame{Data: buf}
			if err := stream.Send(request); err != nil {
				logger.Error(err)
				return
			}
		}
	}
}

func isStream(c *fiber.Ctx, svc *apipb.Service) bool {
	// check if it's a web socket
	if !isWebSocket(c) {
		return false
	}
	// check if the endpoint supports streaming
	for _, service := range svc.Services {
		for _, ep := range service.Endpoints {
			// skip if it doesn't match the name
			if ep.Name != svc.Endpoint.Name {
				continue
			}
			// matched if the name
			if v := ep.Metadata["stream"]; v == "true" {
				return true
			}
		}
	}
	return false
}

func isWebSocket(c *fiber.Ctx) bool {
	contains := func(key, val string) bool {
		vv := strings.Split(c.Get(key), ",")
		for _, v := range vv {
			if val == strings.ToLower(strings.TrimSpace(v)) {
				return true
			}
		}
		return false
	}

	if contains("Connection", "upgrade") && contains("Upgrade", "websocket") {
		return true
	}

	return false
}
