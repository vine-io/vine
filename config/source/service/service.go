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
	"context"
	"net/http"

	"github.com/lack-io/vine/client"
	"github.com/lack-io/vine/config/source"
	"github.com/lack-io/vine/errors"
	"github.com/lack-io/vine/log"
	proto "github.com/lack-io/vine/proto/config"
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
		log.Error("watch err: ", err)
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
