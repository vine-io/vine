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

package server

import (
	"github.com/lack-io/cli"

	pb "github.com/lack-io/vine/proto/router"
	"github.com/lack-io/vine/service"
	muregistry "github.com/lack-io/vine/service/registry"
	"github.com/lack-io/vine/service/router"
	"github.com/lack-io/vine/service/router/registry"
)

var (
	// name of the router vine service
	name = "router"
	// address is the router vine service bind address
	address = ":8084"
	// network is the network name
	network = router.DefaultNetwork

	// Flags specific to the router
	Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "network",
			Usage:   "Set the vine network name: local",
			EnvVars: []string{"VINE_NETWORK_NAME"},
		},
		&cli.StringFlag{
			Name:    "gateway",
			Usage:   "Set the vine default gateway address. Defaults to none.",
			EnvVars: []string{"VINE_GATEWAY_ADDRESS"},
		},
	}
)

// Run the vine router
func Run(ctx *cli.Context) error {
	if len(ctx.String("server-name")) > 0 {
		name = ctx.String("server-name")
	}
	if len(ctx.String("address")) > 0 {
		address = ctx.String("address")
	}
	if len(ctx.String("network")) > 0 {
		network = ctx.String("network")
	}
	// default gateway address
	var gateway string
	if len(ctx.String("gateway")) > 0 {
		gateway = ctx.String("gateway")
	}

	// Initialise service
	srv := service.New(
		service.Name(name),
		service.Address(address),
	)

	r := registry.NewRouter(
		router.Id(srv.Server().Options().Id),
		router.Address(srv.Server().Options().Id),
		router.Network(network),
		router.Registry(muregistry.DefaultRegistry),
		router.Gateway(gateway),
	)

	// register handlers
	pb.RegisterRouterHandler(srv.Server(), &Router{Router: r})
	pb.RegisterTableHandler(srv.Server(), &Table{Router: r})

	return srv.Run()
}
