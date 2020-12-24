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
	"github.com/lack-io/vine/service/client"
	gcli "github.com/lack-io/vine/service/client/grpc"
	"github.com/lack-io/vine/service/server"
	gsrv "github.com/lack-io/vine/service/server/grpc"
	"github.com/lack-io/vine/service/store"
	memStore "github.com/lack-io/vine/service/store/memory"
	"github.com/lack-io/vine/util/debug/trace"
	memTrace "github.com/lack-io/vine/util/debug/trace/memory"
)

func init() {
	// default client
	client.DefaultClient = gcli.NewClient()
	// default server
	server.DefaultServer = gsrv.NewServer()
	// default store
	store.DefaultStore = memStore.NewStore()
	// default trace
	trace.DefaultTracer = memTrace.NewTracer()
}
