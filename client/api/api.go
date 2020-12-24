// Package api is an API Gateway
package api

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-acme/lego/v3/providers/dns/cloudflare"
	"github.com/gorilla/mux"
	"github.com/lack-io/cli"

	"github.com/lack-io/vine/client/api/auth"
	"github.com/lack-io/vine/plugin"
	"github.com/lack-io/vine/service"
	log "github.com/lack-io/vine/service/logger"
	ahandler "github.com/lack-io/vine/util/api/handler"
	aapi "github.com/lack-io/vine/util/api/handler/api"
	"github.com/lack-io/vine/util/api/handler/event"
	ahttp "github.com/lack-io/vine/util/api/handler/http"
	arpc "github.com/lack-io/vine/util/api/handler/rpc"
	aweb "github.com/lack-io/vine/util/api/handler/web"
	"github.com/lack-io/vine/util/api/resolver"
	"github.com/lack-io/vine/util/api/resolver/grpc"
	"github.com/lack-io/vine/util/api/resolver/host"
	"github.com/lack-io/vine/util/api/resolver/path"
	"github.com/lack-io/vine/util/api/router"
	regRouter "github.com/lack-io/vine/util/api/router/registry"
	"github.com/lack-io/vine/util/api/server"
	"github.com/lack-io/vine/util/api/server/acme"
	"github.com/lack-io/vine/util/api/server/acme/autocert"
	"github.com/lack-io/vine/util/api/server/acme/certmagic"
	httpapi "github.com/lack-io/vine/util/api/server/http"
	"github.com/lack-io/vine/util/handler"
	"github.com/lack-io/vine/util/helper"
	"github.com/lack-io/vine/util/namespace"
	rrvine "github.com/lack-io/vine/util/resolver/api"
	"github.com/lack-io/vine/util/stats"
	"github.com/lack-io/vine/util/sync/memory"
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

func Run(ctx *cli.Context, srvOpts ...service.Option) {

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
	srvOpts = append(srvOpts, service.Name(Name))

	// initialise service
	srv := service.NewService(srvOpts...)

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
				srv.Options().Store,
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
			fmt.Println(err.Error())
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

	// return version and list of services
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			return
		}

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
			router.WithRegistry(srv.Options().Registry),
		)
		rp := arpc.NewHandler(
			ahandler.WithNamespace(apiNamespace),
			ahandler.WithRouter(rt),
			ahandler.WithClient(srv.Client()),
		)
		r.PathPrefix(APIPath).Handler(rp)
	case "api":
		log.Infof("Registering API Request Handler at %s", APIPath)
		rt := regRouter.NewRouter(
			router.WithHandler(aapi.Handler),
			router.WithResolver(rr),
			router.WithRegistry(srv.Options().Registry),
		)
		ap := aapi.NewHandler(
			ahandler.WithNamespace(apiNamespace),
			ahandler.WithRouter(rt),
			ahandler.WithClient(srv.Client()),
		)
		r.PathPrefix(APIPath).Handler(ap)
	case "event":
		log.Infof("Registering API Event Handler at %s", APIPath)
		rt := regRouter.NewRouter(
			router.WithHandler(event.Handler),
			router.WithResolver(rr),
			router.WithRegistry(srv.Options().Registry),
		)
		ev := event.NewHandler(
			ahandler.WithNamespace(apiNamespace),
			ahandler.WithRouter(rt),
			ahandler.WithClient(srv.Client()),
		)
		r.PathPrefix(APIPath).Handler(ev)
	case "http", "proxy":
		log.Infof("Registering API HTTP Handler at %s", ProxyPath)
		rt := regRouter.NewRouter(
			router.WithHandler(ahttp.Handler),
			router.WithResolver(rr),
			router.WithRegistry(srv.Options().Registry),
		)
		ht := ahttp.NewHandler(
			ahandler.WithNamespace(apiNamespace),
			ahandler.WithRouter(rt),
			ahandler.WithClient(srv.Client()),
		)
		r.PathPrefix(ProxyPath).Handler(ht)
	case "web":
		log.Infof("Registering API Web Handler at %s", APIPath)
		rt := regRouter.NewRouter(
			router.WithHandler(aweb.Handler),
			router.WithResolver(rr),
			router.WithRegistry(srv.Options().Registry),
		)
		w := aweb.NewHandler(
			ahandler.WithNamespace(apiNamespace),
			ahandler.WithRouter(rt),
			ahandler.WithClient(srv.Client()),
		)
		r.PathPrefix(APIPath).Handler(w)
	default:
		log.Infof("Registering API Default Handler at %s", APIPath)
		rt := regRouter.NewRouter(
			router.WithResolver(rr),
			router.WithRegistry(srv.Options().Registry),
		)
		r.PathPrefix(APIPath).Handler(handler.Meta(srv, rt, nsResolver.ResolveWithType))
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
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}

	// Stop API
	if err := api.Stop(); err != nil {
		log.Fatal(err)
	}
}

func Commands(options ...service.Option) []*cli.Command {
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
