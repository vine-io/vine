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

package mucp

import (
	"github.com/lack-io/vine/internal/codec"
	"github.com/lack-io/vine/internal/network/transport"
	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/service/broker"
	"github.com/lack-io/vine/service/registry"
	"github.com/lack-io/vine/service/server"
)

func newOptions(opts ...server.Option) server.Options {
	options := server.Options{
		Codecs:           make(map[string]codec.NewCodec),
		Metadata:         map[string]string{},
		RegisterInterval: server.DefaultRegisterInterval,
		RegisterTTL:      server.DefaultRegisterTTL,
	}

	for _, o := range opts {
		o(&options)
	}

	if options.Auth == nil {
		options.Auth = auth.DefaultAuth
	}

	if options.Broker == nil {
		options.Broker = broker.DefaultBroker
	}

	if options.Registry == nil {
		options.Registry = registry.DefaultRegistry
	}

	if options.Transport == nil {
		options.Transport = transport.DefaultTransport
	}

	if options.RegisterCheck == nil {
		options.RegisterCheck = server.DefaultRegisterCheck
	}

	if len(options.Address) == 0 {
		options.Address = server.DefaultAddress
	}

	if len(options.Name) == 0 {
		options.Name = server.DefaultName
	}

	if len(options.Id) == 0 {
		options.Id = server.DefaultId
	}

	if len(options.Version) == 0 {
		options.Version = server.DefaultVersion
	}

	return options
}
