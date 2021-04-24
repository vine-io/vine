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
	SingleMod = `package pkg

var (
	Name = "{{.Alias}}"
	Namespace = "{{.Namespace}}"
	Version = "latest"

	GitTag    string
	GitCommit string
	BuildDate string
	Version = GitTag + "-" + GitCommit + "-" + BuildDate
)
`

	ClusterMod = `package {{.Name}}

var (
	Name = "{{.Alias}}"
	Namespace = "{{.Namespace}}"
	Version = "latest"

	GitTag    string
	GitCommit string
	BuildDate string
	Version = GitTag + "-" + GitCommit + "-" + BuildDate
)
`

	SinglePlugin = `package pkg
{{if .Plugins}}
import ({{range .Plugins}}
	_ "github.com/lack-io/plugins/{{.}}"{{end}}
){{end}}
`

	ClusterPlugin = `package {{.Name}}
{{if .Plugins}}
import ({{range .Plugins}}
	_ "github.com/lack-io/plugins/{{.}}"{{end}}
){{end}}
`

	SingleDefault = `package pkg

func init() {
	// TODO: setup default lib
}
`

	ClusterDefault = `package {{.Name}}

func init() {
	// TODO: setup default lib
}
`

	SingleApp = `package pkg

import (
	"github.com/lack-io/vine"
	log "github.com/lack-io/vine/lib/logger"

	"{{.Dir}}/pkg/server"
)

func Run() {
	s := server.New(
		vine.Name(Name),
		vine.Version(Version),
	)

	if err := s.Init(); err != nil {
		log.Fatal(err)
	}

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}`

	ClusterApp = `package {{.Name}}

import (
	"github.com/lack-io/vine"
	log "github.com/lack-io/vine/lib/logger"

	"{{.Dir}}/pkg/{{.Name}}/server"
)

func Run() {
	s := server.New(
		vine.Name(Name),
		vine.Version(Version),
	)

	if err := s.Init(); err != nil {
		log.Fatal(err)
	}

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}`

	GatewayApp = `package {{.Name}}


import (
	"mime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/lack-io/cli"

	"github.com/lack-io/vine"
	"github.com/lack-io/vine/cmd/vine/app/api/auth"
	ahandler "github.com/lack-io/vine/lib/api/handler"
	"github.com/lack-io/vine/lib/api/handler/openapi"
	arpc "github.com/lack-io/vine/lib/api/handler/rpc"
	"github.com/lack-io/vine/lib/api/resolver"
	"github.com/lack-io/vine/lib/api/resolver/grpc"
	"github.com/lack-io/vine/lib/api/router"
	regRouter "github.com/lack-io/vine/lib/api/router/registry"
	"github.com/lack-io/vine/lib/api/server"
	httpapi "github.com/lack-io/vine/lib/api/server/http"
	log "github.com/lack-io/vine/lib/logger"
	"github.com/lack-io/vine/util/helper"
	"github.com/lack-io/vine/util/namespace"
	"github.com/rakyll/statik/fs"

	_ "github.com/lack-io/vine/lib/api/handler/openapi/statik"
)

var (
	Address       = ":8080"
	Handler       = "rpc"
	Type          = "api"
	APIPath       = "/"
	enableOpenAPI = false

	flags = []cli.Flag{
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
		vine.Name(Name),
		vine.Metadata(map[string]string{"api-address": Address}),
		vine.Flags(flags...),
		vine.Action(func(ctx *cli.Context) error {
			if len(ctx.String("server-name")) > 0 {
				Name = ctx.String("server-name")
			}
			if len(ctx.String("server-address")) > 0 {
				Address = ctx.String("server-address")
			}
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
		openAPI := openapi.New(svc)
		mime.AddExtensionType(".svg", "image/svg+xml")
		statikFs, err := fs.New()
		if err != nil {
			log.Fatalf("Starting OpenAPI: %v", err)
		}
		prefix := "/openapi-ui/"
		app.All(prefix, openAPI.OpenAPIHandler)
		app.Use(prefix, filesystem.New(filesystem.Config{Root: statikFs}))
		app.Get("/openapi.json", openAPI.OpenAPIJOSNHandler)
		app.Get("/services", openAPI.OpenAPIServiceHandler)
		log.Infof("Starting OpenAPI at %v", prefix)
	}

	// create the namespace resolver
	nsResolver := namespace.NewResolver(Type, Namespace)
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
		ahandler.WithNamespace(Namespace),
		ahandler.WithRouter(rt),
		ahandler.WithClient(svc.Client()),
	)
	app.Group(APIPath, rp.Handle)

	// create the auth wrapper and the server
	// TODO: app middleware
	authWrapper := auth.Wrapper(rr, nsResolver)
	api := httpapi.NewServer(Address, server.WrapHandler(authWrapper))

	api.Init(opts...)
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

	WebSRV = `package {{.Name}}

import (
	"github.com/gofiber/fiber/v2"

	"github.com/lack-io/vine/lib/web"
	log "github.com/lack-io/vine/lib/logger"
)

func Run() {
	srv := web.NewService(
		web.Name("go.vine.web.helloworld"),
	)

	//service.Handle("/", http.RedirectHandler("/index.html", 301))
	srv.Handle(web.MethodGet, "/", func(c *fiber.Ctx) error {
		return c.SendString("hello world")
	})

	if err := service.Init(); err != nil {
		log.Fatal(err)
	}

	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}`

	DaoHandler = `package dao`
)
