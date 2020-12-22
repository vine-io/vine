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

// Package proxy is a transparent proxy built on the vine/server
package proxy

import (
	"context"

	"github.com/lack-io/vine/service/server"
)

// Proxy can be used as a proxy server for vine services
type Proxy interface {
	// ProcessMessage handles inbound messages
	ProcessMessage(context.Context, server.Message) error
	// ServeRequest handles inbound requests
	ServeRequest(context.Context, server.Request, server.Response) error
	// Name of the proxy protocol
	String() string
}

var (
	DefaultEndpoint = "localhost:9090"
)
