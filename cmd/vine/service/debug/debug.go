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

// Package debug implements metrics/logging/introspection/... of vine services
package debug

import (
	"github.com/lack-io/cli"

	pblog "github.com/lack-io/vine/proto/debug/log"
	pbstats "github.com/lack-io/vine/proto/debug/stats"
	pbtrace "github.com/lack-io/vine/proto/debug/trace"
	"github.com/lack-io/vine/service"
	"github.com/lack-io/vine/util/debug/log"

	// "github.com/lack-io/vine/debug/log/kubernetes"
	dservice "github.com/lack-io/vine/service/debug"
	ulog "github.com/lack-io/vine/service/logger"

	logHandler "github.com/lack-io/vine/cmd/vine/service/debug/log"
	statshandler "github.com/lack-io/vine/cmd/vine/service/debug/stats"
	tracehandler "github.com/lack-io/vine/cmd/vine/service/debug/trace"
)

var (
	// Name of the service
	Name = "go.vine.debug"
	// Address of the service
	Address = ":8089"
)

func Run(ctx *cli.Context, srvOpts ...service.Option) {

	// Init plugins
	for _, p := range Plugins() {
		p.Init(ctx)
	}

	if len(ctx.String("address")) > 0 {
		Address = ctx.String("address")
	}

	if len(ctx.String("server-name")) > 0 {
		Name = ctx.String("server-name")
	}

	if len(Address) > 0 {
		srvOpts = append(srvOpts, service.Address(Address))
	}

	// append name
	srvOpts = append(srvOpts, service.Name(Name))

	// new service
	srv := service.NewService(srvOpts...)

	// default log initialiser
	newLog := func(service string) log.Log {
		// service log calls the actual service for the log
		return dservice.NewLog(
			// log with service name
			log.Name(service),
		)
	}

	source := ctx.String("log")
	switch source {
	case "service":
		newLog = func(service string) log.Log {
			return dservice.NewLog(
				log.Name(service),
			)
		}
		//case "kubernetes":
		//	newLog = func(service string) log.Log {
		//		return kubernetes.NewLog(
		//			log.Name(service),
		//		)
		//	}
		//}

		done := make(chan bool)
		defer func() {
			close(done)
		}()

		// create a service cache
		c := newCache(done)

		// log handler
		lgHandler := &logHandler.Log{
			// create the log map
			Logs: make(map[string]log.Log),
			// Create the new func
			New: newLog,
		}

		// Register the logs handler
		pblog.RegisterLogHandler(srv.Server(), lgHandler)

		// stats handler
		statsHandler, err := statshandler.New(done, ctx.Int("window"), c.services)
		if err != nil {
			ulog.Fatal(err)
		}

		// stats handler
		traceHandler, err := tracehandler.New(done, ctx.Int("window"), c.services)
		if err != nil {
			ulog.Fatal(err)
		}

		// Register the stats handler
		pbstats.RegisterStatsHandler(srv.Server(), statsHandler)
		// register trace handler
		pbtrace.RegisterTraceHandler(srv.Server(), traceHandler)

		// TODO: implement debug service for k8s cruft

		// start debug service
		if err := srv.Run(); err != nil {
			ulog.Fatal(err)
		}
	}
}

// Commands populates the debug commands
func Commands(options ...service.Option) []*cli.Command {
	command := []*cli.Command{
		{
			Name:  "debug",
			Usage: "Run the vine debug service",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "address",
					Usage:   "Set the registry http address e.g 0.0.0.0:8089",
					EnvVars: []string{"VINE_SERVER_ADDRESS"},
				},
				&cli.IntFlag{
					Name:    "window",
					Usage:   "Specifies how many seconds of stats snapshots to retain in memory",
					EnvVars: []string{"VINE_DEBUG_WINDOW"},
					Value:   60,
				},
			},
			Action: func(ctx *cli.Context) error {
				Run(ctx, options...)
				return nil
			},
		},
		{
			Name:  "trace",
			Usage: "Get tracing info from a service",
			Action: func(ctx *cli.Context) error {
				getTrace(ctx, options...)
				return nil
			},
		},
	}

	for _, p := range Plugins() {
		if cmds := p.Commands(); len(cmds) > 0 {
			command[0].Subcommands = append(command[0].Subcommands, cmds...)
		}

		if flags := p.Flags(); len(flags) > 0 {
			command[0].Flags = append(command[0].Flags, flags...)
		}
	}

	return command
}
