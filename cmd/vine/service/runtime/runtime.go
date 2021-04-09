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

// Package runtime is the vine runtime
package runtime

import (
	"os"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine"
	"github.com/lack-io/vine/cmd/vine/service/runtime/handler"
	"github.com/lack-io/vine/cmd/vine/service/runtime/manager"
	"github.com/lack-io/vine/cmd/vine/service/runtime/profile"
	pb "github.com/lack-io/vine/proto/services/runtime"
	"github.com/lack-io/vine/service/config/cmd"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/runtime"
)

var (
	// Name of the runtime
	Name = "go.vine.runtime"
	// Address of the runtime
	Address = ":8088"
)

// Run the runtime service
func Run(ctx *cli.Context, svcOpts ...vine.Option) {

	// Get the profile
	var prof []string
	switch ctx.String("profile") {
	case "local":
		prof = profile.Local()
	case "server":
		prof = profile.Server()
	case "kubernetes":
		prof = profile.Kubernetes()
	case "platform":
		prof = profile.Platform()
	}

	if len(ctx.String("address")) > 0 {
		Address = ctx.String("address")
	}

	if len(ctx.String("server-name")) > 0 {
		Name = ctx.String("server-name")
	}

	if len(Address) > 0 {
		svcOpts = append(svcOpts, vine.Address(Address))
	}

	// create runtime
	muRuntime := *cmd.DefaultCmd.Options().Runtime
	if ctx.IsSet("source") {
		muRuntime.Init(runtime.WithSource(ctx.String("source")))
	}

	// append name
	svcOpts = append(svcOpts, vine.Name(Name))

	// new service
	svc := vine.NewService(svcOpts...)

	// create a new runtime manager
	manager := manager.New(muRuntime,
		manager.Store(svc.Options().Store),
		manager.Profile(prof),
	)

	// start the manager
	if err := manager.Start(); err != nil {
		log.Errorf("failed to start: %s", err)
		os.Exit(1)
	}

	// register the runtime handler
	pb.RegisterRuntimeHandler(svc.Server(), &handler.Runtime{
		// Client to publish events
		Client: vine.NewEvent("go.vine.runtime.events", svc.Client()),
		// using the vine runtime
		Runtime: manager,
	})

	// start runtime service
	if err := svc.Run(); err != nil {
		log.Errorf("error running service: %v", err)
	}

	// stop the manager
	if err := manager.Stop(); err != nil {
		log.Errorf("failed to stop: %s", err)
		os.Exit(1)
	}
}

// Flags is shared flags so we don't have to continually re-add
func Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "source",
			Usage: "Set the source url of the service e.g github.com/vine/services",
		},
		&cli.StringFlag{
			Name:  "image",
			Usage: "Set the image to use for the container",
		},
		&cli.StringFlag{
			Name:  "command",
			Usage: "Command to exec",
		},
		&cli.StringFlag{
			Name:  "args",
			Usage: "Command args",
		},
		&cli.StringFlag{
			Name:  "type",
			Usage: "The type of service operate on",
		},
		&cli.StringSliceFlag{
			Name:  "env-vars",
			Usage: "Set the environment variables e.g. foo=bar",
		},
	}
}

func Commands(options ...vine.Option) []*cli.Command {
	command := []*cli.Command{
		{
			Name:  "runtime",
			Usage: "Run the vine runtime",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "address",
					Usage:   "Set the registry http address e.g 0.0.0.0:8088",
					EnvVars: []string{"VINE_SERVER_ADDRESS"},
				},
				&cli.StringFlag{
					Name:    "profile",
					Usage:   "Set the runtime profile to use for services e.g local, kubernetes, platform",
					EnvVars: []string{"VINE_RUNTIME_PROFILE"},
				},
				&cli.StringFlag{
					Name:    "source",
					Usage:   "Set the runtime source, e.g. vine/services",
					EnvVars: []string{"VINE_RUNTIME_SOURCE"},
				},
				&cli.IntFlag{
					Name:    "retries",
					Usage:   "Set the max retries per service",
					EnvVars: []string{"VINE_RUNTIME_RETRIES"},
				},
			},
			Action: func(ctx *cli.Context) error {
				Run(ctx, options...)
				return nil
			},
		},
		{
			// In future we'll also have `vine run [x]` hence `vine run service` requiring "service"
			Name:  "run",
			Usage: RunUsage,
			Description: `Examples:
			vine run github.com/vine/examples/helloworld
			vine run .  # deploy local folder to your local vine server
			vine run ../path/to/folder # deploy local folder to your local vine server
			vine run helloworld # deploy latest version, translates to vine run github.com/vine/services/helloworld
			vine run helloworld@9342934e6180 # deploy certain version
			vine run helloworld@branchname	# deploy certain branch`,
			Flags: Flags(),
			Action: func(ctx *cli.Context) error {
				runService(ctx, options...)
				return nil
			},
		},
		{
			Name:  "update",
			Usage: UpdateUsage,
			Description: `Examples:
			vine update github.com/vine/examples/helloworld
			vine update .  # deploy local folder to your local vine server
			vine update ../path/to/folder # deploy local folder to your local vine server
			vine update helloworld # deploy master branch, translates to vine update github.com/vine/services/helloworld
			vine update helloworld@branchname	# deploy certain branch`,
			Flags: Flags(),
			Action: func(ctx *cli.Context) error {
				updateService(ctx, options...)
				return nil
			},
		},
		{
			Name:  "kill",
			Usage: KillUsage,
			Flags: Flags(),
			Description: `Examples:
			vine kill github.com/vine/examples/helloworld
			vine kill .  # kill service deployed from local folder
			vine kill ../path/to/folder # kill service deployed from local folder
			vine kill helloworld # kill serviced deployed from master branch, translates to vine kill github.com/vine/services/helloworld
			vine kill helloworld@branchname	# kill service deployed from certain branch`,
			Action: func(ctx *cli.Context) error {
				killService(ctx, options...)
				return nil
			},
		},
		{
			Name:  "status",
			Usage: GetUsage,
			Flags: Flags(),
			Action: func(ctx *cli.Context) error {
				getService(ctx, options...)
				return nil
			},
		},
		{
			Name:  "logs",
			Usage: "Get logs for a service",
			Flags: logFlags(),
			Action: func(ctx *cli.Context) error {
				getLogs(ctx, options...)
				return nil
			},
		},
	}

	return command
}
