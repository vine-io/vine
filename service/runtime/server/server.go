// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"os"

	"github.com/lack-io/cli"

	pb "github.com/lack-io/vine/proto/runtime"
	"github.com/lack-io/vine/service"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/runtime"
	goruntime "github.com/lack-io/vine/service/runtime"
	"github.com/lack-io/vine/service/runtime/server/handler"
	"github.com/lack-io/vine/service/runtime/server/manager"
)

var (
	// name of the runtime
	name = "runtime"
	// address of the runtime
	address = ":8088"

	// Flags specific to the runtime service
	Flags = []cli.Flag{
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
	}
)

// Run the runtime service
func Run(ctx *cli.Context) error {
	if len(ctx.String("address")) > 0 {
		address = ctx.String("address")
	}

	if len(ctx.String("server-name")) > 0 {
		name = ctx.String("server-name")
	}

	var srvOpts []service.Option
	if len(address) > 0 {
		srvOpts = append(srvOpts, service.Address(address))
	}

	// create runtime
	if ctx.IsSet("source") {
		runtime.DefaultRuntime.Init(goruntime.WithSource(ctx.String("source")))
	}

	// append name
	srvOpts = append(srvOpts, service.Name(name))

	// new service
	srv := service.New(srvOpts...)

	// create a new runtime manager
	manager := manager.New()

	// start the manager
	if err := manager.Start(); err != nil {
		log.Errorf("failed to start: %s", err)
		os.Exit(1)
	}

	// register the handlers
	pb.RegisterRuntimeHandler(srv.Server(), &handler.Runtime{Runtime: manager})
	pb.RegisterBuildHandler(srv.Server(), new(handler.Build))
	pb.RegisterSourceHandler(srv.Server(), new(handler.Source))

	// start runtime service
	if err := srv.Run(); err != nil {
		log.Errorf("error running service: %v", err)
	}

	// stop the manager
	if err := manager.Stop(); err != nil {
		log.Errorf("failed to stop: %s", err)
		os.Exit(1)
	}

	return nil
}
