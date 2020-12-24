// Copyright 2020 The vine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"github.com/lack-io/vine/internal/debug/trace"
	memTrace "github.com/lack-io/vine/internal/debug/trace/memory"
	"github.com/lack-io/vine/internal/network/transport"
	httpTransport "github.com/lack-io/vine/internal/network/transport/http"
	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/service/auth/noop"
	"github.com/lack-io/vine/service/broker"
	httpBroker "github.com/lack-io/vine/service/broker/http"
	"github.com/lack-io/vine/service/client"
	gcli "github.com/lack-io/vine/service/client/grpc"
	"github.com/lack-io/vine/service/client/selector"
	regSelector "github.com/lack-io/vine/service/client/selector/registry"
	"github.com/lack-io/vine/service/registry"
	"github.com/lack-io/vine/service/registry/mdns"
	"github.com/lack-io/vine/service/router"
	regRouter "github.com/lack-io/vine/service/router/registry"
	"github.com/lack-io/vine/service/runtime"
	localRuntime "github.com/lack-io/vine/service/runtime/local"
	"github.com/lack-io/vine/service/server"
	gsrv "github.com/lack-io/vine/service/server/grpc"
	mucpServer "github.com/lack-io/vine/service/server/mucp"
	"github.com/lack-io/vine/service/store"
	memStore "github.com/lack-io/vine/service/store/memory"
)

func init() {
	// default auth
	auth.DefaultAuth = noop.NewAuth()
	// default runtime
	runtime.DefaultRuntime = localRuntime.NewRuntime()
	// default transport
	transport.DefaultTransport = httpTransport.NewTransport()
	// default broker
	broker.DefaultBroker = httpBroker.NewBroker()
	// default registry
	registry.DefaultRegistry = mdns.NewRegistry()
	// default router
	router.DefaultRouter = regRouter.NewRouter()
	// default selector
	selector.DefaultSelector = regSelector.NewSelector()
	// default client
	client.DefaultClient = gcli.NewClient()
	// default server
	server.DefaultRouter = mucpServer.NewRpcRouter()
	server.DefaultServer = gsrv.NewServer()
	// default store
	store.DefaultStore = memStore.NewStore()
	// default trace
	trace.DefaultTracer = memTrace.NewTracer()
}
