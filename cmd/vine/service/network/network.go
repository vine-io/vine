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

// package network implements vine network node
package network

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lack-io/cli"
	router2 "github.com/lack-io/vine/core/router"
	registry2 "github.com/lack-io/vine/core/router/registry"

	"github.com/lack-io/vine"
	mcli "github.com/lack-io/vine/cmd/vine/client/cli"
	"github.com/lack-io/vine/cmd/vine/service/network/api"
	netdns "github.com/lack-io/vine/cmd/vine/service/network/dns"
	"github.com/lack-io/vine/cmd/vine/service/network/handler"
	"github.com/lack-io/vine/core/server"
	"github.com/lack-io/vine/core/transport"
	"github.com/lack-io/vine/core/transport/quic"
	log "github.com/lack-io/vine/lib/logger"
	"github.com/lack-io/vine/lib/network"
	"github.com/lack-io/vine/lib/network/resolver"
	"github.com/lack-io/vine/lib/network/resolver/dns"
	"github.com/lack-io/vine/lib/network/resolver/http"
	"github.com/lack-io/vine/lib/network/resolver/registry"
	"github.com/lack-io/vine/lib/network/tunnel"
	"github.com/lack-io/vine/lib/proxy"
	"github.com/lack-io/vine/lib/proxy/mucp"
	"github.com/lack-io/vine/util/helper"
	mux "github.com/lack-io/vine/util/muxer"
)

var (
	// Name of the network service
	Name = "go.vine.network"
	// Name of the vine network
	Network = "go.vine"
	// Address is the network address
	Address = ":8085"
	// Set the advertise address
	Advertise = ""
	// Resolver is the network resolver
	Resolver = "registry"
	// The tunnel token
	Token = "vine"
)

// Run runs the vine server
func Run(ctx *cli.Context, svcOpts ...vine.Option) {

	if len(ctx.String("server-name")) > 0 {
		Name = ctx.String("server-name")
	}
	if len(ctx.String("address")) > 0 {
		Address = ctx.String("address")
	}
	if len(ctx.String("advertise")) > 0 {
		Advertise = ctx.String("advertise")
	}
	if len(ctx.String("network")) > 0 {
		Network = ctx.String("network")
	}
	if len(ctx.String("token")) > 0 {
		Token = ctx.String("token")
	}

	var nodes []string
	if len(ctx.String("nodes")) > 0 {
		nodes = strings.Split(ctx.String("nodes"), ",")
	}
	if len(ctx.String("resolver")) > 0 {
		Resolver = ctx.String("resolver")
	}
	var res resolver.Resolver
	switch Resolver {
	case "dns":
		res = &dns.Resolver{}
	case "http":
		res = &http.Resolver{}
	case "registry":
		res = &registry.Resolver{}
	}

	// advertise the best routes
	strategy := router2.AdvertiseLocal
	if a := ctx.String("advertise-strategy"); len(a) > 0 {
		switch a {
		case "all":
			strategy = router2.AdvertiseAll
		case "best":
			strategy = router2.AdvertiseBest
		case "local":
			strategy = router2.AdvertiseLocal
		case "none":
			strategy = router2.AdvertiseNone
		}
	}

	// Initialise service
	svc := vine.NewService(
		vine.Name(Name),
		vine.RegisterTTL(time.Duration(ctx.Int("register-ttl"))*time.Second),
		vine.RegisterInterval(time.Duration(ctx.Int("register-interval"))*time.Second),
	)

	// create a tunnel
	tunOpts := []tunnel.Option{
		tunnel.Address(Address),
		tunnel.Token(Token),
	}

	if ctx.Bool("enable-tls") {
		config, err := helper.TLSConfig(ctx)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		config.InsecureSkipVerify = true

		tunOpts = append(tunOpts, tunnel.Transport(
			quic.NewTransport(transport.TLSConfig(config)),
		))
	}

	gateway := ctx.String("gateway")
	tun := tunnel.NewTunnel(tunOpts...)
	id := svc.Server().Options().Id

	// local tunnel router
	rtr := registry2.NewRouter(
		router2.Network(Network),
		router2.Id(id),
		router2.Registry(svc.Client().Options().Registry),
		router2.Advertise(strategy),
		router2.Gateway(gateway),
	)

	// create new network
	net := network.NewNetwork(
		network.Id(id),
		network.Name(Network),
		network.Address(Address),
		network.Advertise(Advertise),
		network.Nodes(nodes...),
		network.Tunnel(tun),
		network.Router(rtr),
		network.Resolver(res),
	)

	// local proxy
	prx := mucp.NewProxy(
		proxy.WithRouter(rtr),
		proxy.WithClient(svc.Client()),
		proxy.WithLink("network", net.Client()),
	)

	// create a handler
	h := server.NewHandler(
		&handler.Network{
			Network: net,
		},
	)

	// register the handler
	server.Handle(h)

	// create a new muxer
	mux := mux.New(Name, prx)

	// init server
	svc.Server().Init(
		server.WithRouter(mux),
	)

	// set network server to proxy
	net.Server().Init(
		server.WithRouter(mux),
	)

	// connect network
	if err := net.Connect(); err != nil {
		log.Errorf("Network failed to connect: %v", err)
		os.Exit(1)
	}

	// netClose hard exits if we have problems
	netClose := func(net network.Network) error {
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

	log.Infof("Network [%s] listening on %s", Network, Address)

	if err := svc.Run(); err != nil {
		log.Errorf("Network %s failed: %v", Network, err)
		netClose(net)
		os.Exit(1)
	}

	// close the network
	netClose(net)
}

func Commands(options ...vine.Option) []*cli.Command {
	command := &cli.Command{
		Name:  "network",
		Usage: "Run the vine network node",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "address",
				Usage:   "Set the vine network address :8085",
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
				Usage:   "Set the vine network name: go.vine",
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
			&cli.StringFlag{
				Name:    "resolver",
				Usage:   "Set the vine network resolver. This can be a comma separated list.",
				EnvVars: []string{"VINE_NETWORK_RESOLVER"},
			},
			&cli.StringFlag{
				Name:    "advertise-strategy",
				Usage:   "Set the route advertise strategy; all, best, local, none",
				EnvVars: []string{"VINE_NETWORK_ADVERTISE_STRATEGY"},
			},
		},
		Subcommands: append([]*cli.Command{
			{
				Name:        "api",
				Usage:       "Run the network api",
				Description: "Run the network api",
				Action: func(ctx *cli.Context) error {
					api.Run(ctx)
					return nil
				},
			},
			{
				Name:        "dns",
				Usage:       "Start a DNS resolver service that registers core nodes in DNS",
				Description: "Start a DNS resolver service that registers core nodes in DNS",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "provider",
						Usage:   "The DNS provider to use. Currently, only cloudflare is implemented",
						EnvVars: []string{"VINE_NETWORK_DNS_PROVIDER"},
						Value:   "cloudflare",
					},
					&cli.StringFlag{
						Name:    "api-token",
						Usage:   "The provider's API Token.",
						EnvVars: []string{"VINE_NETWORK_DNS_API_TOKEN"},
					},
					&cli.StringFlag{
						Name:    "zone-id",
						Usage:   "The provider's Zone ID.",
						EnvVars: []string{"VINE_NETWORK_DNS_ZONE_ID"},
					},
					&cli.StringFlag{
						Name:    "token",
						Usage:   "Shared secret that must be presented to the service to authorize requests.",
						EnvVars: []string{"VINE_NETWORK_DNS_TOKEN"},
					},
				},
				Action: func(ctx *cli.Context) error {
					if err := helper.UnexpectedSubcommand(ctx); err != nil {
						return err
					}
					netdns.Run(ctx)
					return nil
				},
				Subcommands: mcli.NetworkDNSCommands(),
			},
		}, mcli.NetworkCommands()...),
		Action: func(ctx *cli.Context) error {
			if err := helper.UnexpectedSubcommand(ctx); err != nil {
				return err
			}
			Run(ctx, options...)
			return nil
		},
	}

	return []*cli.Command{command}
}
