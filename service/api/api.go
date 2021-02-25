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

package api

import (
	"errors"
	"regexp"
	"strings"

	apipb "github.com/lack-io/vine/proto/apis/api"
	"github.com/lack-io/vine/service/server"
)

type Api interface {
	// Initialise options
	Init(...Option) error
	// Get the options
	Options() Options
	// Register a http handler
	Register(*apipb.Endpoint) error
	// Register a route
	Deregister(*apipb.Endpoint) error
	// Implemenation of api
	String() string
}

type Options struct{}

type Option func(*Options) error

func strip(s string) string {
	return strings.TrimSpace(s)
}

func slice(s string) []string {
	var sl []string

	for _, p := range strings.Split(s, ",") {
		if str := strip(p); len(str) > 0 {
			sl = append(sl, str)
		}
	}

	return sl
}

// Encode encodes an endpoint to endpoint metadata
func Encode(e *apipb.Endpoint) map[string]string {
	if e == nil {
		return nil
	}

	// endpoint map
	ep := make(map[string]string)

	// set vals only if they exist
	set := func(k, v string) {
		if len(v) == 0 {
			return
		}
		ep[k] = v
	}

	set("endpoint", e.Name)
	set("description", e.Description)
	set("handler", e.Handler)
	set("method", strings.Join(e.Method, ","))
	set("path", strings.Join(e.Path, ","))
	set("host", strings.Join(e.Host, ","))

	return ep
}

// Decode decodes endpoint metadata into an endpoint
func Decode(e map[string]string) *apipb.Endpoint {
	if e == nil {
		return nil
	}

	return &apipb.Endpoint{
		Name:        e["endpoint"],
		Description: e["description"],
		Method:      slice(e["method"]),
		Path:        slice(e["path"]),
		Host:        slice(e["host"]),
		Handler:     e["handler"],
	}
}

// Validate validates an endpoint to guarantee it won't blow up when being served
func Validate(e *apipb.Endpoint) error {
	if e == nil {
		return errors.New("endpoint is nil")
	}

	if len(e.Name) == 0 {
		return errors.New("name required")
	}

	for _, p := range e.Path {
		ps := p[0]
		pe := p[len(p)-1]

		if ps == '^' && pe == '$' {
			_, err := regexp.CompilePOSIX(p)
			if err != nil {
				return err
			}
		} else if ps == '^' && pe != '$' {
			return errors.New("invalid path")
		} else if ps != '^' && pe == '$' {
			return errors.New("invalid path")
		}
	}

	if len(e.Handler) == 0 {
		return errors.New("invalid handler")
	}

	return nil
}

/*
Design ideas

// Gateway is an api gateway interface
type Gateway interface {
	// Register a http handler
	Handle(pattern string, http.Handler)
	// Register a route
	RegisterRoute(r Route)
	// Init initialises the command line.
	// It also parses further options.
	Init(...Option) error
	// Run the gateway
	Run() error
}

// NewGateway returns a new api gateway
func NewGateway() Gateway {
	return newGateway()
}
*/

// WithEndpoint returns a server.HandlerOption with endpoint metadata set
//
// Usage:
//
// 	proto.RegisterHandler(service.Server(), new(Handler), api.WithEndpoint(
//		&api.Endpoint{
//			Name: "Greeter.Hello",
//			Path: []string{"/greeter"},
//		},
//	))
func WithEndpoint(e *apipb.Endpoint) server.HandlerOption {
	return server.EndpointMetadata(e.Name, Encode(e))
}
