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
	"github.com/lack-io/vine/core/broker"
	"github.com/lack-io/vine/core/broker/http"
	"github.com/lack-io/vine/core/client"
	"github.com/lack-io/vine/core/client/grpc"
	"github.com/lack-io/vine/core/registry"
	"github.com/lack-io/vine/core/registry/mdns"
	"github.com/lack-io/vine/core/router"
	rreg "github.com/lack-io/vine/core/router/registry"
	"github.com/lack-io/vine/core/server"
	serverGrpc "github.com/lack-io/vine/core/server/grpc"
	"github.com/lack-io/vine/lib/config"
	configMemory "github.com/lack-io/vine/lib/config/memory"
	"github.com/lack-io/vine/lib/dao"
	daoNop "github.com/lack-io/vine/lib/dao/nop"
	"github.com/lack-io/vine/lib/store"
	storeMem "github.com/lack-io/vine/lib/store/memory"
	"github.com/lack-io/vine/lib/trace"
	traceMem "github.com/lack-io/vine/lib/trace/memory"
)

func init() {
	// default registry
	registry.DefaultRegistry = mdns.NewRegistry()
	// default broker
	broker.DefaultBroker = http.NewBroker()
	// default client
	client.DefaultClient = grpc.NewClient()
	// default server
	server.DefaultServer = serverGrpc.NewServer()
	// default router
	router.DefaultRouter = rreg.NewRouter()
	// default config
	config.DefaultConfig = configMemory.NewConfig()
	// default dao
	dao.DefaultDialect = daoNop.NewDialect()
	// default store
	store.DefaultStore = storeMem.NewStore()
	// default trace
	trace.DefaultTracer = traceMem.NewTracer()
}
