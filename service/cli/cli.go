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

// Package service provides a vine service
package service

import (
	"fmt"
	"os"
	"strings"

	ccli "github.com/lack-io/cli"

	"github.com/lack-io/vine/cmd"
	"github.com/lack-io/vine/plugin"
	"github.com/lack-io/vine/service"
	"github.com/lack-io/vine/service/logger"
	prox "github.com/lack-io/vine/service/proxy"
	"github.com/lack-io/vine/service/proxy/grpc"
	"github.com/lack-io/vine/service/proxy/http"
	"github.com/lack-io/vine/service/proxy/mucp"
	muruntime "github.com/lack-io/vine/service/runtime"
	rt "github.com/lack-io/vine/service/runtime"
	"github.com/lack-io/vine/service/server"

	// services
	api "github.com/lack-io/vine/service/api/server"
	auth "github.com/lack-io/vine/service/auth/server"
	broker "github.com/lack-io/vine/service/broker/server"
	config "github.com/lack-io/vine/service/config/server"
	events "github.com/lack-io/vine/service/events/server"
	network "github.com/lack-io/vine/service/network/server"
	proxy "github.com/lack-io/vine/service/proxy/server"
	registry "github.com/lack-io/vine/service/registry/server"
	router "github.com/lack-io/vine/service/router/server"
	runtime "github.com/lack-io/vine/service/runtime/server"
	store "github.com/lack-io/vine/service/store/server"

	// misc commands
	"github.com/lack-io/vine/service/handler/exec"
	"github.com/lack-io/vine/service/handler/file"
)

// Run starts a vine service sidecar to encapsulate any app
func Run(ctx *ccli.Context) {
	name := ctx.String("name")
	address := ctx.String("address")
	endpoint := ctx.String("endpoint")

	metadata := make(map[string]string)
	for _, md := range ctx.StringSlice("metadata") {
		parts := strings.Split(md, "=")
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		val := strings.Join(parts[1:], "=")

		// set the key/val
		metadata[key] = val
	}

	var opts []service.Option
	if len(metadata) > 0 {
		opts = append(opts, service.Metadata(metadata))
	}
	if len(name) > 0 {
		opts = append(opts, service.Name(name))
	}
	if len(address) > 0 {
		opts = append(opts, service.Address(address))
	}

	if len(endpoint) == 0 {
		endpoint = prox.DefaultEndpoint
	}

	var p prox.Proxy

	switch {
	case strings.HasPrefix(endpoint, "grpc"):
		endpoint = strings.TrimPrefix(endpoint, "grpc://")
		p = grpc.NewProxy(prox.WithEndpoint(endpoint))
	case strings.HasPrefix(endpoint, "http"):
		p = http.NewProxy(prox.WithEndpoint(endpoint))
	case strings.HasPrefix(endpoint, "file"):
		endpoint = strings.TrimPrefix(endpoint, "file://")
		p = file.NewProxy(prox.WithEndpoint(endpoint))
	case strings.HasPrefix(endpoint, "exec"):
		endpoint = strings.TrimPrefix(endpoint, "exec://")
		p = exec.NewProxy(prox.WithEndpoint(endpoint))
	default:
		p = mucp.NewProxy(prox.WithEndpoint(endpoint))
	}

	// run the service if asked to
	if ctx.Args().Len() > 0 {
		args := []rt.CreateOption{
			rt.WithCommand(ctx.Args().Slice()...),
			rt.WithOutput(os.Stdout),
		}

		// create new local runtime
		r := muruntime.DefaultRuntime

		// start the runtime
		r.Start()

		// register the service
		r.Create(&rt.Service{
			Name: name,
		}, args...)

		// stop the runtime
		defer func() {
			r.Delete(&rt.Service{
				Name: name,
			})
			r.Stop()
		}()
	}

	logger.Infof("Service [%s] Serving %s at endpoint %s\n", p.String(), name, endpoint)

	// new service
	srv := service.New(opts...)

	// create new muxer
	//	muxer := mux.New(name, p)

	// set the router
	srv.Server().Init(
		server.WithRouter(p),
	)

	// run service
	srv.Run()
}

type srvCommand struct {
	Name    string
	Command ccli.ActionFunc
	Flags   []ccli.Flag
}

var srvCommands = []srvCommand{
	{
		Name:    "api",
		Command: api.Run,
		Flags:   api.Flags,
	},
	{
		Name:    "auth",
		Command: auth.Run,
		Flags:   auth.Flags,
	},
	{
		Name:    "broker",
		Command: broker.Run,
	},
	{
		Name:    "config",
		Command: config.Run,
		Flags:   config.Flags,
	},
	{
		Name:    "events",
		Command: events.Run,
	},
	{
		Name:    "network",
		Command: network.Run,
		Flags:   network.Flags,
	},
	{
		Name:    "proxy",
		Command: proxy.Run,
		Flags:   proxy.Flags,
	},
	{
		Name:    "registry",
		Command: registry.Run,
	},
	{
		Name:    "router",
		Command: router.Run,
		Flags:   router.Flags,
	},
	{
		Name:    "runtime",
		Command: runtime.Run,
		Flags:   runtime.Flags,
	},
	{
		Name:    "store",
		Command: store.Run,
	},
}

func init() {
	// move newAction outside the loop and pass c as an arg to
	// set the scope of the variable
	newAction := func(c srvCommand) func(ctx *ccli.Context) error {
		return func(ctx *ccli.Context) error {
			// configure the loggerger
			logger.DefaultLogger.Init(logger.WithFields(map[string]interface{}{"service": c.Name}))

			// run the service
			c.Command(ctx)
			return nil
		}
	}

	subcommands := make([]*ccli.Command, len(srvCommands))
	for i, c := range srvCommands {
		// construct the command
		command := &ccli.Command{
			Name:   c.Name,
			Flags:  c.Flags,
			Usage:  fmt.Sprintf("Run vine %v", c.Name),
			Action: newAction(c),
		}

		// setup the plugins
		for _, p := range plugin.Plugins(plugin.Module(c.Name)) {
			if cmds := p.Commands(); len(cmds) > 0 {
				command.Subcommands = append(command.Subcommands, cmds...)
			}

			if flags := p.Flags(); len(flags) > 0 {
				command.Flags = append(command.Flags, flags...)
			}
		}

		// set the command
		subcommands[i] = command
	}

	command := &ccli.Command{
		Name:  "service",
		Usage: "Run a vine service",
		Action: func(ctx *ccli.Context) error {
			Run(ctx)
			return nil
		},
		Flags: []ccli.Flag{
			&ccli.StringFlag{
				Name:    "name",
				Usage:   "Name of the service",
				EnvVars: []string{"VINE_SERVICE_NAME"},
				Value:   "service",
			},
			&ccli.StringFlag{
				Name:    "address",
				Usage:   "Address of the service",
				EnvVars: []string{"VINE_SERVICE_ADDRESS"},
			},
			&ccli.StringFlag{
				Name:    "endpoint",
				Usage:   "The local service endpoint (Defaults to localhost:9090); {http, grpc, file, exec}://path-or-address e.g http://localhost:9090",
				EnvVars: []string{"VINE_SERVICE_ENDPOINT"},
			},
			&ccli.StringSliceFlag{
				Name:    "metadata",
				Usage:   "Add metadata as key-value pairs describing the service e.g owner=john@example.com",
				EnvVars: []string{"VINE_SERVICE_METADATA"},
			},
		},
		Subcommands: subcommands,
	}

	// register global plugins and flags
	for _, p := range plugin.Plugins() {
		if cmds := p.Commands(); len(cmds) > 0 {
			command.Subcommands = append(command.Subcommands, cmds...)
		}

		if flags := p.Flags(); len(flags) > 0 {
			command.Flags = append(command.Flags, flags...)
		}
	}

	cmd.Register(command)
}
