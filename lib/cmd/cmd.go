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
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
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
	"gopkg.in/yaml.v3"

	// servers
	sgrpc "github.com/vine-io/vine/core/server/grpc"

	daoNop "github.com/vine-io/vine/lib/dao/nop"

	memCache "github.com/vine-io/vine/lib/cache/memory"
	nopCache "github.com/vine-io/vine/lib/cache/noop"
	// config
	uc "github.com/vine-io/vine/util/config"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Cmd interface {
	// App The cli app within this cmd
	App() *cobra.Command
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

	DefaultBrokers = map[string]func(...broker.Option) broker.Broker{
		"memory": memory.NewBroker,
		"http":   brokerHttp.NewBroker,
	}

	DefaultClients = map[string]func(...client.Option) client.Client{
		"grpc": cGrpc.NewClient,
	}

	DefaultRegistries = map[string]func(...registry.Option) registry.Registry{
		"mdns":   mdns.NewRegistry,
		"memory": regMemory.NewRegistry,
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

	c := new(cmd)
	rootCmd := &cobra.Command{
		Use:   options.Name,
		Short: options.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.before(cmd, args)
		},
	}
	rootCmd.SetHelpFunc(help)
	rootCmd.Version = ""
	rootCmd.AddCommand(c.Commands()...)
	rootCmd.InitDefaultCompletionCmd()

	rootCmd.ResetFlags()
	rootCmd.PersistentFlags().StringP("input", "i", "", "Reads the configuration from given filename")
	rootCmd.PersistentFlags().AddFlagSet(registry.Flag)
	rootCmd.PersistentFlags().AddFlagSet(broker.Flag)
	rootCmd.PersistentFlags().AddFlagSet(client.Flag)
	rootCmd.PersistentFlags().AddFlagSet(selector.Flag)
	rootCmd.PersistentFlags().AddFlagSet(server.Flag)
	rootCmd.PersistentFlags().AddFlagSet(dao.Flag)
	rootCmd.PersistentFlags().AddFlagSet(cache.Flag)
	rootCmd.PersistentFlags().AddFlagSet(trace.Flag)

	options.app = rootCmd
	c.opts = options
	return c
}

func (c *cmd) App() *cobra.Command {
	return c.opts.app
}

func (c *cmd) Init(opts ...Option) error {
	for _, o := range opts {
		o(&c.opts)
	}
	if len(c.opts.Name) > 0 {
		c.opts.app.Use = c.opts.Name
	}

	c.opts.app.Short = c.opts.Description
	if c.opts.app.RunE == nil {
		c.opts.app.RunE = func(cmd *cobra.Command, args []string) error {
			return c.before(cmd, args)
		}
	}

	err := uc.BindPFlags(c.opts.app.PersistentFlags())
	if err != nil {
		return fmt.Errorf("binding flags: %v", err)
	}

	return c.opts.app.Execute()
}

func (c *cmd) Options() Options {
	return c.opts
}

func (c *cmd) Commands() []*cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Prints the version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(c.opts.Version)
			os.Exit(0)
			return nil
		},
	}

	defaultCmd := &cobra.Command{
		Use:   "default",
		Short: "Prints configuration data",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := map[string]interface{}{}
			err := uc.Unmarshal(&out)
			if err != nil {
				return err
			}
			data, _ := yaml.Marshal(out)
			cmd.Print(string(data))
			os.Exit(0)
			return nil
		},
	}

	return []*cobra.Command{versionCmd, defaultCmd}
}

func (c *cmd) before(cmd *cobra.Command, args []string) error {
	// If flags are set then use them otherwise do nothing
	var serverOpts []server.Option
	var clientOpts []client.Option
	options := c.opts

	// setup a client to use when calling the runtime. It is important the auth client is wrapped
	// after the cache client since the wrappers are applied in reverse order and the cache will use
	vineClient := client.DefaultClient

	if cfg, _ := cmd.PersistentFlags().GetString("input"); len(cfg) != 0 {
		data, err := os.ReadFile(cfg)
		if err != nil {
			return fmt.Errorf("open configuration file: %v", err)
		}
		err = uc.ReadConfig(bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("parse configration data: %v", err)
		}
	}

	// Set the cache
	if name := uc.GetString("cache.default"); len(name) > 0 {
		s, ok := options.Caches[name]
		if !ok {
			return fmt.Errorf("unsuported cache: %s", name)
		}

		*options.Cache = s(cache.WithClient(vineClient))
		cache.DefaultCache = *options.Cache
	}

	// Set the dialect
	if name := uc.GetString("dao.dialect"); len(name) > 0 {
		d, ok := options.Dialects[name]
		if !ok {
			return fmt.Errorf("unsuported dialect: %s", name)
		}

		*options.Dialect = d()
		dao.DefaultDialect = *options.Dialect
	}

	// Set the tracer
	if name := uc.GetString("tracer.default"); len(name) > 0 {
		r, ok := options.Tracers[name]
		if !ok {
			return fmt.Errorf("unsupported tracer: %s", name)
		}

		*options.Tracer = r()
		trace.DefaultTracer = *options.Tracer
	}

	// Set the client
	if name := uc.GetString("client.default"); len(name) > 0 {
		// only change if we have the client and type differs
		if cl, ok := options.Clients[name]; ok && (*options.Client).String() != name {
			*options.Client = cl()
			client.DefaultClient = *options.Client
		}
	}

	// Set the server
	if name := uc.GetString("server.default"); len(name) > 0 {
		// only change if we have the server and type differs
		if s, ok := options.Servers[name]; ok && (*options.Server).String() != name {
			*options.Server = s()
			server.DefaultServer = *options.Server
		}
	}

	// Set the registry
	if name := uc.GetString("registry.default"); len(name) > 0 && (*options.Registry).String() != name {
		r, ok := options.Registries[name]
		if !ok {
			return fmt.Errorf("registry %s not found", name)
		}

		*options.Registry = r()
		registry.DefaultRegistry = *options.Registry

		serverOpts = append(serverOpts, server.Registry(*options.Registry))
		clientOpts = append(clientOpts, client.Registry(*options.Registry))

		if err := (*options.Selector).Init(selector.Registry(*options.Registry)); err != nil {
			log.Fatalf("Error configuring registry: %v", err)
		}

		clientOpts = append(clientOpts, client.Selector(*options.Selector))

		if err := (*options.Broker).Init(broker.Registry(*options.Registry)); err != nil {
			log.Errorf("Error configuring broker: %v", err)
		}
	}

	// Set the broker
	if name := uc.GetString("broker.default"); len(name) > 0 && (*options.Broker).String() != name {
		b, ok := options.Brokers[name]
		if !ok {
			return fmt.Errorf("broker %s not found", name)
		}

		*options.Broker = b()
		broker.DefaultBroker = *options.Broker

		serverOpts = append(serverOpts, server.Broker(*options.Broker))
		clientOpts = append(clientOpts, client.Broker(*options.Broker))
	}

	// Set the selector
	if name := uc.GetString("selector.default"); len(name) > 0 && (*options.Selector).String() != name {
		s, ok := options.Selectors[name]
		if !ok {
			return fmt.Errorf("selector %s not found", name)
		}

		*options.Selector = s(selector.Registry(*options.Registry))
		selector.DefaultSelector = *options.Selector

		// No server option here. Should there be?
		clientOpts = append(clientOpts, client.Selector(*options.Selector))
	}

	// Parse the server options
	metadata := make(map[string]string)
	for _, d := range uc.GetStringSlice("server.metadata") {
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

	if addrs := uc.GetString("broker.address"); len(addrs) > 0 {
		if err := (*options.Broker).Init(broker.Addrs(strings.Split(addrs, ",")...)); err != nil {
			log.Fatalf("Error configuring broker: %v", err)
		}
	}

	if addrs := uc.GetString("registry.address"); len(addrs) > 0 {
		if err := (*options.Registry).Init(registry.Addrs(strings.Split(addrs, ",")...)); err != nil {
			log.Fatalf("Error configuring registry: %v", err)
		}
	}

	if dsn := uc.GetString("dao.dsn"); len(dsn) > 0 {
		if strings.HasPrefix(dsn, "base64:") {
			b, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(dsn, "base64:"))
			if err != nil {
				log.Fatalf("Error configuring dialect dsn: decode base64 string: %v", err)
			}
			dsn = string(b)
		}
		if err := (*options.Dialect).Init(dao.DSN(dsn)); err != nil {
			log.Fatalf("Error configuring dialect dsn: %v", err)
		}
	}

	if addrs := uc.GetString("cache.address"); len(addrs) > 0 {
		if err := (*options.Cache).Init(cache.Nodes(strings.Split(addrs, ",")...)); err != nil {
			log.Fatalf("Error configuring cache: %v", err)
		}
	}

	if name := uc.GetString("server.name"); len(name) > 0 {
		serverOpts = append(serverOpts, server.Name(name))
	}

	if version := uc.GetString("server.version"); len(version) > 0 {
		serverOpts = append(serverOpts, server.Version(version))
	}

	if id := uc.GetString("server.id"); len(id) > 0 {
		serverOpts = append(serverOpts, server.Id(id))
	}

	if addr := uc.GetString("server.address"); len(addr) > 0 {
		serverOpts = append(serverOpts, server.Address(addr))
	}

	if advertise := uc.GetString("server.advertise"); len(advertise) > 0 {
		serverOpts = append(serverOpts, server.Advertise(advertise))
	}

	if ttl := uc.GetDuration("registry.ttl"); ttl >= 0 {
		serverOpts = append(serverOpts, server.RegisterTTL(ttl*time.Second))
	}

	if val := uc.GetDuration("registry.interval"); val >= 0 {
		serverOpts = append(serverOpts, server.RegisterInterval(val*time.Second))
	}

	// client opts
	if r := uc.GetInt("client.retries"); r >= 0 {
		clientOpts = append(clientOpts, client.Retries(r))
	}

	if t := uc.GetString("client.dial-timeout"); len(t) > 0 {
		d, err := time.ParseDuration(t)
		if err != nil {
			return fmt.Errorf("failed to parse client.dialTimeout: %v", t)
		}
		clientOpts = append(clientOpts, client.DialTimeout(d))
	}

	if t := uc.GetString("client.request-timeout"); len(t) > 0 {
		d, err := time.ParseDuration(t)
		if err != nil {
			return fmt.Errorf("failed to parse client.requestTimeout: %v", t)
		}
		clientOpts = append(clientOpts, client.RequestTimeout(d))
	}

	if r := uc.GetInt("client.pool-size"); r > 0 {
		clientOpts = append(clientOpts, client.PoolSize(r))
	}

	if t := uc.GetString("client.pool-ttl"); len(t) > 0 {
		d, err := time.ParseDuration(t)
		if err != nil {
			return fmt.Errorf("failed to parse client.pool.ttl: %v", t)
		}
		clientOpts = append(clientOpts, client.PoolTTL(d))
	}

	// We have some command line opts for the server.
	// Let's set it up
	if len(serverOpts) > 0 && *options.Server != nil {
		if err := (*options.Server).Init(serverOpts...); err != nil {
			log.Fatalf("Error configuring server: %v", err)
		}
	}

	// Use an init option?
	if len(clientOpts) > 0 && *options.Client != nil {
		if err := (*options.Client).Init(clientOpts...); err != nil {
			log.Fatalf("Error configuring client: %v", err)
		}
	}

	return nil
}

func help(cmd *cobra.Command, _ []string) {
	cmd.Print(cmd.UsageString())
	os.Exit(0)
}

func DefaultOptions() Options {
	return DefaultCmd.Options()
}

func App() *cobra.Command {
	return DefaultCmd.App()
}

func Init(opts ...Option) error {
	return DefaultCmd.Init(opts...)
}

func NewCmd(opts ...Option) Cmd {
	return newCmd(opts...)
}
