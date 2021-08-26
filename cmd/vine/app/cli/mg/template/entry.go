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
	log "github.com/vine-io/vine/lib/logger"

	"{{.Dir}}/pkg/interfaces"
)

func Run() {
	s := interfaces.New()

	if err := s.Init(); err != nil {
		log.Fatal(err)
	}

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}`

	ClusterEntry = `package {{.Name}}

import (
	log "github.com/vine-io/vine/lib/logger"

	"{{.Dir}}/pkg/{{.Name}}/interfaces"
)

func Run() {
	app := interfaces.New()

	if err := app.Init(); err != nil {
		log.Fatal(err)
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}`

	GatewayEntry = `package {{.Name}}

import (
	"github.com/gofiber/fiber/v2"
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
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

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
	app.Group(APIPath, rp.Handle)

	api := httpapi.NewServer(Address)

	if err := api.Init(opts...); err != nil {
		log.Fatal(err)
    }
	api.Handle("/", app)

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
