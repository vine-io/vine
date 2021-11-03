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

// Package api is an API Gateway
package api

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/vine-io/cli"
	"github.com/vine-io/vine"
	"github.com/vine-io/vine/cmd/vine/app/api/handler"
	rrvine "github.com/vine-io/vine/cmd/vine/client/resolver/api"
	ahandler "github.com/vine-io/vine/lib/api/handler"
	aapi "github.com/vine-io/vine/lib/api/handler/api"
	"github.com/vine-io/vine/lib/api/handler/event"
	ahttp "github.com/vine-io/vine/lib/api/handler/http"
	"github.com/vine-io/vine/lib/api/handler/openapi"
	arpc "github.com/vine-io/vine/lib/api/handler/rpc"
	aweb "github.com/vine-io/vine/lib/api/handler/web"
	"github.com/vine-io/vine/lib/api/resolver"
	"github.com/vine-io/vine/lib/api/resolver/grpc"
	"github.com/vine-io/vine/lib/api/resolver/host"
	"github.com/vine-io/vine/lib/api/resolver/path"
	"github.com/vine-io/vine/lib/api/router"
	regRouter "github.com/vine-io/vine/lib/api/router/registry"
	"github.com/vine-io/vine/lib/api/server"
	httpapi "github.com/vine-io/vine/lib/api/server/http"
	log "github.com/vine-io/vine/lib/logger"
	"github.com/vine-io/vine/util/helper"
	"github.com/vine-io/vine/util/namespace"
	"github.com/vine-io/vine/util/stats"
)

var (
	Name         = "go.vine.api"
	Address      = ":8080"
	Handler      = "meta"
	Resolver     = "vine"
	RPCPath      = "/rpc"
	APIPath      = "/"
	ProxyPath    = "/{service:[a-zA-Z0-9]+}"
	Namespace    = "go.vine"
	Type         = "api"
	HeaderPrefix = "X-Vine-"
	EnableRPC    = false
)

func Run(ctx *cli.Context, svcOpts ...vine.Option) {

	if len(ctx.String("server-name")) > 0 {
		Name = ctx.String("server-name")
	}
	if len(ctx.String("address")) > 0 {
		Address = ctx.String("address")
	}
	if len(ctx.String("handler")) > 0 {
		Handler = ctx.String("handler")
	}
	if len(ctx.String("resolver")) > 0 {
		Resolver = ctx.String("resolver")
	}
	if len(ctx.String("enable-rpc")) > 0 {
		EnableRPC = ctx.Bool("enable-rpc")
	}
	if len(ctx.String("type")) > 0 {
		Type = ctx.String("type")
	}
	if len(ctx.String("namespace")) > 0 {
		// remove the service type from the namespace to allow for
		// backwards compatability
		Namespace = strings.TrimSuffix(ctx.String("namespace"), "."+Type)
	}

	// apiNamespace has the format: "go.vine.api"
	apiNamespace := Namespace + "." + Type

	// append name to opts
	svcOpts = append(
		svcOpts,
		vine.Name(Name),
		vine.Metadata(map[string]string{"api-address": Address}),
	)

	// initialise service
	svc := vine.NewService(svcOpts...)

	// Init API
	var opts []server.Option

	if ctx.Bool("enable-tls") {
		config, err := helper.TLSConfig(ctx)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		opts = append(opts, server.EnableTLS(true))
		opts = append(opts, server.TLSConfig(config))
	}

	if ctx.Bool("enable-cors") {
		opts = append(opts, server.EnableCORS(true))
	}

	// create the router
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	if ctx.Bool("enable-stats") {
		st := stats.New()
		app.Any("/stats", st.StatsHandler)
		st.Start()
		defer st.Stop()
	}

	if ctx.Bool("enable-openapi") {
		openapi.RegisterOpenAPI(app)
	}

	app.GET(APIPath, func(c *gin.Context) {
		c.JSON(200, gin.H{"version": ctx.App.Version})
		return
	})

	// strip favicon.ico
	app.GET("/favicon.ico", func(ctx *gin.Context) { return })

	// register rpc handler
	if EnableRPC {
		log.Infof("Registering RPC Handler at %s", RPCPath)
		app.Use(handler.RPC)
		return
	}

	// create the namespace resolver
	nsResolver := namespace.NewResolver(Type, Namespace)

	// resolver options
	ropts := []resolver.Option{
		resolver.WithNamespace(nsResolver.ResolveWithType),
		resolver.WithHandler(Handler),
	}

	// default resolver
	rr := rrvine.NewResolver(ropts...)

	switch Resolver {
	case "host":
		rr = host.NewResolver(ropts...)
	case "path":
		rr = path.NewResolver(ropts...)
	case "grpc":
		rr = grpc.NewResolver(ropts...)
	}

	switch Handler {
	case "rpc":
		log.Infof("Registering API RPC Handler at %s", APIPath)
		rt := regRouter.NewRouter(
			router.WithHandler(arpc.Handler),
			router.WithResolver(rr),
			router.WithRegistry(svc.Options().Registry),
		)
		rp := arpc.NewHandler(
			ahandler.WithNamespace(apiNamespace),
			ahandler.WithRouter(rt),
			ahandler.WithClient(svc.Client()),
		)
		app.Use(rp.Handle)
	case "api":
		log.Infof("Registering API Request Handler at %s", APIPath)
		rt := regRouter.NewRouter(
			router.WithHandler(aapi.Handler),
			router.WithResolver(rr),
			router.WithRegistry(svc.Options().Registry),
		)
		ap := aapi.NewHandler(
			ahandler.WithNamespace(apiNamespace),
			ahandler.WithRouter(rt),
			ahandler.WithClient(svc.Client()),
		)
		app.Use(ap.Handle)
	case "event":
		log.Infof("Registering API Event Handler at %s", APIPath)
		rt := regRouter.NewRouter(
			router.WithHandler(event.Handler),
			router.WithResolver(rr),
			router.WithRegistry(svc.Options().Registry),
		)
		ev := event.NewHandler(
			ahandler.WithNamespace(apiNamespace),
			ahandler.WithRouter(rt),
			ahandler.WithClient(svc.Client()),
		)
		app.Use(ev.Handle)
	case "http", "proxy":
		log.Infof("Registering API HTTP Handler at %s", ProxyPath)
		rt := regRouter.NewRouter(
			router.WithHandler(ahttp.Handler),
			router.WithResolver(rr),
			router.WithRegistry(svc.Options().Registry),
		)
		ht := ahttp.NewHandler(
			ahandler.WithNamespace(apiNamespace),
			ahandler.WithRouter(rt),
			ahandler.WithClient(svc.Client()),
		)
		app.Group(ProxyPath, ht.Handle)
	case "web":
		log.Infof("Registering API Web Handler at %s", APIPath)
		rt := regRouter.NewRouter(
			router.WithHandler(aweb.Handler),
			router.WithResolver(rr),
			router.WithRegistry(svc.Options().Registry),
		)
		w := aweb.NewHandler(
			ahandler.WithNamespace(apiNamespace),
			ahandler.WithRouter(rt),
			ahandler.WithClient(svc.Client()),
		)
		app.Group(ProxyPath, w.Handle)
	default:
		log.Infof("Registering API Default Handler at %s", APIPath)
		rt := regRouter.NewRouter(
			router.WithResolver(rr),
			router.WithRegistry(svc.Options().Registry),
		)
		app.Group(ProxyPath, handler.Meta(svc, rt, nsResolver.ResolveWithType).Handle)
	}

	// create the auth wrapper and the server
	// TODO: app middleware
	api := httpapi.NewServer(Address)

	api.Init(opts...)
	api.Handle(APIPath, app)

	// Start API
	if err := api.Start(); err != nil {
		log.Fatal(err)
	}

	// Run server
	if err := svc.Run(); err != nil {
		log.Fatal(err)
	}

	// Stop API
	if err := api.Stop(); err != nil {
		log.Fatal(err)
	}
}

func Commands(options ...vine.Option) []*cli.Command {
	command := &cli.Command{
		Name:  "api",
		Usage: "Run the api gateway",
		Action: func(ctx *cli.Context) error {
			Run(ctx, options...)
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "address",
				Usage:   "Set the api address e.g 0.0.0.0:8080",
				EnvVars: []string{"VINE_API_ADDRESS"},
			},
			&cli.StringFlag{
				Name:    "handler",
				Usage:   "Specify the request handler to be used for mapping HTTP requests to services; {api, event, http, rpc}",
				EnvVars: []string{"VINE_API_HANDLER"},
			},
			&cli.StringFlag{
				Name:    "namespace",
				Usage:   "Set the namespace used by the API e.g. com.example",
				EnvVars: []string{"VINE_API_NAMESPACE"},
			},
			&cli.StringFlag{
				Name:    "type",
				Usage:   "Set the service type used by the API e.g. api",
				EnvVars: []string{"VINE_API_TYPE"},
			},
			&cli.StringFlag{
				Name:    "resolver",
				Usage:   "Set the hostname resolver used by the API {host, path, grpc}",
				EnvVars: []string{"VINE_API_RESOLVER"},
			},
			&cli.BoolFlag{
				Name:    "enable-openapi",
				Usage:   "Enable OpenAPI3",
				EnvVars: []string{"VINE_ENABLE_OPENAPI"},
			},
			&cli.BoolFlag{
				Name:    "enable-rpc",
				Usage:   "Enable call the backend directly via /rpc",
				EnvVars: []string{"VINE_API_ENABLE_RPC"},
			},
			&cli.BoolFlag{
				Name:    "enable-cors",
				Usage:   "Enable CORS, allowing the API to be called by frontend applications",
				EnvVars: []string{"VINE_API_ENABLE_CORS"},
				Value:   true,
			},
		},
	}

	return []*cli.Command{command}
}
