// MIT License
//
// Copyright (c) 2020 The vine Authors
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

package app

import (
	"math"
	"sort"

	"github.com/spf13/cobra"
	"github.com/vine-io/vine"
	"github.com/vine-io/vine/cmd/vine/app/api"
	cliBuild "github.com/vine-io/vine/cmd/vine/app/cli/build"
	cliMg "github.com/vine-io/vine/cmd/vine/app/cli/mg"
	cliRun "github.com/vine-io/vine/cmd/vine/app/cli/run"
	"github.com/vine-io/vine/cmd/vine/version"
	"github.com/vine-io/vine/lib/cmd"
)

var (
	name        = "vine"
	description = `A vine service runtime
        _
 _   __(_)___  ___
| | / / / __ \/ _ \
| |/ / / / / /  __/
|___/_/_/ /_/\___/`
)

func init() {
}

func setup(cmd *cobra.Command) {
	//app.Flags = append(
	//	app.Flags,
	//	&ccli.BoolFlag{
	//		Name:  "local",
	//		Usage: "Enable local only development: Defaults to true.",
	//	},
	//	&ccli.BoolFlag{
	//		Name:    "enable-tls",
	//		Usage:   "Enable TLS support. Expects cert and key file to be specified",
	//		EnvVars: []string{"VINE_ENABLE_TLS"},
	//	},
	//	&ccli.StringFlag{
	//		Name:    "tls-cert-file",
	//		Usage:   "Path to the TLS Certificate file",
	//		EnvVars: []string{"VINE_TLS_CERT_FILE"},
	//	},
	//	&ccli.StringFlag{
	//		Name:    "tls-key-file",
	//		Usage:   "Path to the TLS Key file",
	//		EnvVars: []string{"VINE_TLS_KEY_FILE"},
	//	},
	//	&ccli.StringFlag{
	//		Name:    "tls-client-ca-file",
	//		Usage:   "Path to the TLS CA file to verify clients against",
	//		EnvVars: []string{"VINE_TLS_CLIENT_CA_FILE"},
	//	},
	//	&ccli.StringFlag{
	//		Name:    "api-address",
	//		Usage:   "Set the api address e.g 0.0.0.0:8080",
	//		EnvVars: []string{"VINE_API_ADDRESS"},
	//	},
	//	&ccli.StringFlag{
	//		Name:    "namespace",
	//		Usage:   "Set the vine service namespace",
	//		EnvVars: []string{"VINE_NAMESPACE"},
	//		Value:   "vine",
	//	},
	//	&ccli.StringFlag{
	//		Name:    "proxy-address",
	//		Usage:   "Proxy requests via the HTTP address specified",
	//		EnvVars: []string{"VINE_PROXY_ADDRESS"},
	//	},
	//	&ccli.StringFlag{
	//		Name:    "web-address",
	//		Usage:   "Set the web UI address e.g 0.0.0.0:8082",
	//		EnvVars: []string{"VINE_WEB_ADDRESS"},
	//	},
	//	&ccli.StringFlag{
	//		Name:    "network",
	//		Usage:   "Set the vine network name: local, go.vine",
	//		EnvVars: []string{"VINE_NETWORK"},
	//	},
	//	&ccli.StringFlag{
	//		Name:    "network-address",
	//		Usage:   "Set the vine network address e.g. :9093",
	//		EnvVars: []string{"VINE_NETWORK_ADDRESS"},
	//	},
	//	&ccli.StringFlag{
	//		Name:    "router-address",
	//		Usage:   "Set the vine router address e.g. :8084",
	//		EnvVars: []string{"VINE_ROUTER_ADDRESS"},
	//	},
	//	&ccli.StringFlag{
	//		Name:    "gateway-address",
	//		Usage:   "Set the vine default gateway address e.g. :9094",
	//		EnvVars: []string{"VINE_GATEWAY_ADDRESS"},
	//	},
	//	&ccli.StringFlag{
	//		Name:    "tunnel-address",
	//		Usage:   "Set the vine tunnel address e.g. :8083",
	//		EnvVars: []string{"VINE_TUNNEL_ADDRESS"},
	//	},
	//	&ccli.StringFlag{
	//		Name:    "api-handler",
	//		Usage:   "Specify the request handler to be used for mapping HTTP requests to services; {api, proxy, rpc}",
	//		EnvVars: []string{"VINE_API_HANDLER"},
	//	},
	//	&ccli.StringFlag{
	//		Name:    "api-namespace",
	//		Usage:   "Set the namespace used by the API e.g. com.example.api",
	//		EnvVars: []string{"VINE_API_NAMESPACE"},
	//	},
	//	&ccli.StringFlag{
	//		Name:    "web-namespace",
	//		Usage:   "Set the namespace used by the Web proxy e.g. com.example.web",
	//		EnvVars: []string{"VINE_WEB_NAMESPACE"},
	//	},
	//	&ccli.StringFlag{
	//		Name:    "web-url",
	//		Usage:   "Set the host used for the web dashboard e.g web.example.com",
	//		EnvVars: []string{"VINE_WEB_HOST"},
	//	},
	//	&ccli.BoolFlag{
	//		Name:    "enable-stats",
	//		Usage:   "Enable stats",
	//		EnvVars: []string{"VINE_ENABLE_STATS"},
	//	},
	//	&ccli.BoolFlag{
	//		Name:    "auto-update",
	//		Usage:   "Enable automatic updates",
	//		EnvVars: []string{"VINE_AUTO_UPDATE"},
	//	},
	//	&ccli.BoolFlag{
	//		Name:    "report-usage",
	//		Usage:   "Report usage statistics",
	//		EnvVars: []string{"VINE_REPORT_USAGE"},
	//		Value:   true,
	//	},
	//	&ccli.StringFlag{
	//		Name:    "env",
	//		Aliases: []string{"e"},
	//		Usage:   "Override environment",
	//		EnvVars: []string{"VINE_ENV"},
	//	},
	//)

	//before := app.Before

	cmd.PreRunE = func(c *cobra.Command, args []string) error {

		//if len(ctx.String("api-handler")) > 0 {
		//	api.Handler = ctx.String("api-handler")
		//}
		//if len(ctx.String("api-address")) > 0 {
		//	api.Address = ctx.String("api-address")
		//}
		//if len(ctx.String("proxy-address")) > 0 {
		//	proxy.Address = ctx.String("proxy-address")
		//}
		//if len(ctx.String("web-address")) > 0 {
		//	web.Address = ctx.String("web-address")
		//}
		//if len(ctx.String("network-address")) > 0 {
		//	network.Address = ctx.String("network-address")
		//}
		//if len(ctx.String("router-address")) > 0 {
		//	router.Address = ctx.String("router-address")
		//}
		//if len(ctx.String("tunnel-address")) > 0 {
		//	tunnel.Address = ctx.String("tunnel-address")
		//}
		//if len(ctx.String("api-namespace")) > 0 {
		//	api.Namespace = ctx.String("api-namespace")
		//}
		//if len(ctx.String("web-namespace")) > 0 {
		//	web.Namespace = ctx.String("web-namespace")
		//}
		//if len(ctx.String("web-host")) > 0 {
		//	web.Host = ctx.String("web-host")
		//}
		//
		//util.SetupCommand(ctx)
		//// now do previous before
		//if err := before(ctx); err != nil {
		//	// DO NOT return this error otherwise the action will fail
		//	// and help will be printed.
		//	fmt.Println(err)
		//	os.Exit(1)
		//	return err
		//}
		//
		//var opts []gostore.Option
		//
		//// the database is not overriden by flag then set it
		//if len(ctx.String("store-database")) == 0 {
		//	opts = append(opts, gostore.Database(cmd.App().Name))
		//}
		//
		//// if the table is not overriden by flag then set it
		//if len(ctx.String("store-table")) == 0 {
		//	table := cmd.App().Name
		//
		//	// if an arg is specified use that as the name
		//	// so each service has its own table preconfigured
		//	if name := ctx.Args().First(); len(name) > 0 {
		//		table = name
		//	}
		//
		//	opts = append(opts, gostore.Table(table))
		//}
		//
		//// TODO: move this entire initialisation elsewhere
		//// maybe in service.Run so all things are configured
		//if len(opts) > 0 {
		//	(*cmd.DefaultCmd.Options().Store).Init(opts...)
		//}
		//
		//// add the system rules if we're using the JWT implementation
		//// which doesn't have access to the rules in the auth service
		//if (*cmd.DefaultCmd.Options().Auth).String() == "jwt" {
		//	for _, rule := range inauth.SystemRules {
		//		if err := (*cmd.DefaultCmd.Options().Auth).Grant(rule); err != nil {
		//			return err
		//		}
		//	}
		//}

		return nil
	}
}

// Init initialised the command line
func Init(options ...vine.Option) {
	root := &cobra.Command{}
	Setup(root, options...)

	cmd.Init(
		cmd.Name(name),
		cmd.Description(description),
		cmd.Command(root),
	)
}

var commandOrder = []string{"api", "new", "init", "build"}

type commands []*cobra.Command

func (s commands) Len() int      { return len(s) }
func (s commands) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s commands) Less(i, j int) bool {
	index := map[string]int{}
	for i, v := range commandOrder {
		index[v] = i
	}
	iVal, ok := index[s[i].Use]
	if !ok {
		iVal = math.MaxInt32
	}
	jVal, ok := index[s[j].Use]
	if !ok {
		jVal = math.MaxInt32
	}
	return iVal < jVal
}

// Setup sets up a cli.App
func Setup(root *cobra.Command, options ...vine.Option) {
	// Add the various commands
	//app.Commands = append(app.Commands, runtime.Commands(options...)...)
	//app.Commands = append(app.Commands, store.Commands(options...)...)
	//app.Commands = append(app.Commands, config.Commands(options...)...)
	root.AddCommand(api.Commands(options...)...)
	//app.Commands = append(app.Commands, broker.Commands(options...)...)
	//app.Commands = append(app.Commands, health.Commands(options...)...)
	//app.Commands = append(app.Commands, proxy.Commands(options...)...)
	//app.Commands = append(app.Commands, router.Commands(options...)...)
	//app.Commands = append(app.Commands, tunnel.Commands(options...)...)
	//app.Commands = append(app.Commands, network.Commands(options...)...)
	//app.Commands = append(app.Commands, registry.Commands(options...)...)
	//app.Commands = append(app.Commands, debug.Commands(options...)...)
	//app.Commands = append(app.Commands, server.Commands(options...)...)
	//app.Commands = append(app.Commands, Commands(options...)...)
	//app.Commands = append(app.Commands, web.Commands(options...)...)
	root.AddCommand(cliMg.Commands()...)
	root.AddCommand(cliRun.Commands()...)
	root.AddCommand(cliBuild.Commands()...)
	//app.Commands = append(app.Commands, auth.Commands()...)
	//app.Commands = append(app.Commands, bot.Commands()...)
	//app.Commands = append(app.Commands, cli.Commands()...)

	sort.Sort(commands(root.Commands()))

	root.Version = version.Version()
	// boot vine runtime
	root.RunE = func(c *cobra.Command, args []string) error {
		return c.Help()
	}

	setup(root)
}
