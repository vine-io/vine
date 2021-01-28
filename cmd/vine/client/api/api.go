// Copyright 2021 lack
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

// Package api is an API Gateway
package api

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/go-acme/lego/v3/providers/dns/cloudflare"
	"github.com/gorilla/mux"
	"github.com/lack-io/cli"
	"github.com/rakyll/statik/fs"

	"github.com/lack-io/vine"
	"github.com/lack-io/vine/cmd/vine/client/api/auth"
	"github.com/lack-io/vine/cmd/vine/client/api/handler"
	rrvine "github.com/lack-io/vine/cmd/vine/client/resolver/api"
	"github.com/lack-io/vine/plugin"
	ahandler "github.com/lack-io/vine/service/api/handler"
	aapi "github.com/lack-io/vine/service/api/handler/api"
	"github.com/lack-io/vine/service/api/handler/event"
	ahttp "github.com/lack-io/vine/service/api/handler/http"
	arpc "github.com/lack-io/vine/service/api/handler/rpc"
	aweb "github.com/lack-io/vine/service/api/handler/web"
	"github.com/lack-io/vine/service/api/resolver"
	"github.com/lack-io/vine/service/api/resolver/grpc"
	"github.com/lack-io/vine/service/api/resolver/host"
	"github.com/lack-io/vine/service/api/resolver/path"
	"github.com/lack-io/vine/service/api/router"
	regRouter "github.com/lack-io/vine/service/api/router/registry"
	"github.com/lack-io/vine/service/api/server"
	"github.com/lack-io/vine/service/api/server/acme"
	"github.com/lack-io/vine/service/api/server/acme/autocert"
	"github.com/lack-io/vine/service/api/server/acme/certmagic"
	httpapi "github.com/lack-io/vine/service/api/server/http"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/sync/memory"
	"github.com/lack-io/vine/util/helper"
	"github.com/lack-io/vine/util/namespace"
	"github.com/lack-io/vine/util/openapi"
	"github.com/lack-io/vine/util/stats"

	_ "github.com/lack-io/vine/util/openapi/statik"
)

var (
	Name                  = "go.vine.api"
	Address               = ":8080"
	Handler               = "meta"
	Resolver              = "vine"
	RPCPath               = "/rpc"
	APIPath               = "/"
	ProxyPath             = "/{service:[a-zA-Z0-9]+}"
	Namespace             = "go.vine"
	Type                  = "api"
	HeaderPrefix          = "X-Vine-"
	EnableRPC             = false
	ACMEProvider          = "autocert"
	ACMEChallengeProvider = "cloudflare"
	ACMECA                = acme.LetsEncryptProductionCA
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
	if len(ctx.String("acme-provider")) > 0 {
		ACMEProvider = ctx.String("acme-provider")
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

	// Init plugins
	for _, p := range Plugins() {
		p.Init(ctx)
	}

	// Init API
	var opts []server.Option

	if ctx.Bool("enable-acme") {
		hosts := helper.ACMEHosts(ctx)
		opts = append(opts, server.EnableACME(true))
		opts = append(opts, server.ACMEHosts(hosts...))
		switch ACMEProvider {
		case "autocert":
			opts = append(opts, server.ACMEProvider(autocert.NewProvider()))
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
				svc.Options().Store,
			)

			config := cloudflare.NewDefaultConfig()
			config.AuthToken = apiToken
			config.ZoneToken = apiToken
			challengeProvider, err := cloudflare.NewDNSProviderConfig(config)
			if err != nil {
				log.Fatal(err.Error())
			}

			opts = append(opts,
				server.ACMEProvider(
					certmagic.NewProvider(
						acme.AcceptToS(true),
						acme.CA(ACMECA),
						acme.Cache(storage),
						acme.ChallengeProvider(challengeProvider),
						acme.OnDemand(false),
					),
				),
			)
		default:
			log.Fatalf("%s is not a valid ACME provider\n", ACMEProvider)
		}
	} else if ctx.Bool("enable-tls") {
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
	var h http.Handler
	r := mux.NewRouter()
	h = r

	if ctx.Bool("enable-stats") {
		st := stats.New()
		r.HandleFunc("/stats", st.StatsHandler)
		h = st.ServeHTTP(r)
		st.Start()
		defer st.Stop()
	}

	if ctx.Bool("enable-openapi") {
		openAPI := openapi.New(svc)
		mime.AddExtensionType(".svg", "image/svg+xml")
		statikFs, err := fs.New()
		if err != nil {
			log.Fatalf("Starting OpenAPI: %v", err)
		}
		prefix := "/openapi-ui/"
		fileServer := http.FileServer(statikFs)
		r.HandleFunc(prefix, openAPI.OpenAPIHandler)
		r.PathPrefix(prefix).Handler(http.StripPrefix(prefix, fileServer))
		r.HandleFunc("/openapi.json", openAPI.OpenAPIJOSNHandler)
		log.Infof("Starting OpenAPI at %v", prefix)
		h = openAPI.ServeHTTP(r)
	}

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprintf(`{"version": "%s"}`, ctx.App.Version)
		w.Write([]byte(response))
	})

	// strip favicon.ico
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {})

	// register rpc handler
	if EnableRPC {
		log.Infof("Registering RPC Handler at %s", RPCPath)
		r.HandleFunc(RPCPath, handler.RPC)
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
		r.PathPrefix(APIPath).Handler(rp)
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
		r.PathPrefix(APIPath).Handler(ap)
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
		r.PathPrefix(APIPath).Handler(ev)
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
		r.PathPrefix(ProxyPath).Handler(ht)
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
		r.PathPrefix(APIPath).Handler(w)
	default:
		log.Infof("Registering API Default Handler at %s", APIPath)
		rt := regRouter.NewRouter(
			router.WithResolver(rr),
			router.WithRegistry(svc.Options().Registry),
		)
		r.PathPrefix(APIPath).Handler(handler.Meta(svc, rt, nsResolver.ResolveWithType))
	}

	// reverse wrap handler
	plugins := append(Plugins(), plugin.Plugins()...)
	for i := len(plugins); i > 0; i-- {
		h = plugins[i-1].Handler()(h)
	}

	// create the auth wrapper and the server
	authWrapper := auth.Wrapper(rr, nsResolver)
	api := httpapi.NewServer(Address, server.WrapHandler(authWrapper))

	api.Init(opts...)
	api.Handle("/", h)

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
