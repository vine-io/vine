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

package vine

import (
	"context"
	"time"

	"github.com/lack-io/vine/server"
)

type function struct {
	cancel context.CancelFunc
	Service
}

func fnHandlerWrapper(f Function) server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			defer f.Done()
			return h(ctx, req, rsp)
		}
	}
}

func fnSubWrapper(f Function) server.SubscriberWrapper {
	return func(s server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			defer f.Done()
			return s(ctx, msg)
		}
	}
}

func newFunction(opts ...Option) Function {
	ctx, cancel := context.WithCancel(context.Background())

	// fore ttl/interval
	fopts := []Option{
		RegisterTTL(time.Minute),
		RegisterInterval(time.Second * 30),
	}

	// prepend to opts
	fopts = append(fopts, opts...)

	// make context the last thing
	fopts = append(fopts, Context(ctx))

	service := newService(fopts...)

	fn := &function{
		cancel:  cancel,
		Service: service,
	}

	service.Server().Init(
		// ensure the service waits for requests to finish
		server.Wait(nil),
		// wrap handlers and subscribers to finish execution
		server.WrapHandler(fnHandlerWrapper(fn)),
		server.WrapSubscriber(fnSubWrapper(fn)),
	)

	return fn
}

func (f *function) Done() error {
	f.cancel()
	return nil
}

func (f *function) Handle(v interface{}) error {
	return f.Service.Server().Handle(
		f.Service.Server().NewHandler(v),
	)
}

func (f *function) Subscribe(topic string, v interface{}) error {
	return f.Service.Server().Subscribe(
		f.Service.Server().NewSubscriber(topic, v),
	)
}
