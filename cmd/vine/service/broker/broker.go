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

// Package broker is the vine broker
package broker

import (
	"time"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/cmd/vine/service/broker/handler"
	pb "github.com/lack-io/vine/proto/broker"
	"github.com/lack-io/vine/service"
)

var (
	// Name of the broker
	Name = "go.vine.broker"
	// The address of the broker
	Address = ":8001"
)

func Run(ctx *cli.Context, srvOpts ...service.Option) {

	if len(ctx.String("server-name")) > 0 {
		Name = ctx.String("server-name")
	}
	if len(ctx.String("address")) > 0 {
		Address = ctx.String("address")
	}

	// Init plugins
	for _, p := range Plugins() {
		p.Init(ctx)
	}

	// service opts
	srvOpts = append(srvOpts, service.Name(Name))
	if i := time.Duration(ctx.Int("register-ttl")); i > 0 {
		srvOpts = append(srvOpts, service.RegisterTTL(i*time.Second))
	}
	if i := time.Duration(ctx.Int("register-interval")); i > 0 {
		srvOpts = append(srvOpts, service.RegisterInterval(i*time.Second))
	}

	// set address
	if len(Address) > 0 {
		srvOpts = append(srvOpts, service.Address(Address))
	}

	// new service
	srv := service.NewService(srvOpts...)

	// connect to the broker
	srv.Options().Broker.Connect()

	// register the broker handler
	pb.RegisterBrokerHandler(srv.Server(), &handler.Broker{
		// using the mdns broker
		Broker: srv.Options().Broker,
	})

	// run the service
	srv.Run()
}

func Commands(options ...service.Option) []*cli.Command {
	command := &cli.Command{
		Name:  "broker",
		Usage: "Run the message broker",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "address",
				Usage:   "Set the broker http address e.g 0.0.0.0:8001",
				EnvVars: []string{"VINE_SERVER_ADDRESS"},
			},
		},
		Action: func(ctx *cli.Context) error {
			Run(ctx, options...)
			return nil
		},
	}

	for _, p := range Plugins() {
		if cmds := p.Commands(); len(cmds) > 0 {
			command.Subcommands = append(command.Subcommands, cmds...)
		}

		if flags := p.Flags(); len(flags) > 0 {
			command.Flags = append(command.Flags, flags...)
		}
	}

	return []*cli.Command{command}
}
