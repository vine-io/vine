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
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gobwas/httphead"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	json "github.com/json-iterator/go"

	apipb "github.com/lack-io/vine/proto/apis/api"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/client/selector"
	raw "github.com/lack-io/vine/service/codec/bytes"
	"github.com/lack-io/vine/service/logger"
)

// serveWebsocket will stream rpc back over websockets assuming json
func serveWebsocket(ctx context.Context, w http.ResponseWriter, r *http.Request, service *apipb.Service, c client.Client) {
	var op ws.OpCode

	ct := r.Header.Get("Content-Type")
	// Strip charset from Content-Type (like `application/json; charset=UTF-8`)
	if idx := strings.IndexRune(ct, ';'); idx >= 0 {
		ct = ct[:idx]
	}

	// check proto from request
	switch ct {
	case "application/json":
		op = ws.OpText
	default:
		op = ws.OpBinary
	}

	hdr := make(http.Header)
	if proto, ok := r.Header["Sec-WebSocket-Protocol"]; ok {
		for _, p := range proto {
			switch p {
			case "binary":
				hdr["Sec-WebSocket-Protocol"] = []string{"binary"}
				op = ws.OpBinary
			}
		}
	}
	payload, err := requestPayload(r)
	if err != nil {
		logger.Error(err)
		return
	}

	upgrader := ws.HTTPUpgrader{Timeout: 5 * time.Second,
		Protocol: func(proto string) bool {
			if strings.Contains(proto, "binary") {
				return true
			}
			// fallback to support all protocols now
			return true
		},
		Extension: func(httphead.Option) bool {
			// disable extensions for compatibility
			return false
		},
		Header: hdr,
	}

	conn, rw, _, err := upgrader.Upgrade(r, w)
	if err != nil {
		logger.Error(err)
		return
	}

	defer func() {
		if err := conn.Close(); err != nil {
			logger.Error(err)
			return
		}
	}()

	var request interface{}
	if !bytes.Equal(payload, []byte(`{}`)) {
		switch ct {
		case "application/json", "":
			m := json.RawMessage(payload)
			request = &m
		default:
			request = &raw.Frame{Data: payload}
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
	stream, err := c.Stream(ctx, req, client.WithSelectOption(so))
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

	go writeLoop(rw, stream)

	rsp := stream.Response()

	// receive from stream and send to client
	for {
		select {
		case <-ctx.Done():
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
			if err := wsutil.WriteServerMessage(rw, op, buf); err != nil {
				logger.Error(err)
				return
			}
			if err = rw.Flush(); err != nil {
				logger.Error(err)
				return
			}
		}
	}
}

// writeLoop
func writeLoop(rw io.ReadWriter, stream client.Stream) {
	// close stream when done
	defer stream.Close()

	for {
		select {
		case <-stream.Context().Done():
			return
		default:
			buf, op, err := wsutil.ReadClientData(rw)
			if err != nil {
				if wserr, ok := err.(wsutil.ClosedError); ok {
					switch wserr.Code {
					case ws.StatusGoingAway:
						// this happens when user leave the page
						return
					case ws.StatusNormalClosure, ws.StatusNoStatusRcvd:
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
			case ws.OpText, ws.OpBinary:
				break
			}
			// send to backend
			// default to trying json
			// if the extracted payload isn't empty lets use it
			request := &raw.Frame{Data: buf}
			if err := stream.Send(request); err != nil {
				logger.Error(err)
				return
			}
		}
	}
}

func isStream(r *http.Request, svc *apipb.Service) bool {
	// check if it's a web socket
	if !isWebSocket(r) {
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

func isWebSocket(r *http.Request) bool {
	contains := func(key, val string) bool {
		vv := strings.Split(r.Header.Get(key), ",")
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
