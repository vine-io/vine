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

// Package proxy is a cli proxy
package proxy

import (
	"os"
	"strings"

	"github.com/lack-io/cli"
	"github.com/lack-io/vine"
	"github.com/lack-io/vine/core/broker/memory"
	"github.com/lack-io/vine/core/client"
	"github.com/lack-io/vine/core/client/mucp"
	"github.com/lack-io/vine/core/registry"
	regMemory "github.com/lack-io/vine/core/registry/memory"
	rr "github.com/lack-io/vine/core/router"
	"github.com/lack-io/vine/core/router/grpc"
	rreg "github.com/lack-io/vine/core/router/registry"
	"github.com/lack-io/vine/core/server"
	sgrpc "github.com/lack-io/vine/core/server/grpc"
	mucpServer "github.com/lack-io/vine/core/server/mucp"
	log "github.com/lack-io/vine/lib/logger"
	"github.com/lack-io/vine/lib/proxy"
	"github.com/lack-io/vine/lib/proxy/http"
	pmucp "github.com/lack-io/vine/lib/proxy/mucp"
	"github.com/lack-io/vine/util/helper"
	"github.com/lack-io/vine/util/muxer"
)

var (
	// Name of the proxy
	Name = "go.vine.proxy"
	// Address the address of the proxy
	Address = ":8081"
	// Protocol the proxy protocol
	Protocol = "grpc"
	// Endpoint the endpoint host to route to
	Endpoint string
)

func Run(ctx *cli.Context, svcOpts ...vine.Option) {

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
	// service opts
	svcOpts = append(svcOpts, vine.Name(Name))

	// new service
	svc := vine.NewService(svcOpts...)

	// set the context
	var popts []proxy.Option

	// create new router
	var r rr.Router

	routerName := ctx.String("router")
	routerAddr := ctx.String("router-address")

	ropts := []rr.Option{
		rr.Id(server.DefaultId),
		rr.Client(client.DefaultClient),
		rr.Address(routerAddr),
		rr.Registry(registry.DefaultRegistry),
	}

	// check if we need to use the router service
	switch {
	case routerName == "go.vine.router":
		r = grpc.NewRouter(ropts...)
	case routerName == "service":
		r = grpc.NewRouter(ropts...)
	case len(routerAddr) > 0:
		r = grpc.NewRouter(ropts...)
	default:
		r = rreg.NewRouter(ropts...)
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
		server.Registry(regMemory.NewRegistry()),
		server.Broker(memory.NewBroker()),
	}

	// enable acme will create a net.Listener which
	if ctx.Bool("enable-tls") {
		// get certificates from the context
		config, err := helper.TLSConfig(ctx)
		if err != nil {
			log.Fatal(err)
			return
		}
		serverOpts = append(serverOpts, server.TLSConfig(config))
	}

	// set proxy
	switch Protocol {
	case "http":
		p = http.NewProxy(popts...)
		serverOpts = append(serverOpts, server.WithRouter(p))
		// TODO: http server
		ss = mucpServer.NewServer(serverOpts...)
	case "mucp":
		popts = append(popts, proxy.WithClient(mucp.NewClient()))
		p = pmucp.NewProxy(popts...)
		serverOpts = append(serverOpts, server.WithRouter(p))
		ss = mucpServer.NewServer(serverOpts...)
	default:
		p = pmucp.NewProxy(popts...)
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
	svc.Server().Init(
		server.WithRouter(muxer),
	)

	// Start the proxy server
	if err := ss.Start(); err != nil {
		log.Fatal(err)
	}

	// Run internal service
	if err := svc.Run(); err != nil {
		log.Fatal(err)
	}

	// Stop the server
	if err := ss.Stop(); err != nil {
		log.Fatal(err)
	}
}

func Commands(options ...vine.Option) []*cli.Command {
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

	return []*cli.Command{command}
}
