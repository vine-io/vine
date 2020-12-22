// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpc

import (
	"runtime/debug"

	"google.golang.org/grpc/peer"

	"github.com/lack-io/vine/internal/network/transport"
	pb "github.com/lack-io/vine/proto/transport"
	"github.com/lack-io/vine/service/errors"
	"github.com/lack-io/vine/service/logger"
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
