// Copyright 2020 lack
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

// Package health is a healthchecking sidecar
package health

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine"
	mcli "github.com/lack-io/vine/cmd/vine/client/cli"
	proto "github.com/lack-io/vine/proto/services/debug"
	"github.com/lack-io/vine/service/client"
	log "github.com/lack-io/vine/service/logger"
	qcli "github.com/lack-io/vine/util/command/cli"
)

var (
	healthAddress = ":8088"
	serverAddress string
	serverName    string
)

func Run(ctx *cli.Context) {

	// just check service health
	if ctx.Args().Len() > 0 {
		mcli.Print(qcli.QueryHealth)(ctx)
		return
	}

	serverName = ctx.String("check-service")
	serverAddress = ctx.String("check-address")

	if addr := ctx.String("address"); len(addr) > 0 {
		healthAddress = addr
	}

	if len(healthAddress) == 0 {
		log.Fatal("health address not set")
	}
	if len(serverName) == 0 {
		log.Fatal("service name not set")
	}
	if len(serverAddress) == 0 {
		log.Fatal("service address not set")
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		req := client.NewRequest(serverName, "Debug.Health", &proto.HealthRequest{})
		rsp := &proto.HealthResponse{}

		err := client.Call(context.TODO(), req, rsp, client.WithAddress(serverAddress))
		if err != nil || rsp.Status != "ok" {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "NOT_HEALTHY")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	log.Infof("Health check running at %s/health", healthAddress)
	log.Infof("Health check defined for %s at %s", serverName, serverAddress)

	if err := http.ListenAndServe(healthAddress, nil); err != nil {
		log.Fatal(err)
	}
}

func Commands(options ...vine.Option) []*cli.Command {
	command := &cli.Command{
		Name:  "health",
		Usage: "Check the health of a service",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "address",
				Usage:   "Set the address exposed for the http server e.g :8088",
				EnvVars: []string{"VINE_HEALTH_ADDRESS"},
			},
			&cli.StringFlag{
				Name:    "check-service",
				Usage:   "Name of the service to query",
				EnvVars: []string{"VINE_HEALTH_CHECK_SERVICE"},
			},
			&cli.StringFlag{
				Name:    "check-address",
				Usage:   "Set the service address to query",
				EnvVars: []string{"VINE_HEALTH_CHECK_ADDRESS"},
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
