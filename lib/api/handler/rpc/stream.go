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
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
	json "github.com/json-iterator/go"
	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/core/codec/bytes"
	"github.com/vine-io/vine/lib/api"
	"github.com/vine-io/vine/lib/logger"
)

// serveWebSocket will stream rpc back over websockets assuming json
func serveWebSocket(ctx *gin.Context, service *api.Service, c client.Client) {
	var op int

	ct := ctx.GetHeader("Content-Type")
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
	if proto := ctx.GetHeader("Set-WebSocket-Protocol"); proto != "" {
		for _, p := range strings.Split(proto, ",") {
			switch p {
			case "binary":
				hdr["Set-WebSocket-Protocol"] = []string{"binary"}
				op = websocket.BinaryMessage
			}
		}
	}
	payload, err := requestPayload(ctx.Request)
	if err != nil {
		logger.Error(err)
		return
	}

	upgrader := websocket.Upgrader{
		HandshakeTimeout: 30 * time.Second,
		ReadBufferSize:   1024 * 32,
		WriteBufferSize:  1024 * 32,
		//Subprotocols:      nil,
		Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {

		},
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		EnableCompression: false,
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, ctx.Request.Header)

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

	go writeLoop(conn, stream)

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
			if err := conn.WriteMessage(op, buf); err != nil {
				logger.Error(err)
				return
			}
		}
	}
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

func isStream(c *gin.Context, svc *api.Service) bool {
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
			if v := ep.Metadata["stream"]; v != "" {
				return true
			}
		}
	}
	return false
}

func isWebSocket(c *gin.Context) bool {
	contains := func(key, val string) bool {
		vv := strings.Split(c.GetHeader(key), ",")
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

func multipartHandler(ctx *gin.Context, service *api.Service, c client.Client) {
	if service.Endpoint.Stream != api.Client {
		writeError(ctx, fmt.Errorf("server endpoint must be gRPC client stream"))
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		writeError(ctx, err)
		return
	}

	vals := make(map[string]string)
	for key, values := range ctx.Request.URL.Query() {
		vv, ok := vals[key]
		if !ok {
			vals[key] = strings.Join(values, ",")
		} else {
			vals[key] = vv + "," + strings.Join(values, ",")
		}
	}

	for key, values := range form.Value {
		vv, ok := vals[key]
		if !ok {
			vals[key] = strings.Join(values, ",")
		} else {
			vals[key] = vv + "," + strings.Join(values, ",")
		}
	}

	br, err := json.Marshal(vals)
	if err != nil {
		writeError(ctx, err)
		return
	}

	var request interface{}
	m := json.RawMessage(br)
	request = &m

	ct := "application/json"
	ctx.Request.Header.Set("Content-Type", ct)
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
	defer stream.Close()

	done := make(chan struct{})
	ech := make(chan error, 1)

	process := func(service *api.Service, s client.Stream, files map[string][]*multipart.FileHeader, vals map[string]string) {

		data := map[string]interface{}{}
		for key, values := range vals {
			data[key] = values
		}

		var e error
		if form.File == nil || len(form.File) == 0 {
			if e = stream.Send(data); e != nil {
				ech <- e
			}
		} else {
			var fd multipart.File
			for _, headers := range form.File {
				for _, header := range headers {
					data["name"] = header.Filename
					data["size"] = header.Size

					fd, e = header.Open()
					if e != nil {
						ech <- e
						return
					}

					buf := make([]byte, 4096)
					for {
						n, e1 := fd.Read(buf)
						if e1 != nil && e1 != io.EOF {
							ech <- e1
							_ = fd.Close()
							return
						}

						if n > 0 {
							data["length"] = n
							data["chunk"] = buf[:n]
							if ee := stream.Send(data); ee != nil {
								ech <- ee
								_ = fd.Close()
								return
							}
						}

						if e1 == io.EOF {
							_ = fd.Close()
							break
						}
					}
				}
			}
		}

		if e = stream.CloseSend(); e != nil {
			ech <- e
			return
		}

		close(done)
	}

	go process(service, stream, form.File, vals)

	select {
	case err = <-ech:
		writeError(ctx, err)
		return
	case <-done:
	}

	rsp := stream.Response()
	result, err := rsp.Read()
	if err != nil {
		writeError(ctx, err)
		return
	}

	writeResponse(ctx, result)
	return
}

func isMultipart(c *gin.Context) bool {
	return strings.Contains(c.ContentType(), "multipart/form-data")
}

func downLoadHandler(ctx *gin.Context, service *api.Service, c client.Client) {
	vals := make(map[string]string)
	for key, values := range ctx.Request.URL.Query() {
		vv, ok := vals[key]
		if !ok {
			vals[key] = strings.Join(values, ",")
		} else {
			vals[key] = vv + "," + strings.Join(values, ",")
		}
	}

	for key, values := range ctx.Request.Form {
		vv, ok := vals[key]
		if !ok {
			vals[key] = strings.Join(values, ",")
		} else {
			vals[key] = vv + "," + strings.Join(values, ",")
		}
	}

	br, err := json.Marshal(vals)
	if err != nil {
		writeError(ctx, err)
		return
	}

	request := &api.FileDesc{}
	err = json.Unmarshal(br, request)
	if err != nil {
		writeError(ctx, fmt.Errorf("read request values: %v", err))
		return
	}

	ct := "application/proto"
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
		writeError(ctx, fmt.Errorf("create stream: %v", err))
		return
	}
	defer stream.Close()

	if err = stream.Send(request); err != nil {
		logger.Error(err)
		writeError(ctx, err)
		return
	}

	ctx.Writer.Header().Set("Accept-Ranges", "bytes")
	ctx.Writer.Header().Set("Content-Type", "application/octet-stream")
	disposition := fmt.Sprintf(`attachment; filename="%s"`, request.Name)
	ctx.Writer.Header().Set("Content-Disposition", disposition)

	rsp := stream.Response()
	reader := &responseReader{conn: rsp}

	ctx.Writer.WriteHeader(200)
	_, err = io.Copy(ctx.Writer, reader)
	if err != nil {
		writeError(ctx, err)
		return
	}
	return
}

func isDownLoadLink(s *api.Service) bool {
	return s.Endpoint.Stream == api.Server && strings.HasSuffix(strings.ToLower(s.Endpoint.Name), "download")
}

type responseReader struct {
	conn client.Response
}

func (rr *responseReader) Read(b []byte) (n int, err error) {
	frame := &api.FileHeader{}
	buf, e := rr.conn.Read()
	if e != nil && e != io.EOF {
		return 0, e
	}
	err = proto.Unmarshal(buf, frame)
	if err != nil {
		return 0, err
	}
	if frame.Length > 0 {
		n = copy(b, frame.Chunk[:frame.Length])
		return n, nil
	}
	if e == io.EOF {
		return 0, e
	}
	return
}
