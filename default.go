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
	"github.com/vine-io/vine/core/broker"
	hBroker "github.com/vine-io/vine/core/broker/http"
	"github.com/vine-io/vine/core/client"
	gclient "github.com/vine-io/vine/core/client/grpc"
	"github.com/vine-io/vine/core/registry"
	mRegistry "github.com/vine-io/vine/core/registry/mdns"
	"github.com/vine-io/vine/core/server"
	gserver "github.com/vine-io/vine/core/server/grpc"
	"github.com/vine-io/vine/lib/cache"
	mCache "github.com/vine-io/vine/lib/cache/memory"
	"github.com/vine-io/vine/lib/config"
	mConfig "github.com/vine-io/vine/lib/config/memory"
	"github.com/vine-io/vine/lib/trace"
	mTrace "github.com/vine-io/vine/lib/trace/memory"
)

func init() {
	// default registry
	registry.DefaultRegistry = mRegistry.NewRegistry()
	// default broker
	broker.DefaultBroker = hBroker.NewBroker()
	// default client
	client.DefaultClient = gclient.NewClient()
	// default server
	server.DefaultServer = gserver.NewServer()
	// default config
	config.DefaultConfig = mConfig.NewConfig()
	// default cache
	cache.DefaultCache = mCache.NewCache()
	// default trace
	trace.DefaultTracer = mTrace.NewTracer()
}
