// MIT License
//
// Copyright (c) 2021 Lack
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

package template

var (
	SingleEntry = `package pkg

import (
	"github.com/vine-io/apimachinery/inject"
	"github.com/vine-io/apimachinery/server"
	"github.com/vine-io/vine"
	log "github.com/vine-io/vine/lib/logger"

	"{{.Dir}}/pkg/version"
)

func Run() {
	var err error

	srv := vine.NewService()
	opts := []vine.Option{
		vine.Name(version.{{title .Name}}Name),
		vine.Id(version.{{title .Name}}Id),
		vine.Version(version.GetVersion()),
		vine.Metadata(map[string]string{
			"namespace": version.Namespace,
		}),
	}

	srv.Init(opts...)

	if err = inject.Provide(srv, srv.Server(), srv.Client()); err != nil {
		log.Fatal(err)
	}

	if err = inject.Populate(); err != nil {
		log.Fatal(err)
	}

	for _, o := range inject.Objects() {
		if h, ok := o.Value.(server.Service); ok {
			if err = h.Register(srv.Server()); err != nil {
				log.Fatalf("register vine service: %v", err)
			}
			continue
		}

		if impl, ok := o.Value.(server.BizImpl); ok {
			if err = impl.Init(); err != nil {
				log.Fatalf("biz init: %v", o.Name, err)
			}

			if err = impl.Start(); err != nil {
				log.Fatalf("biz start: %v", o.Name, err)
			}
		}
	}

	if err = srv.Run(); err != nil {
		log.Fatalf("start server: %v", err)
	}
}`

	SimpleBuiltin = `package pkg

import (
	_ "{{.Dir}}/pkg/biz"
	_ "{{.Dir}}/pkg/service"
)
`

	ClusterEntry = `package {{.Name}}

import (
	"github.com/vine-io/apimachinery/inject"
	"github.com/vine-io/apimachinery/server"
	"github.com/vine-io/vine"
	log "github.com/vine-io/vine/lib/logger"

	"{{.Dir}}/pkg/version"
)

func Run() {
	var err error

	srv := vine.NewService()
	opts := []vine.Option{
		vine.Name(version.{{title .Name}}Name),
		vine.Id(version.{{title .Name}}Id),
		vine.Version(version.GetVersion()),
		vine.Metadata(map[string]string{
			"namespace": version.Namespace,
		}),
	}

	srv.Init(opts...)

	if err = inject.Provide(srv, srv.Server(), srv.Client()); err != nil {
		log.Fatal(err)
	}

	if err = inject.Populate(); err != nil {
		log.Fatal(err)
	}

	for _, o := range inject.Objects() {
		if h, ok := o.Value.(server.Service); ok {
			if err = h.Register(srv.Server()); err != nil {
				log.Fatalf("register vine service: %v", err)
			}
			continue
		}

		if impl, ok := o.Value.(server.BizImpl); ok {
			if err = impl.Init(); err != nil {
				log.Fatalf("biz init: %v", o.Name, err)
			}

			if err = impl.Start(); err != nil {
				log.Fatalf("biz start: %v", o.Name, err)
			}
		}
	}

	if err = srv.Run(); err != nil {
		log.Fatalf("start server: %v", err)
	}
}`

	ClusterBuiltin = `package {{.Name}}

import (
	_ "{{.Dir}}/pkg/{{.Name}}/biz"
	_ "{{.Dir}}/pkg/{{.Name}}/service"
)
`

	GatewayEntry = `package {{.Name}}

import (
	"github.com/gin-gonic/gin"
	"github.com/vine-io/cli"

	"github.com/vine-io/vine"
	ahandler "github.com/vine-io/vine/lib/api/handler"
	"github.com/vine-io/vine/lib/api/handler/openapi"
	arpc "github.com/vine-io/vine/lib/api/handler/rpc"
	"github.com/vine-io/vine/lib/api/resolver"
	"github.com/vine-io/vine/lib/api/resolver/grpc"
	"github.com/vine-io/vine/lib/api/router"
	regRouter "github.com/vine-io/vine/lib/api/router/registry"
	"github.com/vine-io/vine/lib/api/server"
	httpapi "github.com/vine-io/vine/lib/api/server/http"
	log "github.com/vine-io/vine/lib/logger"
	"github.com/vine-io/vine/util/helper"
	"github.com/vine-io/vine/util/namespace"

	"{{.Dir}}/pkg/runtime"
)

var (
	Address       = ":8080"
	Handler       = "rpc"
	Type          = "api"
	APIPath       = "/"
	enableOpenAPI = false

	flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "api-address",
			Usage:       "The specify for api address",
			EnvVars:     []string{"VINE_API_ADDRESS"},
			Required:    true,
			Value:       Address,
			Destination: &Address,
		},
		&cli.BoolFlag{
			Name:    "enable-openapi",
			Usage:   "Enable OpenAPI3",
			EnvVars: []string{"VINE_ENABLE_OPENAPI"},
			Value:   true,
		},
		&cli.BoolFlag{
			Name:    "enable-cors",
			Usage:   "Enable CORS, allowing the API to be called by frontend applications",
			EnvVars: []string{"VINE_API_ENABLE_CORS"},
			Value:   true,
		},
	}
)

func Run() {
	// Init API
	var opts []server.Option

	// initialise service
	svc := vine.NewService(
		vine.Name(runtime.{{title .Name}}Name),
		vine.Id(runtime.{{title .Name}}Id),
		vine.Version(runtime.GetVersion()),
		vine.Metadata(map[string]string{
			"api-address": Address,
			"namespace": runtime.Namespace,
		}),
		vine.Flags(flags...),
		vine.Action(func(ctx *cli.Context) error {
			enableOpenAPI = ctx.Bool("enable-openapi")

			if ctx.Bool("enable-tls") {
				config, err := helper.TLSConfig(ctx)
				if err != nil {
					log.Errorf(err.Error())
					return err
				}

				opts = append(opts, server.EnableTLS(true))
				opts = append(opts, server.TLSConfig(config))
			}
			return nil
		}),
	)

	svc.Init()

	opts = append(opts, server.EnableCORS(true))

	// create the router
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(gin.Recovery())

	if enableOpenAPI {
		openapi.RegisterOpenAPI(app)
	}

	// create the namespace resolver
	nsResolver := namespace.NewResolver(Type, runtime.Namespace)
	// resolver options
	ropts := []resolver.Option{
		resolver.WithNamespace(nsResolver.ResolveWithType),
		resolver.WithHandler(Handler),
	}

	log.Infof("Registering API RPC Handler at %s", APIPath)
	rr := grpc.NewResolver(ropts...)
	rt := regRouter.NewRouter(
		router.WithHandler(arpc.Handler),
		router.WithResolver(rr),
		router.WithRegistry(svc.Options().Registry),
	)
	rp := arpc.NewHandler(
		ahandler.WithNamespace(runtime.Namespace),
		ahandler.WithRouter(rt),
		ahandler.WithClient(svc.Client()),
	)
	app.Use(rp.Handle)

	api := httpapi.NewServer(Address)

	if err := api.Init(opts...); err != nil {
		log.Fatal(err)
    }
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
`
)
