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
	"runtime/debug"

	"google.golang.org/grpc/peer"

	"github.com/lack-io/vine/core/transport"
	"github.com/lack-io/vine/lib/logger"
	"github.com/lack-io/vine/proto/apis/errors"
	pb "github.com/lack-io/vine/proto/services/transport"
)

// vineTransport satisfies the pb.TransportServer interface
type vineTransport struct {
	addr string
	fn   func(transport.Socket)
}

func (m *vineTransport) Stream(ts pb.Transport_StreamServer) (err error) {

	sock := &grpcTransportSocket{
		stream: ts,
		local:  m.addr,
	}

	p, ok := peer.FromContext(ts.Context())
	if ok {
		sock.remote = p.Addr.String()
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error(r, string(debug.Stack()))
			sock.Close()
			err = errors.InternalServerError("go.vine.transport", "panic recovered: %v", r)
		}
	}()

	// execute socket func
	m.fn(sock)

	return err
}
