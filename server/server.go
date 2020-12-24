// Copyright 2020 The vine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lack-io/cli"

	handler "github.com/lack-io/vine/internal/file"
	"github.com/lack-io/vine/internal/platform"
	"github.com/lack-io/vine/internal/update"
	"github.com/lack-io/vine/service"
	"github.com/lack-io/vine/service/config/cmd"
	log "github.com/lack-io/vine/service/logger"
	gorun "github.com/lack-io/vine/service/runtime"
)

var (
	// list of services managed
	services = []string{
		// runtime services
		"config",   // ????
		"auth",     // :8010
		"network",  // :8085
		"runtime",  // :8088
		"registry", // :8000
		"broker",   // :8001
		"store",    // :8002
		"router",   // :8084
		"debug",    // :????
		"proxy",    // :8081
		"api",      // :8080
		"web",      // :8082
		"bot",      // :????
		"init",     // no port, manage self
	}
)

var (
	// Name of the server vineservice
	Name = "go.vine.server"
	// Address is the router vineservice bind address
	Address = ":10001"
)

func Commands(options ...service.Option) []*cli.Command {
	command := &cli.Command{
		Name:  "server",
		Usage: "Run the vine server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "address",
				Usage:   "Set the vine server address :10001",
				EnvVars: []string{"VINE_SERVER_ADDRESS"},
			},
			&cli.BoolFlag{
				Name:  "peer",
				Usage: "Peer with the global network to share services",
			},
			&cli.StringFlag{
				Name:    "profile",
				Usage:   "Set the runtime profile to use for services e.g local, kubernetes, platform",
				EnvVars: []string{"VINE_RUNTIME_PROFILE"},
			},
		},
		Action: func(ctx *cli.Context) error {
			Run(ctx)
			return nil
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

// Run runs the entire platform
func Run(context *cli.Context) error {
	if context.Args().Len() > 0 {
		cli.ShowSubcommandHelp(context)
		os.Exit(1)
	}
	// set default profile
	if len(context.String("profile")) == 0 {
		context.Set("profile", "server")
	}

	// get the network flag
	peer := context.Bool("peer")

	// pass through the environment
	// TODO: perhaps don't do this
	env := []string{"VINE_STORE=file"}
	env = append(env, "VINE_RUNTIME_PROFILE="+context.String("profile"))
	env = append(env, os.Environ()...)

	// connect to the network if specified
	if peer {
		log.Info("Setting global network")

		if v := os.Getenv("VINE_NETWORK_NODES"); len(v) == 0 {
			// set the resolver to use https://vine.mu/network
			env = append(env, "VINE_NETWORK_NODES=network.vine.mu")
			log.Info("Setting default network vine.mu")
		}
		if v := os.Getenv("VINE_NETWORK_TOKEN"); len(v) == 0 {
			// set the network token
			env = append(env, "VINE_NETWORK_TOKEN=vine.mu")
			log.Info("Setting default network token")
		}
	}

	log.Info("Loading core services")

	// create new vine runtime
	muRuntime := cmd.DefaultCmd.Options().Runtime

	// Use default update notifier
	if context.Bool("auto-update") {
		updateURL := context.String("update-url")
		if len(updateURL) == 0 {
			updateURL = update.DefaultURL
		}

		options := []gorun.Option{
			gorun.WithScheduler(update.NewScheduler(updateURL, platform.Version)),
		}
		(*muRuntime).Init(options...)
	}

	for _, service := range services {
		name := service

		if namespace := context.String("namespace"); len(namespace) > 0 {
			name = fmt.Sprintf("%s.%s", namespace, service)
		}

		log.Infof("Registering %s", name)
		// @todo this is a hack
		envs := env
		switch service {
		case "proxy", "web", "api":
			envs = append(envs, "VINE_AUTH=service")
		}

		// runtime based on environment we run the service in
		args := []gorun.CreateOption{
			gorun.WithCommand(os.Args[0]),
			gorun.WithArgs(service),
			gorun.WithEnv(envs),
			gorun.WithOutput(os.Stdout),
			gorun.WithRetries(10),
		}

		// NOTE: we use Version right now to check for the latest release
		muService := &gorun.Service{Name: name, Version: platform.Version}
		if err := (*muRuntime).Create(muService, args...); err != nil {
			log.Errorf("Failed to create runtime enviroment: %v", err)
			return err
		}
	}

	log.Info("Starting service runtime")

	// start the runtime
	if err := (*muRuntime).Start(); err != nil {
		log.Fatal(err)
		return err
	}

	log.Info("Service runtime started")

	// TODO: should we launch the console?
	// start the console
	// cli.Init(context)

	server := service.NewService(
		service.Name(Name),
		service.Address(Address),
	)

	// @todo make this configurable
	uploadDir := filepath.Join(os.TempDir(), "vine", "uploads")
	os.MkdirAll(uploadDir, 0777)
	handler.RegisterHandler(server.Server(), uploadDir)
	// start the server
	server.Run()

	log.Info("Stopping service runtime")

	// stop all the things
	if err := (*muRuntime).Stop(); err != nil {
		log.Fatal(err)
		return err
	}

	log.Info("Service runtime shutdown")

	// exit success
	os.Exit(0)
	return nil
}
