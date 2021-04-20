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
	"context"
	"errors"
	"io"
	"sync"

	"github.com/lack-io/vine/core/codec"
	"github.com/lack-io/vine/core/server"
)

// Implements the Streamer interface
type rpcStream struct {
	sync.RWMutex
	id      string
	closed  bool
	err     error
	request server.Request
	codec   codec.Codec
	ctx     context.Context
}

func (r *rpcStream) Context() context.Context {
	return r.ctx
}

func (r *rpcStream) Request() server.Request {
	return r.request
}

func (r *rpcStream) Send(msg interface{}) error {
	r.Lock()
	defer r.Unlock()

	resp := codec.Message{
		Target:   r.request.Service(),
		Method:   r.request.Method(),
		Endpoint: r.request.Endpoint(),
		Id:       r.id,
		Type:     codec.Response,
	}

	if err := r.codec.Write(&resp, msg); err != nil {
		r.err = err
	}

	return nil
}

func (r *rpcStream) Recv(msg interface{}) error {
	req := new(codec.Message)
	req.Type = codec.Request

	err := r.codec.ReadHeader(req, req.Type)
	r.Lock()
	defer r.Unlock()
	if err != nil {
		// discard body
		r.codec.ReadBody(nil)
		r.err = err
		return err
	}

	// check the error
	if len(req.Error) > 0 {
		// Check the client closed the stream
		switch req.Error {
		case lastStreamResponseError.Error():
			// discard body
			r.Unlock()
			r.codec.ReadBody(nil)
			r.Lock()
			r.err = io.EOF
			return io.EOF
		default:
			return errors.New(req.Error)
		}
	}

	// we need to stay up to date with sequence numbers
	r.id = req.Id
	r.Unlock()
	err = r.codec.ReadBody(msg)
	r.Lock()
	if err != nil {
		r.err = err
		return err
	}

	return nil
}

func (r *rpcStream) Error() error {
	r.RLock()
	defer r.RUnlock()
	return r.err
}

func (r *rpcStream) Close() error {
	r.Lock()
	defer r.Unlock()
	r.closed = true
	return r.codec.Close()
}
