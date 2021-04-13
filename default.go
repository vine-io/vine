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

package vine

import (
	"github.com/lack-io/vine/service/auth"
	authNoop "github.com/lack-io/vine/service/auth/noop"
	"github.com/lack-io/vine/service/broker"
	brokerHttp "github.com/lack-io/vine/service/broker/http"
	"github.com/lack-io/vine/service/client"
	clientGrpc "github.com/lack-io/vine/service/client/grpc"
	"github.com/lack-io/vine/service/config"
	configMemory "github.com/lack-io/vine/service/config/memory"
	"github.com/lack-io/vine/service/dao"
	daoNop "github.com/lack-io/vine/service/dao/nop"
	"github.com/lack-io/vine/service/debug/trace"
	traceMem "github.com/lack-io/vine/service/debug/trace/memory"
	"github.com/lack-io/vine/service/registry"
	registryMdns "github.com/lack-io/vine/service/registry/mdns"
	"github.com/lack-io/vine/service/router"
	rr "github.com/lack-io/vine/service/router/registry"
	"github.com/lack-io/vine/service/server"
	serverGrpc "github.com/lack-io/vine/service/server/grpc"
	"github.com/lack-io/vine/service/store"
	storeMem "github.com/lack-io/vine/service/store/memory"
	"github.com/lack-io/vine/service/transport"
	transportHTTP "github.com/lack-io/vine/service/transport/http"
)

func init() {
	// default registry
	registry.DefaultRegistry = registryMdns.NewRegistry()
	// default transport
	transport.DefaultTransport = transportHTTP.NewTransport()
	// default broker
	broker.DefaultBroker = brokerHttp.NewBroker()
	// default client
	client.DefaultClient = clientGrpc.NewClient()
	// default server
	server.DefaultServer = serverGrpc.NewServer()
	// default router
	router.DefaultRouter = rr.NewRouter()
	// default auth
	auth.DefaultAuth = authNoop.NewAuth()
	// default config
	config.DefaultConfig = configMemory.NewConfig()
	// default dao
	dao.DefaultDialect = daoNop.NewDialect()
	// default store
	store.DefaultStore = storeMem.NewStore()
	// default trace
	trace.DefaultTracer = traceMem.NewTracer()
}
