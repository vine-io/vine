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

package vine

import (
	"github.com/lack-io/vine/service/broker"
	brokerHttp "github.com/lack-io/vine/service/broker/http"
	"github.com/lack-io/vine/service/client"
	clientGrpc "github.com/lack-io/vine/service/client/grpc"
	"github.com/lack-io/vine/service/debug/trace"
	traceMem "github.com/lack-io/vine/service/debug/trace/memory"
	"github.com/lack-io/vine/service/network/transport"
	transportHTTP "github.com/lack-io/vine/service/network/transport/http"
	"github.com/lack-io/vine/service/registry"
	registryMdns "github.com/lack-io/vine/service/registry/mdns"
	"github.com/lack-io/vine/service/server"
	serverGrpc "github.com/lack-io/vine/service/server/grpc"
	"github.com/lack-io/vine/service/store"
	storeMem "github.com/lack-io/vine/service/store/memory"
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
	// default store
	store.DefaultStore = storeMem.NewStore()
	// default trace
	trace.DefaultTracer = traceMem.NewTracer()
}
