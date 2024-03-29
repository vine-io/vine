// MIT License
//
// Copyright (c) 2020 The vine Authors
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

package grpc

import (
	"context"
	"io"
	"sync"

	"google.golang.org/grpc"

	"github.com/vine-io/vine/core/client"
)

// Implements the streamer interface
type grpcStream struct {
	sync.RWMutex
	closed   bool
	err      error
	conn     *grpc.ClientConn
	stream   grpc.ClientStream
	request  client.Request
	response client.Response
	ctx      context.Context
	cancel   func()
}

func (g *grpcStream) Context() context.Context {
	return g.ctx
}

func (g *grpcStream) Request() client.Request {
	return g.request
}

func (g *grpcStream) Response() client.Response {
	return g.response
}

func (g *grpcStream) Send(msg interface{}) error {
	if err := g.stream.SendMsg(msg); err != nil {
		g.setError(err)
		return err
	}
	return nil
}

func (g *grpcStream) Recv(msg interface{}) (err error) {
	defer g.setError(err)
	if err = g.stream.RecvMsg(msg); err != nil {
		// #202 - inconsistent gRPC stream behavior
		// the only way to tell if the stream is done is when we get an EOF on the Recv
		// here we should close the underlying gRPC ClientConn
		closeErr := g.Close()
		if err == io.EOF && closeErr != nil {
			err = closeErr
		}
	}
	return
}

func (g *grpcStream) Error() error {
	g.RLock()
	defer g.RUnlock()
	return g.err
}

func (g *grpcStream) setError(e error) {
	g.Lock()
	g.err = e
	g.Unlock()
}

// Close the gRPC send stream and gRPC connection
// #202 - inconsistent gRPC stream behavior
// The underlying gRPC stream should not be closed here since the
// stream should still be able to receive after this function call
func (g *grpcStream) Close() error {
	g.Lock()
	defer g.Unlock()

	if g.closed {
		_ = g.conn.Close()
		return nil
	}
	// cancel the context
	defer g.cancel()
	g.closed = true
	_ = g.stream.CloseSend()
	return g.conn.Close()
}

// CloseSend the gRPC send stream
func (g *grpcStream) CloseSend() error {
	g.Lock()
	defer g.Unlock()

	if g.closed {
		return nil
	}

	g.closed = true
	return g.stream.CloseSend()
}
