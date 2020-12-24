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
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/internal/helper"
	"github.com/lack-io/vine/internal/muxer"
	"github.com/lack-io/vine/internal/network/transport"
	"github.com/lack-io/vine/internal/network/transport/grpc"
	"github.com/lack-io/vine/internal/network/tunnel"
	tmucp "github.com/lack-io/vine/internal/network/tunnel/mucp"
	"github.com/lack-io/vine/service"
	log "github.com/lack-io/vine/service/logger"
	net "github.com/lack-io/vine/service/network"
	"github.com/lack-io/vine/service/network/mucp"
	"github.com/lack-io/vine/service/proxy"
	grpcProxy "github.com/lack-io/vine/service/proxy/grpc"
	mucpProxy "github.com/lack-io/vine/service/proxy/mucp"
	muregistry "github.com/lack-io/vine/service/registry"
	"github.com/lack-io/vine/service/router"
	murouter "github.com/lack-io/vine/service/router"
	"github.com/lack-io/vine/service/server"
	mucpServer "github.com/lack-io/vine/service/server/mucp"
)

var (
	// name of the network service
	name = "network"
	// name of the vine network
	networkName = "vine"
	// address is the network address
	address = ":8443"
	// peerAddress is the address the network peers on
	peerAddress = ":8085"
	// set the advertise address
	advertise = ""
	// the tunnel token
	token = "vine"

	// Flags specific to the network
	Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "address",
			Usage:   "Set the address of the network service",
			EnvVars: []string{"VINE_NETWORK_ADDRESS"},
		},
		&cli.StringFlag{
			Name:    "advertise",
			Usage:   "Set the vine network address to advertise",
			EnvVars: []string{"VINE_NETWORK_ADVERTISE"},
		},
		&cli.StringFlag{
			Name:    "gateway",
			Usage:   "Set the default gateway",
			EnvVars: []string{"VINE_NETWORK_GATEWAY"},
		},
		&cli.StringFlag{
			Name:    "network",
			Usage:   "Set the vine network name: vine",
			EnvVars: []string{"VINE_NETWORK"},
		},
		&cli.StringFlag{
			Name:    "nodes",
			Usage:   "Set the vine network nodes to connect to. This can be a comma separated list.",
			EnvVars: []string{"VINE_NETWORK_NODES"},
		},
		&cli.StringFlag{
			Name:    "token",
			Usage:   "Set the vine network token for authentication",
			EnvVars: []string{"VINE_NETWORK_TOKEN"},
		},
	}
)

// Run runs the vine server
func Run(ctx *cli.Context) error {
	if len(ctx.String("server-name")) > 0 {
		name = ctx.String("server-name")
	}
	if len(ctx.String("address")) > 0 {
		address = ctx.String("address")
	}
	if len(ctx.String("peer-address")) > 0 {
		peerAddress = ctx.String("peer-address")
	}
	if len(ctx.String("advertise")) > 0 {
		advertise = ctx.String("advertise")
	}
	if len(ctx.String("network")) > 0 {
		networkName = ctx.String("network")
	}
	if len(ctx.String("token")) > 0 {
		token = ctx.String("token")
	}

	var nodes []string
	if len(ctx.String("nodes")) > 0 {
		nodes = strings.Split(ctx.String("nodes"), ",")
	}

	// Initialise the local service
	service := service.New(
		service.Name(name),
		service.Address(address),
	)

	// create a tunnel
	tunOpts := []tunnel.Option{
		tunnel.Address(peerAddress),
		tunnel.Token(token),
	}

	if ctx.Bool("enable-tls") {
		config, err := helper.TLSConfig(ctx)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		config.InsecureSkipVerify = true

		tunOpts = append(tunOpts, tunnel.Transport(
			grpc.NewTransport(transport.TLSConfig(config)),
		))
	}

	gateway := ctx.String("gateway")
	tun := tmucp.NewTunnel(tunOpts...)
	id := service.Server().Options().Id

	// local tunnel router
	rtr := murouter.DefaultRouter

	rtr.Init(
		router.Network(networkName),
		router.Id(id),
		router.Registry(muregistry.DefaultRegistry),
		router.Gateway(gateway),
		router.Cache(),
	)

	// create new network
	netService := mucp.NewNetwork(
		net.Id(id),
		net.Name(networkName),
		net.Address(peerAddress),
		net.Advertise(advertise),
		net.Nodes(nodes...),
		net.Tunnel(tun),
		net.Router(rtr),
	)

	// local proxy using grpc
	// TODO: reenable after PR
	localProxy := grpcProxy.NewProxy(
		proxy.WithRouter(rtr),
		proxy.WithClient(service.Client()),
	)

	// network proxy
	// used by the network nodes to cluster
	// and share routes or route through
	// each other
	networkProxy := mucpProxy.NewProxy(
		proxy.WithRouter(rtr),
		proxy.WithClient(service.Client()),
		proxy.WithLink("network", netService.Client()),
	)

	// create a handler
	h := mucpServer.DefaultRouter.NewHandler(
		&Network{Network: netService},
	)

	// register the handler
	mucpServer.DefaultRouter.Handle(h)

	// local mux
	localMux := muxer.New(name, localProxy)

	// network mux
	networkMux := muxer.New(name, networkProxy)

	// init the local grpc server
	service.Server().Init(
		server.WithRouter(localMux),
	)

	// set network server to proxy
	netService.Server().Init(
		server.WithRouter(networkMux),
	)

	// connect network
	if err := netService.Connect(); err != nil {
		log.Fatalf("Network failed to connect: %v", err)
	}

	// netClose hard exits if we have problems
	netClose := func(net net.Network) error {
		errChan := make(chan error, 1)

		go func() {
			errChan <- net.Close()
		}()

		select {
		case err := <-errChan:
			return err
		case <-time.After(time.Second):
			return errors.New("Network timeout closing")
		}
	}

	log.Infof("Network [%s] listening on %s", networkName, peerAddress)

	if err := service.Run(); err != nil {
		log.Errorf("Network %s failed: %v", networkName, err)
		netClose(netService)
		os.Exit(1)
	}

	// close the network
	netClose(netService)

	return nil
}
