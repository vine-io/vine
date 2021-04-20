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

package mucp

import (
	"bytes"

	"github.com/lack-io/vine/core/codec"
	"github.com/lack-io/vine/core/transport"
	"github.com/lack-io/vine/util/client/buf"
)

type rpcRequest struct {
	service     string
	method      string
	endpoint    string
	contentType string
	socket      transport.Socket
	codec       codec.Codec
	header      map[string]string
	body        []byte
	rawBody     interface{}
	stream      bool
	first       bool
}

type rpcMessage struct {
	topic       string
	contentType string
	payload     interface{}
	header      map[string]string
	body        []byte
	codec       codec.NewCodec
}

func (r *rpcRequest) Codec() codec.Reader {
	return r.codec
}

func (r *rpcRequest) ContentType() string {
	return r.contentType
}

func (r *rpcRequest) Service() string {
	return r.service
}

func (r *rpcRequest) Method() string {
	return r.method
}

func (r *rpcRequest) Endpoint() string {
	return r.endpoint
}

func (r *rpcRequest) Header() map[string]string {
	return r.header
}

func (r *rpcRequest) Body() interface{} {
	return r.rawBody
}

func (r *rpcRequest) Read() ([]byte, error) {
	// got a body
	if r.first {
		b := r.body
		r.first = false
		return b, nil
	}

	var msg transport.Message
	err := r.socket.Recv(&msg)
	if err != nil {
		return nil, err
	}
	r.header = msg.Header

	return msg.Body, nil
}

func (r *rpcRequest) Stream() bool {
	return r.stream
}

func (r *rpcMessage) ContentType() string {
	return r.contentType
}

func (r *rpcMessage) Topic() string {
	return r.topic
}

func (r *rpcMessage) Payload() interface{} {
	return r.payload
}

func (r *rpcMessage) Header() map[string]string {
	return r.header
}

func (r *rpcMessage) Body() []byte {
	return r.body
}

func (r *rpcMessage) Codec() codec.Reader {
	b := buf.New(bytes.NewBuffer(r.body))
	return r.codec(b)
}
