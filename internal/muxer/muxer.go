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

// Package muxer provides proxy muxing
package muxer

import (
	"context"
	"sync"

	"github.com/lack-io/vine/service/client/grpc"
	debug "github.com/lack-io/vine/service/debug/handler"
	"github.com/lack-io/vine/service/proxy"
	"github.com/lack-io/vine/service/server"
	"github.com/lack-io/vine/service/server/mucp"
)

// Server is a proxy muxer that incudes the use of the DefaultHandler
type Server struct {
	// name of service
	Name string
	// Proxy handler
	Proxy proxy.Proxy
	// The default handler
	Handler Handler
}

type Handler interface {
	proxy.Proxy
	NewHandler(interface{}, ...server.HandlerOption) server.Handler
	Handle(server.Handler) error
}

var (
	once sync.Once
)

func (s *Server) ProcessMessage(ctx context.Context, msg server.Message) error {
	if msg.Topic() == s.Name {
		return s.Handler.ProcessMessage(ctx, msg)
	}
	return s.Proxy.ProcessMessage(ctx, msg)
}

func (s *Server) ServeRequest(ctx context.Context, req server.Request, rsp server.Response) error {
	if req.Service() == s.Name {
		return s.Handler.ServeRequest(ctx, req, rsp)
	}
	return s.Proxy.ServeRequest(ctx, req, rsp)
}

func New(name string, p proxy.Proxy) *Server {
	r := mucp.DefaultRouter

	// only register this once
	once.Do(func() {
		r.Handle(
			// inject the debug handler
			r.NewHandler(
				debug.NewHandler(grpc.NewClient()),
				server.InternalHandler(true),
			),
		)
	})

	return &Server{
		Name:    name,
		Proxy:   p,
		Handler: r,
	}
}
