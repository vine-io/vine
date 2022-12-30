// MIT License
//
// Copyright (c) 2021 Lack
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
	"context"

	"github.com/vine-io/cli"
	"github.com/vine-io/gscheduler"
	"github.com/vine-io/vine/core/broker"
	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/core/registry"
	"github.com/vine-io/vine/lib/cache"
	"github.com/vine-io/vine/lib/cmd"
	"github.com/vine-io/vine/lib/config"
	"github.com/vine-io/vine/lib/dao"
	"github.com/vine-io/vine/lib/scheduler"
	"github.com/vine-io/vine/lib/trace"
)

// Context be used when vine service handling request
type Context struct {
	context.Context

	App       *cli.App
	Broker    broker.Broker
	Client    client.Client
	Config    config.Config
	Trace     trace.Tracer
	Dialect   dao.Dialect
	Cache     cache.Cache
	Registry  registry.Registry
	Scheduler gscheduler.Scheduler
}

// InitContext context.Context => *vine.Context
func InitContext(ctx context.Context) *Context {
	return &Context{
		Context:   ctx,
		App:       cmd.DefaultCmd.App(),
		Broker:    broker.DefaultBroker,
		Client:    client.DefaultClient,
		Config:    config.DefaultConfig,
		Trace:     trace.DefaultTracer,
		Dialect:   dao.DefaultDialect,
		Cache:     cache.DefaultCache,
		Registry:  registry.DefaultRegistry,
		Scheduler: scheduler.DefaultScheduler,
	}
}
