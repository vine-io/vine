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

// Package debug implements metrics/logging/introspection/... of vine services
package debug

import (
	"github.com/lack-io/cli"

	"github.com/lack-io/vine"
	logHandler "github.com/lack-io/vine/cmd/vine/service/debug/log"
	statshandler "github.com/lack-io/vine/cmd/vine/service/debug/stats"
	tracehandler "github.com/lack-io/vine/cmd/vine/service/debug/trace"
	dservice "github.com/lack-io/vine/lib/debug"
	"github.com/lack-io/vine/lib/debug/log"
	ulog "github.com/lack-io/vine/lib/logger"
	pblog "github.com/lack-io/vine/proto/services/debug/log"
	pbstats "github.com/lack-io/vine/proto/services/debug/stats"
	pbtrace "github.com/lack-io/vine/proto/services/debug/trace"
)

var (
	// Name of the service
	Name = "go.vine.debug"
	// Address of the service
	Address = ":8089"
)

func Run(ctx *cli.Context, svcOpts ...vine.Option) {

	if len(ctx.String("address")) > 0 {
		Address = ctx.String("address")
	}

	if len(ctx.String("server-name")) > 0 {
		Name = ctx.String("server-name")
	}

	if len(Address) > 0 {
		svcOpts = append(svcOpts, vine.Address(Address))
	}

	// append name
	svcOpts = append(svcOpts, vine.Name(Name))

	// new service
	svc := vine.NewService(svcOpts...)

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
		pblog.RegisterLogHandler(svc.Server(), lgHandler)

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
		pbstats.RegisterStatsHandler(svc.Server(), statsHandler)
		// register trace handler
		pbtrace.RegisterTraceHandler(svc.Server(), traceHandler)

		// TODO: implement debug service for k8s cruft

		// start debug service
		if err := svc.Run(); err != nil {
			ulog.Fatal(err)
		}
	}
}

// Commands populates the debug commands
func Commands(options ...vine.Option) []*cli.Command {
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

	return command
}
