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

package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/vine-io/vine/core/server"
)

// rpcStream implements a server side Stream.
type rpcStream struct {
	s       grpc.ServerStream
	request server.Request
}

func (r *rpcStream) Close() error {
	return nil
}

func (r *rpcStream) Error() error {
	return nil
}

func (r *rpcStream) Request() server.Request {
	return r.request
}

func (r *rpcStream) Context() context.Context {
	return r.s.Context()
}

func (r *rpcStream) Send(m interface{}) error {
	return r.s.SendMsg(m)
}

func (r *rpcStream) Recv(m interface{}) error {
	return r.s.RecvMsg(m)
}
