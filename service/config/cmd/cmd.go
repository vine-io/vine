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

package cmd

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/service/auth/provider"
	"github.com/lack-io/vine/service/broker"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/client/selector"
	"github.com/lack-io/vine/service/config"
	configSrc "github.com/lack-io/vine/service/config/source"
	"github.com/lack-io/vine/service/dao"
	"github.com/lack-io/vine/service/debug/profile"
	"github.com/lack-io/vine/service/debug/profile/http"
	"github.com/lack-io/vine/service/debug/profile/pprof"
	"github.com/lack-io/vine/service/debug/trace"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/registry"
	"github.com/lack-io/vine/service/runtime"
	"github.com/lack-io/vine/service/server"
	"github.com/lack-io/vine/service/store"
	"github.com/lack-io/vine/service/transport"
	authutil "github.com/lack-io/vine/util/auth"
	"github.com/lack-io/vine/util/wrapper"

	// registry
	registrySrv "github.com/lack-io/vine/service/registry/grpc"

	// clients
	cgrpc "github.com/lack-io/vine/service/client/grpc"
	cmucp "github.com/lack-io/vine/service/client/mucp"

	// servers
	sgrpc "github.com/lack-io/vine/service/server/grpc"
	smucp "github.com/lack-io/vine/service/server/mucp"

	// brokers
	brokerSrv "github.com/lack-io/vine/service/broker/grpc"
	brokerHttp "github.com/lack-io/vine/service/broker/http"
	"github.com/lack-io/vine/service/broker/memory"

	// registries
	"github.com/lack-io/vine/service/registry/etcd"
	regSrv "github.com/lack-io/vine/service/registry/grpc"
	"github.com/lack-io/vine/service/registry/mdns"
	rmem "github.com/lack-io/vine/service/registry/memory"

	// runtimes
	svcRuntime "github.com/lack-io/vine/service/runtime/grpc"
	lRuntime "github.com/lack-io/vine/service/runtime/local"

	// selectors
	"github.com/lack-io/vine/service/client/selector/dns"
	"github.com/lack-io/vine/service/client/selector/static"

	// transports
	thttp "github.com/lack-io/vine/service/transport/http"
	tmem "github.com/lack-io/vine/service/transport/memory"

	daoNop "github.com/lack-io/vine/service/dao/nop"

	// stores
	fileStore "github.com/lack-io/vine/service/store/bolt"
	svcStore "github.com/lack-io/vine/service/store/grpc"
	memStore "github.com/lack-io/vine/service/store/memory"

	// config
	configSrv "github.com/lack-io/vine/service/config/source/service"

	// tracers
	memTracer "github.com/lack-io/vine/service/debug/trace/memory"

	// auth
	svcAuth "github.com/lack-io/vine/service/auth/grpc"
	jwtAuth "github.com/lack-io/vine/service/auth/jwt"

	// auth providers
	"github.com/lack-io/vine/service/auth/provider/basic"
	"github.com/lack-io/vine/service/auth/provider/oauth"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Cmd interface {
	// App The cli app within this cmd
	App() *cli.App
	// Init Adds options, parses flags and initialise
	// exits on error
	Init(opts ...Option) error
	// Options set within this command
	Options() Options
}

type cmd struct {
	opts Options
	app  *cli.App
}

var (
	DefaultCmd = newCmd()

	DefaultFlags = []cli.Flag{
		&cli.StringFlag{
			Name:    "client",
			EnvVars: []string{"VINE_CLIENT"},
			Usage:   "Client for vine; rpc",
		},
		&cli.StringFlag{
			Name:    "client-request-timeout",
			EnvVars: []string{"VINE_CLIENT_REQUEST_TIMEOUT"},
			Usage:   "Sets the client request timeout. e.g 500ms, 5s, 1m. Default: 5s",
		},
		&cli.IntFlag{
			Name:    "client-retries",
			EnvVars: []string{"VINE_CLIENT_RETIES"},
			Value:   client.DefaultRetries,
			Usage:   "Sets the client retries. Default: 1",
		},
		&cli.IntFlag{
			Name:    "client_pool_size",
			EnvVars: []string{"VINE_CLIENT_POOL_SIZE"},
			Usage:   "Sets the client connection pool size. Default: 1",
		},
		&cli.StringFlag{
			Name:    "client-pool-ttl",
			EnvVars: []string{"VINE_CLIENT_POOL_TTL"},
			Usage:   "Sets the client connection pool ttl. e.g 500ms, 5s, 1m. Default: 1m",
		},
		&cli.IntFlag{
			Name:    "register-ttl",
			EnvVars: []string{"VINE_REGISTER_TTL"},
			Value:   60,
			Usage:   "Register TTL in seconds",
		},
		&cli.IntFlag{
			Name:    "register-interval",
			EnvVars: []string{"VINE_REGISTER_INTERVAL"},
			Value:   30,
			Usage:   "Register interval in seconds",
		},
		&cli.StringFlag{
			Name:    "server",
			EnvVars: []string{"VINE_SERVER"},
			Usage:   "Server for vine; rpc",
		},
		&cli.StringFlag{
			Name:    "server-name",
			EnvVars: []string{"VINE_SERVER_NAME"},
			Usage:   "Name of the server. go.vine.svc.example",
		},
		&cli.StringFlag{
			Name:    "server-version",
			EnvVars: []string{"VINE_SERVER_VERSION"},
			Usage:   "Version of the server. 1.1.0",
		},
		&cli.StringFlag{
			Name:    "server-id",
			EnvVars: []string{"VINE_SERVER_ID"},
			Usage:   "Id of the server. Auto-generated if not specified",
		},
		&cli.StringFlag{
			Name:    "server-address",
			EnvVars: []string{"VINE_SERVER_ADDRESS"},
			Usage:   "Bind address for the server. 127.0.0.1:8080",
		},
		&cli.StringFlag{
			Name:    "server-advertise",
			EnvVars: []string{"VINE_SERVER_ADVERTISE"},
			Usage:   "Use instead of the server-address when registering with discovery. 127.0.0.1:8080",
		},
		&cli.StringSliceFlag{
			Name:    "server-metadata",
			EnvVars: []string{"VINE_SERVER_METADATA"},
			Value:   &cli.StringSlice{},
			Usage:   "A list of key-value pairs defining metadata. version=1.0.0",
		},
		&cli.StringFlag{
			Name:    "broker",
			EnvVars: []string{"VINE_BROKER"},
			Usage:   "Broker for pub/sub. http, nats, rabbitmq",
		},
		&cli.StringFlag{
			Name:    "broker-address",
			EnvVars: []string{"VINE_BROKER_ADDRESS"},
			Usage:   "Comma-separated list of broker addresses",
		},
		&cli.StringFlag{
			Name:    "profile",
			Usage:   "Debug profiler for cpu and memory stats",
			EnvVars: []string{"VINE_DEBUG_PROFILE"},
		},
		&cli.StringFlag{
			Name:    "registry",
			EnvVars: []string{"VINE_REGISTRY"},
			Usage:   "Registry for discovery. etcd, mdns",
		},
		&cli.StringFlag{
			Name:    "registry-address",
			EnvVars: []string{"VINE_REGISTRY_ADDRESS"},
			Usage:   "Comma-separated list of registry addresses",
		},
		&cli.StringFlag{
			Name:    "runtime",
			Usage:   "Runtime for building and running services e.g local",
			EnvVars: []string{"VINE_RUNTIME"},
			Value:   "local",
		},
		&cli.StringFlag{
			Name:    "runtime-source",
			Usage:   "Runtime source for building and running services e.g github.com/lack-io/services",
			EnvVars: []string{"VINE_RUNTIME_SOURCE"},
			Value:   "github.com/lack-io/services",
		},
		&cli.StringFlag{
			Name:    "selector",
			EnvVars: []string{"VINE_SELECTOR"},
			Usage:   "Selector used to pick nodes for querying",
		},
		&cli.StringFlag{
			Name:    "dao",
			EnvVars: []string{"VINE_DAO"},
			Usage:   "Dao used for database service",
		},
		&cli.StringFlag{
			Name:    "dao-dialect",
			EnvVars: []string{"VINE_DAO_DIALECT"},
			Usage:   "Database option for the underlying dao",
		},
		&cli.StringFlag{
			Name:    "dao-dsn",
			EnvVars: []string{"VINE_DSN"},
			Usage:   "DSN database driver name for underlying dao",
		},
		&cli.StringFlag{
			Name:    "store",
			EnvVars: []string{"VINE_STORE"},
			Usage:   "Store used for key-value storage",
		},
		&cli.StringFlag{
			Name:    "store-address",
			EnvVars: []string{"VINE_STORE_ADDRESS"},
			Usage:   "Comma-separated list of store addresses",
		},
		&cli.StringFlag{
			Name:    "store-database",
			EnvVars: []string{"VINE_STORE_DATABASE"},
			Usage:   "Database option for the underlying store",
		},
		&cli.StringFlag{
			Name:    "store-table",
			EnvVars: []string{"VINE_STORE_TABLE"},
			Usage:   "Table option for the underlying store",
		},
		&cli.StringFlag{
			Name:    "transport",
			EnvVars: []string{"VINE_TRANSPORT"},
			Usage:   "Transport mechanism used; http",
		},
		&cli.StringFlag{
			Name:    "transport-address",
			EnvVars: []string{"VINE_TRANSPORT_ADDRESS"},
			Usage:   "Comma-separated list of transport addresses",
		},
		&cli.StringFlag{
			Name:    "tracer",
			EnvVars: []string{"VINE_TRACER"},
			Usage:   "Tracer for distributed tracing, e.g. memory, jaeger",
		},
		&cli.StringFlag{
			Name:    "tracer-address",
			EnvVars: []string{"VINE_TRACER_ADDRESS"},
			Usage:   "Comma-separated list of tracer addresses",
		},
		&cli.StringFlag{
			Name:    "auth",
			EnvVars: []string{"VINE_AUTH"},
			Usage:   "Auth for role based access control, e.g. service",
		},
		&cli.StringFlag{
			Name:    "auth-id",
			EnvVars: []string{"VINE_AUTH_ID"},
			Usage:   "Account ID used for client authentication",
		},
		&cli.StringFlag{
			Name:    "auth-secret",
			EnvVars: []string{"VINE_AUTH_SECRET"},
			Usage:   "Account secret used for client authentication",
		},
		&cli.StringFlag{
			Name:    "auth-namespace",
			EnvVars: []string{"VINE_AUTH_NAMESPACE"},
			Usage:   "Namespace for the services auth account",
			Value:   "go.vine",
		},
		&cli.StringFlag{
			Name:    "auth-public-key",
			EnvVars: []string{"VINE_AUTH_PUBLIC_KEY"},
			Usage:   "Public key for JWT auth (base64 encoded PEM)",
		},
		&cli.StringFlag{
			Name:    "auth-private-key",
			EnvVars: []string{"VINE_AUTH_PRIVATE_KEY"},
			Usage:   "Private key for JWT auth (base64 encoded PEM)",
		},
		&cli.StringFlag{
			Name:    "auth-provider",
			EnvVars: []string{"VINE_AUTH_PROVIDER"},
			Usage:   "Auth provider used to login user",
		},
		&cli.StringFlag{
			Name:    "auth-provide-client-id",
			EnvVars: []string{"VINE_AUTH_PROVIDER_CLIENT_ID"},
			Usage:   "The client id to be used for oauth",
		},
		&cli.StringFlag{
			Name:    "auth-provider-client-secret",
			EnvVars: []string{"VINE_AUTH_PROVIDER_CLIENT_SECRET"},
			Usage:   "The client secret to be used for oauth",
		},
		&cli.StringFlag{
			Name:    "auth-provider-endpoint",
			EnvVars: []string{"VINE_AUTH_PROVIDER_ENDPOINT"},
			Usage:   "The endpoint to be used for oauth",
		},
		&cli.StringFlag{
			Name:    "auth-provider-redirect",
			EnvVars: []string{"VINE_AUTH_PROVIDER_REDIRECT"},
			Usage:   "The redirect to be used for oauth",
		},
		&cli.StringFlag{
			Name:    "auth-provider-scope",
			EnvVars: []string{"VINE_AUTH_PROVIDER_SCOPE"},
			Usage:   "The scope to be used for oauth",
		},
		&cli.StringFlag{
			Name:    "config",
			EnvVars: []string{"VINE_CONFIG"},
			Usage:   "The source of the config to be used to get configuration",
		},
	}

	DefaultBrokers = map[string]func(...broker.Option) broker.Broker{
		"service": brokerSrv.NewBroker,
		"memory":  memory.NewBroker,
		"http":    brokerHttp.NewBroker,
	}

	DefaultClients = map[string]func(...client.Option) client.Client{
		"mucp": cmucp.NewClient,
		"grpc": cgrpc.NewClient,
	}

	DefaultRegistries = map[string]func(...registry.Option) registry.Registry{
		"etcd":    etcd.NewRegistry,
		"service": regSrv.NewRegistry,
		"mdns":    mdns.NewRegistry,
		"memory":  rmem.NewRegistry,
	}

	DefaultSelectors = map[string]func(...selector.Option) selector.Selector{
		"dns":    dns.NewSelector,
		"static": static.NewSelector,
	}

	DefaultServers = map[string]func(...server.Option) server.Server{
		"mucp": smucp.NewServer,
		"grpc": sgrpc.NewServer,
	}

	DefaultTransports = map[string]func(...transport.Option) transport.Transport{
		"memory": tmem.NewTransport,
		"http":   thttp.NewTransport,
	}

	DefaultRuntimes = map[string]func(...runtime.Option) runtime.Runtime{
		"local":   lRuntime.NewRuntime,
		"service": svcRuntime.NewRuntime,
	}

	DefaultDialects = map[string]func(...dao.Option) dao.Dialect{
		"nop": daoNop.NewDialect,
	}

	DefaultStores = map[string]func(...store.Option) store.Store{
		"file":    fileStore.NewStore,
		"memory":  memStore.NewStore,
		"service": svcStore.NewStore,
	}

	DefaultTracers = map[string]func(...trace.Option) trace.Tracer{
		"memory": memTracer.NewTracer,
		// "jaeger": jTracer.NewTracer,
	}

	DefaultAuths = map[string]func(...auth.Option) auth.Auth{
		"service": svcAuth.NewAuth,
		"jwt":     jwtAuth.NewAuth,
	}

	DefaultAuthProviders = map[string]func(...provider.Option) provider.Provider{
		"oauth": oauth.NewProvider,
		"basic": basic.NewProvider,
	}

	DefaultProfiles = map[string]func(...profile.Option) profile.Profile{
		"http":  http.NewProfile,
		"pprof": pprof.NewProfile,
	}

	DefaultConfigs = map[string]func(...config.Option) (config.Config, error){
		"service": config.NewConfig,
	}
)

func newCmd(opts ...Option) Cmd {
	options := Options{
		Auth:      &auth.DefaultAuth,
		Broker:    &broker.DefaultBroker,
		Client:    &client.DefaultClient,
		Registry:  &registry.DefaultRegistry,
		Server:    &server.DefaultServer,
		Selector:  &selector.DefaultSelector,
		Transport: &transport.DefaultTransport,
		Runtime:   &runtime.DefaultRuntime,
		Store:     &store.DefaultStore,
		Tracer:    &trace.DefaultTracer,
		Profile:   &profile.DefaultProfile,
		Config:    &config.DefaultConfig,

		Brokers:    DefaultBrokers,
		Clients:    DefaultClients,
		Registries: DefaultRegistries,
		Selectors:  DefaultSelectors,
		Servers:    DefaultServers,
		Transports: DefaultTransports,
		Runtimes:   DefaultRuntimes,
		Dialects:   DefaultDialects,
		Stores:     DefaultStores,
		Tracers:    DefaultTracers,
		Auths:      DefaultAuths,
		Profiles:   DefaultProfiles,
		Configs:    DefaultConfigs,
	}

	for _, o := range opts {
		o(&options)
	}

	if len(options.Description) == 0 {
		options.Description = "a vine service"
	}

	cmd := new(cmd)
	cmd.opts = options
	cmd.app = cli.NewApp()
	cmd.app.Name = cmd.opts.Name
	cmd.app.Version = cmd.opts.Version
	cmd.app.Usage = cmd.opts.Description
	cmd.app.Before = cmd.Before
	cmd.app.Flags = DefaultFlags
	cmd.app.Action = func(c *cli.Context) error {
		return nil
	}

	if len(options.Version) == 0 {
		cmd.app.HideVersion = true
	}

	return cmd
}

func (c *cmd) App() *cli.App {
	return c.app
}

func (c *cmd) Init(opts ...Option) error {
	for _, o := range opts {
		o(&c.opts)
	}
	if len(c.opts.Name) > 0 {
		c.app.Name = c.opts.Name
	}
	if len(c.opts.Version) > 0 {
		c.app.Version = c.opts.Version
	}
	c.app.HideVersion = len(c.opts.Version) == 0
	c.app.Usage = c.opts.Description
	c.app.RunAndExitOnError()
	return nil
}

func (c *cmd) Options() Options {
	return c.opts
}

func (c *cmd) Before(ctx *cli.Context) error {
	// If flags are set then use them otherwise do nothing
	var serverOpts []server.Option
	var clientOpts []client.Option

	// setup a client to use when calling the runtime. It is important the auth client is wrapped
	// after the cache client since the wrappers are applied in reverse order and the cache will use
	// some of the headers set by the auth client.
	authFn := func() auth.Auth { return *c.opts.Auth }
	cacheFn := func() *client.Cache { return (*c.opts.Client).Options().Cache }
	vineClient := wrapper.CacheClient(cacheFn, cgrpc.NewClient())
	vineClient = wrapper.AuthClient(authFn, vineClient)

	// Set the store
	if name := ctx.String("store"); len(name) > 0 {
		s, ok := c.opts.Stores[name]
		if !ok {
			return fmt.Errorf("unsuported store: %s", name)
		}

		*c.opts.Store = s(store.WithClient(vineClient))
	}

	// Set the runtime
	if name := ctx.String("runtime"); len(name) > 0 {
		r, ok := c.opts.Runtimes[name]
		if !ok {
			return fmt.Errorf("unsupport runtime: %s", name)
		}

		*c.opts.Runtime = r(runtime.WithClient(vineClient))
	}

	// Set the tracer
	if name := ctx.String("tracer"); len(name) > 0 {
		r, ok := c.opts.Tracers[name]
		if !ok {
			return fmt.Errorf("unsupported tracer: %s", name)
		}

		*c.opts.Tracer = r()
	}

	// Set the client
	if name := ctx.String("client"); len(name) > 0 {
		// only change if we have the client and type differs
		if cl, ok := c.opts.Clients[name]; ok && (*c.opts.Client).String() != name {
			*c.opts.Client = cl()
		}
	}

	// Set the server
	if name := ctx.String("server"); len(name) > 0 {
		// only change if we have the server and type differs
		if s, ok := c.opts.Servers[name]; ok && (*c.opts.Server).String() != name {
			*c.opts.Server = s()
		}
	}

	// Set auth
	authOpts := []auth.Option{auth.WithClient(vineClient)}

	if authId, authSecret := ctx.String("auth-id"), ctx.String("auth-secret"); len(authId) > 0 || len(authSecret) > 0 {
		authOpts = append(authOpts, auth.Credentials(authId, authSecret))
	}
	if key := ctx.String("auth-public-key"); len(key) > 0 {
		authOpts = append(authOpts, auth.PublicKey(key))
	}
	if key := ctx.String("auth-private-key"); len(key) > 0 {
		authOpts = append(authOpts, auth.PrivateKey(key))
	}
	if ns := ctx.String("auth-namespace"); len(ns) > 0 {
		authOpts = append(authOpts, auth.Namespace(ns))
	}
	if name := ctx.String("auth-provider"); len(name) > 0 {
		p, ok := DefaultAuthProviders[name]
		if !ok {
			return fmt.Errorf("AuthProvider %s not found", name)
		}

		var provOpts []provider.Option
		clientID := ctx.String("auth-provider-client-id")
		clientSecret := ctx.String("auth-provider-client-secret")
		if len(clientID) > 0 || len(clientSecret) > 0 {
			provOpts = append(provOpts, provider.Credentials(clientID, clientSecret))
		}
		if e := ctx.String("auth-provider-endpoint"); len(e) > 0 {
			provOpts = append(provOpts, provider.Endpoint(e))
		}
		if r := ctx.String("auth-provider-redirect"); len(r) > 0 {
			provOpts = append(provOpts, provider.Redirect(r))
		}
		if s := ctx.String("auth-provider-scope"); len(s) > 0 {
			provOpts = append(provOpts, provider.Scope(s))
		}

		authOpts = append(authOpts, auth.Provider(p(provOpts...)))
	}

	// Set the auth
	if name := ctx.String("auth"); len(name) > 0 {
		a, ok := c.opts.Auths[name]
		if !ok {
			return fmt.Errorf("unsupported auth: %s", name)
		}
		*c.opts.Auth = a(authOpts...)
		serverOpts = append(serverOpts, server.Auth(*c.opts.Auth))
	} else {
		(*c.opts.Auth).Init(authOpts...)
	}

	// Set the registry
	if name := ctx.String("registry"); len(name) > 0 && (*c.opts.Registry).String() != name {
		r, ok := c.opts.Registries[name]
		if !ok {
			return fmt.Errorf("registry %s not found", name)
		}

		*c.opts.Registry = r(registrySrv.WithClient(vineClient))
		serverOpts = append(serverOpts, server.Registry(*c.opts.Registry))
		clientOpts = append(clientOpts, client.Registry(*c.opts.Registry))

		if err := (*c.opts.Selector).Init(selector.Registry(*c.opts.Registry)); err != nil {
			log.Fatalf("Error configuring registry: %v", err)
		}

		clientOpts = append(clientOpts, client.Selector(*c.opts.Selector))

		if err := (*c.opts.Broker).Init(broker.Registry(*c.opts.Registry)); err != nil {
			log.Errorf("Error configuring broker: %v", err)
		}
	}

	// generate the services auth account
	serverID := (*c.opts.Server).Options().Id
	if err := authutil.Generate(serverID, c.App().Name, *c.opts.Auth); err != nil {
		return err
	}

	// Set the profile
	if name := ctx.String("profile"); len(name) > 0 {
		p, ok := c.opts.Profiles[name]
		if !ok {
			return fmt.Errorf("unsupport profile: %s", name)
		}

		*c.opts.Profile = p()
	}

	// Set the broker
	if name := ctx.String("broker"); len(name) > 0 && (*c.opts.Broker).String() != name {
		b, ok := c.opts.Brokers[name]
		if !ok {
			return fmt.Errorf("broker %s not found", name)
		}

		*c.opts.Broker = b()
		serverOpts = append(serverOpts, server.Broker(*c.opts.Broker))
		clientOpts = append(clientOpts, client.Broker(*c.opts.Broker))
	}

	// Set the selector
	if name := ctx.String("selector"); len(name) > 0 && (*c.opts.Selector).String() != name {
		s, ok := c.opts.Selectors[name]
		if !ok {
			return fmt.Errorf("selector %s not found", name)
		}

		*c.opts.Selector = s(selector.Registry(*c.opts.Registry))

		// No server option here. Should there be?
		clientOpts = append(clientOpts, client.Selector(*c.opts.Selector))
	}

	// Set the transport
	if name := ctx.String("transport"); len(name) > 0 && (*c.opts.Transport).String() != name {
		t, ok := c.opts.Transports[name]
		if !ok {
			return fmt.Errorf("transport %s not found", name)
		}

		*c.opts.Transport = t()
		serverOpts = append(serverOpts, server.Transport(*c.opts.Transport))
		clientOpts = append(clientOpts, client.Transport(*c.opts.Transport))
	}

	// Parse the server options
	metadata := make(map[string]string)
	for _, d := range ctx.StringSlice("server-metadata") {
		var key, val string
		parts := strings.Split(d, "=")
		key = parts[0]
		if len(parts) > 1 {
			val = strings.Join(parts[1:], "=")
		}
		metadata[key] = val
	}

	if len(metadata) > 0 {
		serverOpts = append(serverOpts, server.Metadata(metadata))
	}

	if addrs := ctx.String("broker-address"); len(addrs) > 0 {
		if err := (*c.opts.Broker).Init(broker.Addrs(strings.Split(addrs, ",")...)); err != nil {
			log.Fatalf("Error configuring broker: %v", err)
		}
	}

	if addrs := ctx.String("registry-address"); len(addrs) > 0 {
		if err := (*c.opts.Registry).Init(registry.Addrs(strings.Split(addrs, ",")...)); err != nil {
			log.Fatalf("Error configuring registry: %v", err)
		}
	}

	if addrs := ctx.String("transport-address"); len(addrs) > 0 {
		if err := (*c.opts.Transport).Init(transport.Addrs(strings.Split(addrs, ",")...)); err != nil {
			log.Fatalf("Error configuring store: %v", err)
		}
	}

	if addrs := ctx.String("store-address"); len(addrs) > 0 {
		if err := (*c.opts.Store).Init(store.Nodes(strings.Split(addrs, ",")...)); err != nil {
			log.Fatalf("Error configuring store: %v", err)
		}
	}

	if db := ctx.String("store-database"); len(db) > 0 {
		if err := (*c.opts.Store).Init(store.Database(db)); err != nil {
			log.Fatalf("Error configuring store database option: %v", err)
		}
	}

	if table := ctx.String("store-table"); len(table) > 0 {
		if err := (*c.opts.Store).Init(store.Table(table)); err != nil {
			log.Fatalf("Error configuring store table option: %v", err)
		}
	}

	if name := ctx.String("server-name"); len(name) > 0 {
		serverOpts = append(serverOpts, server.Name(name))
	}

	if version := ctx.String("server-version"); len(version) > 0 {
		serverOpts = append(serverOpts, server.Version(version))
	}

	if id := ctx.String("server-id"); len(id) > 0 {
		serverOpts = append(serverOpts, server.Id(id))
	}

	if addr := ctx.String("server-address"); len(addr) > 0 {
		serverOpts = append(serverOpts, server.Address(addr))
	}

	if advertise := ctx.String("server-advertise"); len(advertise) > 0 {
		serverOpts = append(serverOpts, server.Advertise(advertise))
	}

	if ttl := time.Duration(ctx.Int("register-ttl")); ttl >= 0 {
		serverOpts = append(serverOpts, server.RegisterTTL(ttl*time.Second))
	}

	if val := time.Duration(ctx.Int("register-interval")); val >= 0 {
		serverOpts = append(serverOpts, server.RegisterInterval(val*time.Second))
	}

	if source := ctx.String("runtime-source"); len(source) > 0 {
		if err := (*c.opts.Runtime).Init(runtime.WithSource(source)); err != nil {
			log.Fatalf("Error configuring runtime: %v", err)
		}
	}

	if ctx.String("config") == "service" {
		opt := config.WithSource(configSrv.NewSource(configSrc.WithClient(vineClient)))
		if err := (*c.opts.Config).Init(opt); err != nil {
			log.Fatalf("Error configuring config: %v", err)
		}
	}

	// client opts
	if r := ctx.Int("client-retries"); r >= 0 {
		clientOpts = append(clientOpts, client.Retries(r))
	}

	if t := ctx.String("client-request-timeout"); len(t) > 0 {
		d, err := time.ParseDuration(t)
		if err != nil {
			return fmt.Errorf("failed to parse client-request-timeout: %v", t)
		}
		clientOpts = append(clientOpts, client.RequestTimeout(d))
	}

	if r := ctx.Int("client-pool-size"); r > 0 {
		clientOpts = append(clientOpts, client.PoolSize(r))
	}

	if t := ctx.String("client-pool-ttl"); len(t) > 0 {
		d, err := time.ParseDuration(t)
		if err != nil {
			return fmt.Errorf("failed to parse client-pool-ttl: %v", t)
		}
		clientOpts = append(clientOpts, client.PoolTTL(d))
	}

	// We have some command line opts for the server.
	// Lets set it up
	if len(serverOpts) > 0 {
		if err := (*c.opts.Server).Init(serverOpts...); err != nil {
			log.Fatalf("Error configuring server: %v", err)
		}
	}

	// Use an init option?
	if len(clientOpts) > 0 {
		if err := (*c.opts.Client).Init(clientOpts...); err != nil {
			log.Fatalf("Error configuring client: %v", err)
		}
	}

	return nil
}

func DefaultOptions() Options {
	return DefaultCmd.Options()
}

func App() *cli.App {
	return DefaultCmd.App()
}

func Init(opts ...Option) error {
	return DefaultCmd.Init(opts...)
}

func NewCmd(opts ...Option) Cmd {
	return newCmd(opts...)
}
