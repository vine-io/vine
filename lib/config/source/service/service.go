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

package service

import (
	"context"
	"net/http"

	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/lib/config/source"
	"github.com/vine-io/vine/lib/logger"
	"github.com/vine-io/vine/proto/apis/errors"
	proto "github.com/vine-io/vine/proto/services/config"
)

var (
	DefaultName      = "go.vine.config"
	DefaultNamespace = "global"
	DefaultPath      = ""
)

type service struct {
	serviceName string
	namespace   string
	path        string
	opts        source.Options
	client      proto.ConfigService
}

func (m *service) Read() (set *source.ChangeSet, err error) {
	client := proto.NewConfigService(m.serviceName, m.opts.Client)
	req, err := client.Read(context.Background(), &proto.ReadRequest{
		Namespace: m.namespace,
		Path:      m.path,
	})
	if verr, ok := err.(*errors.Error); ok && verr.Code == http.StatusNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return toChangeSet(req.Change.ChangeSet), nil
}

func (m *service) Watch() (w source.Watcher, err error) {
	client := proto.NewConfigService(m.serviceName, m.opts.Client)
	stream, err := client.Watch(context.Background(), &proto.WatchRequest{
		Namespace: m.namespace,
		Path:      m.path,
	})
	if err != nil {
		logger.Error("watch err: ", err)
		return
	}
	return newWatcher(stream)
}

// Write is unsupported
func (m *service) Write(cs *source.ChangeSet) error {
	return nil
}

func (m *service) String() string {
	return "service"
}

func NewSource(opts ...source.Option) source.Source {
	var options source.Options
	for _, o := range opts {
		o(&options)
	}

	addr := DefaultName
	namespace := DefaultNamespace
	path := DefaultPath

	if options.Context != nil {
		a, ok := options.Context.Value(serviceNameKey{}).(string)
		if ok {
			addr = a
		}

		k, ok := options.Context.Value(namespaceKey{}).(string)
		if ok {
			namespace = k
		}

		p, ok := options.Context.Value(pathKey{}).(string)
		if ok {
			path = p
		}
	}

	if options.Client == nil {
		options.Client = client.DefaultClient
	}

	s := &service{
		serviceName: addr,
		opts:        options,
		namespace:   namespace,
		path:        path,
	}

	return s
}
