// Copyright 2020 lack
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

// Package proxy is a cli proxy
package proxy

import (
	"os"
	"strings"

	"github.com/go-acme/lego/v3/providers/dns/cloudflare"
	"github.com/lack-io/cli"

	"github.com/lack-io/vine/service"
	"github.com/lack-io/vine/service/api/server/acme"
	"github.com/lack-io/vine/service/api/server/acme/autocert"
	"github.com/lack-io/vine/service/api/server/acme/certmagic"
	"github.com/lack-io/vine/service/auth"
	bmem "github.com/lack-io/vine/service/broker/memory"
	"github.com/lack-io/vine/service/client"
	mucpCli "github.com/lack-io/vine/service/client/mucp"
	"github.com/lack-io/vine/service/config/cmd"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/proxy"
	"github.com/lack-io/vine/service/proxy/http"
	"github.com/lack-io/vine/service/proxy/mucp"
	"github.com/lack-io/vine/service/registry"
	rmem "github.com/lack-io/vine/service/registry/memory"
	"github.com/lack-io/vine/service/router"
	rs "github.com/lack-io/vine/service/router/grpc"
	regRouter "github.com/lack-io/vine/service/router/registry"
	"github.com/lack-io/vine/service/server"
	sgrpc "github.com/lack-io/vine/service/server/grpc"
	mucpServer "github.com/lack-io/vine/service/server/mucp"
	"github.com/lack-io/vine/service/sync/memory"
	"github.com/lack-io/vine/util/helper"
	"github.com/lack-io/vine/util/muxer"
	"github.com/lack-io/vine/util/wrapper"
)

var (
	// Name of the proxy
	Name = "go.vine.proxy"
	// The address of the proxy
	Address = ":8081"
	// the proxy protocol
	Protocol = "grpc"
	// The endpoint host to route to
	Endpoint string
	// ACME (Cert management)
	ACMEProvider          = "autocert"
	ACMEChallengeProvider = "cloudflare"
	ACMECA                = acme.LetsEncryptProductionCA
)

func Run(ctx *cli.Context, srvOpts ...service.Option) {

	// because VINE_PROXY_ADDRESS is used internally by the vine/client
	// we need to unset it so we don't end up calling ourselves infinitely
	os.Unsetenv("VINE_PROXY_ADDRESS")

	if len(ctx.String("server-name")) > 0 {
		Name = ctx.String("server-name")
	}
	if len(ctx.String("address")) > 0 {
		Address = ctx.String("address")
	}
	if len(ctx.String("endpoint")) > 0 {
		Endpoint = ctx.String("endpoint")
	}
	if len(ctx.String("protocol")) > 0 {
		Protocol = ctx.String("protocol")
	}
	if len(ctx.String("acme-provider")) > 0 {
		ACMEProvider = ctx.String("acme-provider")
	}

	// Init plugins
	for _, p := range Plugins() {
		p.Init(ctx)
	}

	// service opts
	srvOpts = append(srvOpts, service.Name(Name))

	// new service
	srv := service.NewService(srvOpts...)

	// set the context
	var popts []proxy.Option

	// create new router
	var r router.Router

	routerName := ctx.String("router")
	routerAddr := ctx.String("router-address")

	ropts := []router.Option{
		router.Id(server.DefaultId),
		router.Client(client.DefaultClient),
		router.Address(routerAddr),
		router.Registry(registry.DefaultRegistry),
	}

	// check if we need to use the router service
	switch {
	case routerName == "go.vine.router":
		r = rs.NewRouter(ropts...)
	case routerName == "service":
		r = rs.NewRouter(ropts...)
	case len(routerAddr) > 0:
		r = rs.NewRouter(ropts...)
	default:
		r = regRouter.NewRouter(ropts...)
	}

	// start the router
	if err := r.Start(); err != nil {
		log.Errorf("Proxy error starting router: %s", err)
		os.Exit(1)
	}

	// append router to proxy opts
	popts = append(popts, proxy.WithRouter(r))

	// new proxy
	var p proxy.Proxy
	// setup the default server
	var ss server.Server

	// set endpoint
	if len(Endpoint) > 0 {
		switch {
		case strings.HasPrefix(Endpoint, "grpc://"):
			ep := strings.TrimPrefix(Endpoint, "grpc://")
			popts = append(popts, proxy.WithEndpoint(ep))
			Protocol = "grpc"
		case strings.HasPrefix(Endpoint, "http://"):
			// TODO: strip prefix?
			popts = append(popts, proxy.WithEndpoint(Endpoint))
			Protocol = "http"
		default:
			// TODO: strip prefix?
			popts = append(popts, proxy.WithEndpoint(Endpoint))
			Protocol = "mucp"
		}
	}

	serverOpts := []server.Option{
		server.Address(Address),
		server.Registry(rmem.NewRegistry()),
		server.Broker(bmem.NewBroker()),
	}

	// enable acme will create a net.Listener which
	if ctx.Bool("enable-acme") {
		var ap acme.Provider

		switch ACMEProvider {
		case "autocert":
			ap = autocert.NewProvider()
		case "certmagic":
			if ACMEChallengeProvider != "cloudflare" {
				log.Fatal("The only implemented DNS challenge provider is cloudflare")
			}

			apiToken := os.Getenv("CF_API_TOKEN")
			if len(apiToken) == 0 {
				log.Fatal("env variables CF_API_TOKEN and CF_ACCOUNT_ID must be set")
			}

			storage := certmagic.NewStorage(
				memory.NewSync(),
				srv.Options().Store,
			)

			config := cloudflare.NewDefaultConfig()
			config.AuthToken = apiToken
			config.ZoneToken = apiToken
			challengeProvider, err := cloudflare.NewDNSProviderConfig(config)
			if err != nil {
				log.Fatal(err.Error())
			}

			// define the provider
			ap = certmagic.NewProvider(
				acme.AcceptToS(true),
				acme.CA(ACMECA),
				acme.Cache(storage),
				acme.ChallengeProvider(challengeProvider),
				acme.OnDemand(false),
			)
		default:
			log.Fatalf("Unsupported acme provider: %s\n", ACMEProvider)
		}

		// generate the tls config
		config, err := ap.TLSConfig(helper.ACMEHosts(ctx)...)
		if err != nil {
			log.Fatalf("Failed to generate acme tls config: %v", err)
		}

		// set the tls config
		serverOpts = append(serverOpts, server.TLSConfig(config))
		// enable tls will leverage tls certs and generate a tls.Config
	} else if ctx.Bool("enable-tls") {
		// get certificates from the context
		config, err := helper.TLSConfig(ctx)
		if err != nil {
			log.Fatal(err)
			return
		}
		serverOpts = append(serverOpts, server.TLSConfig(config))
	}

	// add auth wrapper to server
	var authOpts []auth.Option
	if ctx.IsSet("auth-public-key") {
		authOpts = append(authOpts, auth.PublicKey(ctx.String("auth-public-key")))
	}
	if ctx.IsSet("auth-private-key") {
		authOpts = append(authOpts, auth.PublicKey(ctx.String("auth-private-key")))
	}

	a := *cmd.DefaultOptions().Auth
	a.Init(authOpts...)
	authFn := func() auth.Auth { return a }
	authOpt := server.WrapHandler(wrapper.AuthHandler(authFn))
	serverOpts = append(serverOpts, authOpt)

	// set proxy
	switch Protocol {
	case "http":
		p = http.NewProxy(popts...)
		serverOpts = append(serverOpts, server.WithRouter(p))
		// TODO: http server
		ss = mucpServer.NewServer(serverOpts...)
	case "mucp":
		popts = append(popts, proxy.WithClient(mucpCli.NewClient()))
		p = mucp.NewProxy(popts...)

		serverOpts = append(serverOpts, server.WithRouter(p))
		ss = mucpServer.NewServer(serverOpts...)
	default:
		p = mucp.NewProxy(popts...)

		serverOpts = append(serverOpts, server.WithRouter(p))
		ss = sgrpc.NewServer(serverOpts...)
	}

	if len(Endpoint) > 0 {
		log.Infof("Proxy [%s] serving endpoint: %s", p.String(), Endpoint)
	} else {
		log.Infof("Proxy [%s] serving protocol: %s", p.String(), Protocol)
	}

	// create a new proxy muxer which includes the debug handler
	muxer := muxer.New(Name, p)

	// set the router
	srv.Server().Init(
		server.WithRouter(muxer),
	)

	// Start the proxy server
	if err := ss.Start(); err != nil {
		log.Fatal(err)
	}

	// Run internal service
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}

	// Stop the server
	if err := ss.Stop(); err != nil {
		log.Fatal(err)
	}
}

func Commands(options ...service.Option) []*cli.Command {
	command := &cli.Command{
		Name:  "proxy",
		Usage: "Run the service proxy",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "router",
				Usage:   "Set the router to use e.g default, go.vine.router",
				EnvVars: []string{"VINE_ROUTER"},
			},
			&cli.StringFlag{
				Name:    "router-address",
				Usage:   "Set the router address",
				EnvVars: []string{"VINE_ROUTER_ADDRESS"},
			},
			&cli.StringFlag{
				Name:    "address",
				Usage:   "Set the proxy http address e.g 0.0.0.0:8081",
				EnvVars: []string{"VINE_PROXY_ADDRESS"},
			},
			&cli.StringFlag{
				Name:    "protocol",
				Usage:   "Set the protocol used for proxying e.g mucp, grpc, http",
				EnvVars: []string{"VINE_PROXY_PROTOCOL"},
			},
			&cli.StringFlag{
				Name:    "endpoint",
				Usage:   "Set the endpoint to route to e.g greeter or localhost:9090",
				EnvVars: []string{"VINE_PROXY_ENDPOINT"},
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
