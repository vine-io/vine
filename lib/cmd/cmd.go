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
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/vine-io/cli"
	"github.com/vine-io/vine/core/broker"
	brokerHttp "github.com/vine-io/vine/core/broker/http"
	"github.com/vine-io/vine/core/broker/memory"
	"github.com/vine-io/vine/core/client"
	cGrpc "github.com/vine-io/vine/core/client/grpc"
	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/core/client/selector/dns"
	"github.com/vine-io/vine/core/client/selector/static"
	"github.com/vine-io/vine/core/registry"
	"github.com/vine-io/vine/core/registry/mdns"
	regMemory "github.com/vine-io/vine/core/registry/memory"
	"github.com/vine-io/vine/core/server"
	"github.com/vine-io/vine/lib/cache"
	"github.com/vine-io/vine/lib/config"
	configMemory "github.com/vine-io/vine/lib/config/memory"
	"github.com/vine-io/vine/lib/dao"
	log "github.com/vine-io/vine/lib/logger"
	"github.com/vine-io/vine/lib/trace"
	memTracer "github.com/vine-io/vine/lib/trace/memory"

	// servers
	sgrpc "github.com/vine-io/vine/core/server/grpc"

	daoNop "github.com/vine-io/vine/lib/dao/nop"

	memCache "github.com/vine-io/vine/lib/cache/memory"
	nopCache "github.com/vine-io/vine/lib/cache/noop"

	// config
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
			Name:    "client-pool-size",
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
			Usage:   "Broker for pub/sub. http, nats",
		},
		&cli.StringFlag{
			Name:    "broker-address",
			EnvVars: []string{"VINE_BROKER_ADDRESS"},
			Usage:   "Comma-separated list of broker addresses",
		},
		&cli.StringFlag{
			Name:    "registry",
			EnvVars: []string{"VINE_REGISTRY"},
			Usage:   "Registry for discovery. memory, mdns",
		},
		&cli.StringFlag{
			Name:    "registry-address",
			EnvVars: []string{"VINE_REGISTRY_ADDRESS"},
			Usage:   "Comma-separated list of registry addresses",
		},
		&cli.StringFlag{
			Name:    "selector",
			EnvVars: []string{"VINE_SELECTOR"},
			Usage:   "Selector used to pick nodes for querying",
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
			Name:    "config",
			EnvVars: []string{"VINE_CONFIG"},
			Usage:   "The source of the config to be used to get configuration",
		},
		&cli.StringFlag{
			Name:    "cache",
			EnvVars: []string{"VINE_CACHE"},
			Usage:   "Cache used for key-value storage",
		},
		&cli.StringFlag{
			Name:    "cache-address",
			EnvVars: []string{"VINE_CACHE_ADDRESS"},
			Usage:   "Comma-separated list of cache addresses",
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
	}

	DefaultBrokers = map[string]func(...broker.Option) broker.Broker{
		"memory":  memory.NewBroker,
		"http":    brokerHttp.NewBroker,
	}

	DefaultClients = map[string]func(...client.Option) client.Client{
		"grpc": cGrpc.NewClient,
	}

	DefaultRegistries = map[string]func(...registry.Option) registry.Registry{
		"mdns":    mdns.NewRegistry,
		"memory":  regMemory.NewRegistry,
	}

	DefaultSelectors = map[string]func(...selector.Option) selector.Selector{
		"dns":    dns.NewSelector,
		"static": static.NewSelector,
	}

	DefaultServers = map[string]func(...server.Option) server.Server{
		"grpc": sgrpc.NewServer,
	}

	DefaultDialects = map[string]func(...dao.Option) dao.Dialect{
		"nop": daoNop.NewDialect,
	}

	DefaultCaches = map[string]func(...cache.Option) cache.Cache{
		"memory": memCache.NewCache,
		"noop":   nopCache.NewCache,
	}

	DefaultTracers = map[string]func(...trace.Option) trace.Tracer{
		"memory": memTracer.NewTracer,
		// "jaeger": jTracer.NewTracer,
	}

	DefaultConfigs = map[string]func(...config.Option) config.Config{
		"memory": configMemory.NewConfig,
	}
)

func newCmd(opts ...Option) Cmd {
	options := Options{
		Broker:   &broker.DefaultBroker,
		Client:   &client.DefaultClient,
		Registry: &registry.DefaultRegistry,
		Server:   &server.DefaultServer,
		Selector: &selector.DefaultSelector,
		Dialect:  &dao.DefaultDialect,
		Cache:    &cache.DefaultCache,
		Tracer:   &trace.DefaultTracer,
		Config:   &config.DefaultConfig,

		Brokers:    DefaultBrokers,
		Clients:    DefaultClients,
		Registries: DefaultRegistries,
		Selectors:  DefaultSelectors,
		Servers:    DefaultServers,
		Dialects:   DefaultDialects,
		Caches:     DefaultCaches,
		Tracers:    DefaultTracers,
		Configs:    DefaultConfigs,
	}

	for _, o := range opts {
		o(&options)
	}

	if len(options.Description) == 0 {
		options.Description = "a vine service"
	}

	if options.Context == nil {
		options.Context = context.Background()
	}

	cmd := new(cmd)
	cmd.opts = options
	cmd.opts.app = cli.NewApp()
	cmd.opts.app.Name = cmd.opts.Name
	cmd.opts.app.Version = cmd.opts.Version
	cmd.opts.app.Usage = cmd.opts.Description
	cmd.opts.app.Before = cmd.Before
	cmd.opts.app.Flags = DefaultFlags
	cmd.opts.app.Action = func(c *cli.Context) error {
		return nil
	}

	if cmd.opts.app.Before == nil {
		cmd.opts.app.Before = func(c *cli.Context) error {
			return nil
		}
	}

	if len(options.Version) == 0 {
		cmd.opts.app.HideVersion = true
	}

	return cmd
}

func (c *cmd) App() *cli.App {
	return c.opts.app
}

func (c *cmd) Init(opts ...Option) error {
	for _, o := range opts {
		o(&c.opts)
	}
	if len(c.opts.Name) > 0 {
		c.opts.app.Name = c.opts.Name
	}
	if len(c.opts.Version) > 0 {
		c.opts.app.Version = c.opts.Version
	}
	c.opts.app.HideVersion = len(c.opts.Version) == 0
	c.opts.app.Usage = c.opts.Description
	c.opts.app.RunAndExitOnError()
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
	vineClient := client.DefaultClient

	// Set the cache
	if name := ctx.String("cache"); len(name) > 0 {
		s, ok := c.opts.Caches[name]
		if !ok {
			return fmt.Errorf("unsuported cache: %s", name)
		}

		*c.opts.Cache = s(cache.WithClient(vineClient))
		cache.DefaultCache = *c.opts.Cache
	}

	// Set the dialect
	if name := ctx.String("dao-dialect"); len(name) > 0 {
		d, ok := c.opts.Dialects[name]
		if !ok {
			return fmt.Errorf("unsuported dialect: %s", name)
		}

		*c.opts.Dialect = d()
		dao.DefaultDialect = *c.opts.Dialect
	}

	// Set the tracer
	if name := ctx.String("tracer"); len(name) > 0 {
		r, ok := c.opts.Tracers[name]
		if !ok {
			return fmt.Errorf("unsupported tracer: %s", name)
		}

		*c.opts.Tracer = r()
		trace.DefaultTracer = *c.opts.Tracer
	}

	// Set the client
	if name := ctx.String("client"); len(name) > 0 {
		// only change if we have the client and type differs
		if cl, ok := c.opts.Clients[name]; ok && (*c.opts.Client).String() != name {
			*c.opts.Client = cl()
			client.DefaultClient = *c.opts.Client
		}
	}

	// Set the server
	if name := ctx.String("server"); len(name) > 0 {
		// only change if we have the server and type differs
		if s, ok := c.opts.Servers[name]; ok && (*c.opts.Server).String() != name {
			*c.opts.Server = s()
			server.DefaultServer = *c.opts.Server
		}
	}

	// Set the registry
	if name := ctx.String("registry"); len(name) > 0 && (*c.opts.Registry).String() != name {
		r, ok := c.opts.Registries[name]
		if !ok {
			return fmt.Errorf("registry %s not found", name)
		}

		*c.opts.Registry = r()
		registry.DefaultRegistry = *c.opts.Registry

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

	// Set the broker
	if name := ctx.String("broker"); len(name) > 0 && (*c.opts.Broker).String() != name {
		b, ok := c.opts.Brokers[name]
		if !ok {
			return fmt.Errorf("broker %s not found", name)
		}

		*c.opts.Broker = b()
		broker.DefaultBroker = *c.opts.Broker

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
		selector.DefaultSelector = *c.opts.Selector

		// No server option here. Should there be?
		clientOpts = append(clientOpts, client.Selector(*c.opts.Selector))
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

	if dsn := ctx.String("dao-dsn"); len(dsn) > 0 {
		if strings.HasPrefix(dsn, "base64:") {
			b, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(dsn, "base64:"))
			if err != nil {
				log.Fatalf("Error configuring dialect dsn: decode base64 string: %v", err)
			}
			dsn = string(b)
		}
		if err := (*c.opts.Dialect).Init(dao.DSN(dsn)); err != nil {
			log.Fatalf("Error configuring dialect dsn: %v", err)
		}
	}

	if addrs := ctx.String("cache-address"); len(addrs) > 0 {
		if err := (*c.opts.Cache).Init(cache.Nodes(strings.Split(addrs, ",")...)); err != nil {
			log.Fatalf("Error configuring cache: %v", err)
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
	// Let's set it up
	if len(serverOpts) > 0 && *c.opts.Server != nil {
		if err := (*c.opts.Server).Init(serverOpts...); err != nil {
			log.Fatalf("Error configuring server: %v", err)
		}
	}

	// Use an init option?
	if len(clientOpts) > 0 && *c.opts.Client != nil {
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
