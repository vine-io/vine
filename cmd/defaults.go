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

package cmd

import (
	"github.com/lack-io/vine/service/auth"
	authSrv "github.com/lack-io/vine/service/auth/client"
	"github.com/lack-io/vine/service/broker"
	brokerSrv "github.com/lack-io/vine/service/broker/client"
	"github.com/lack-io/vine/service/client"
	grpcCli "github.com/lack-io/vine/service/client/grpc"
	"github.com/lack-io/vine/service/events"
	eventsSrv "github.com/lack-io/vine/service/events/client"
	"github.com/lack-io/vine/service/metrics"
	noopMet "github.com/lack-io/vine/service/metrics/noop"
	"github.com/lack-io/vine/service/network"
	mucpNet "github.com/lack-io/vine/service/network/mucp"
	"github.com/lack-io/vine/service/registry"
	registrySrv "github.com/lack-io/vine/service/registry/client"
	"github.com/lack-io/vine/service/router"
	routerSrv "github.com/lack-io/vine/service/router/client"
	"github.com/lack-io/vine/service/runtime"
	runtimeSrv "github.com/lack-io/vine/service/runtime/client"
	"github.com/lack-io/vine/service/server"
	grpcSvr "github.com/lack-io/vine/service/server/grpc"
	"github.com/lack-io/vine/service/store"
	storeSrv "github.com/lack-io/vine/service/store/client"
)

// setupDefaults sets the default auth, broker etc implementations incase they arent configured by
// a profile. The default implementations are always the RPC implementations.
func setupDefaults() {
	client.DefaultClient = grpcCli.NewClient()
	server.DefaultServer = grpcSvr.NewServer()
	network.DefaultNetwork = mucpNet.NewNetwork()
	metrics.DefaultMetricsReporter = noopMet.New()

	// setup rpc implementations after the client is configured
	auth.DefaultAuth = authSrv.NewAuth()
	broker.DefaultBroker = brokerSrv.NewBroker()
	events.DefaultStream = eventsSrv.NewStream()
	events.DefaultStore = eventsSrv.NewStore()
	registry.DefaultRegistry = registrySrv.NewRegistry()
	router.DefaultRouter = routerSrv.NewRouter()
	store.DefaultStore = storeSrv.NewStore()
	store.DefaultBlobStore = storeSrv.NewBlobStore()
	runtime.DefaultRuntime = runtimeSrv.NewRuntime()
}
