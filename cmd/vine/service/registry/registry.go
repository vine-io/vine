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

// Package registry is the vine registry
package registry

import (
	"context"
	"time"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine"
	rcli "github.com/lack-io/vine/cmd/vine/client/cli"
	"github.com/lack-io/vine/cmd/vine/service/registry/handler"
	regpb "github.com/lack-io/vine/proto/apis/registry"
	regsvcpb "github.com/lack-io/vine/proto/services/registry"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/registry"
	"github.com/lack-io/vine/util/helper"
)

var (
	// Name of the registry
	Name = "go.vine.registry"
	// The address of the registry
	Address = ":8000"
	// Topic to publish registry events to
	Topic = "go.vine.registry.events"
)

// Sub processes registry events
type subscriber struct {
	// id is registry id
	Id string
	// registry is service registry
	Registry registry.Registry
}

// Process processes registry events
func (s *subscriber) Process(ctx context.Context, event *regpb.Event) error {
	if event.Id == s.Id {
		log.Tracef("skipping own %s event: %s for: %s", event.Type.String(), event.Id, event.Service.Name)
		return nil
	}

	log.Debugf("received %s event from: %s for: %s", event.Type.String(), event.Id, event.Service.Name)

	// no service
	if event.Service == nil {
		return nil
	}

	// decode protobuf to registry.Service
	svc := event.Service

	// default ttl to 1 minute
	ttl := time.Minute

	// set ttl if it exists
	if opts := event.Service.Options; opts != nil {
		if opts.Ttl > 0 {
			ttl = time.Duration(opts.Ttl) * time.Second
		}
	}

	switch event.Type {
	case regpb.EventType_Create, regpb.EventType_Update:
		log.Debugf("registering service: %s", svc.Name)
		if err := s.Registry.Register(svc, registry.RegisterTTL(ttl)); err != nil {
			log.Debugf("failed to register service: %s", svc.Name)
			return err
		}
	case regpb.EventType_Delete:
		log.Debugf("deregistering service: %s", svc.Name)
		if err := s.Registry.Deregister(svc); err != nil {
			log.Debugf("failed to deregister service: %s", svc.Name)
			return err
		}
	}

	return nil
}

func Run(ctx *cli.Context, svcOpts ...vine.Option) {

	if len(ctx.String("server-name")) > 0 {
		Name = ctx.String("server-name")
	}
	if len(ctx.String("address")) > 0 {
		Address = ctx.String("address")
	}

	// Init plugins
	for _, p := range Plugins() {
		p.Init(ctx)
	}

	// service opts
	svcOpts = append(svcOpts, vine.Name(Name))
	if i := time.Duration(ctx.Int("register-ttl")); i > 0 {
		svcOpts = append(svcOpts, vine.RegisterTTL(i*time.Second))
	}
	if i := time.Duration(ctx.Int("register-interval")); i > 0 {
		svcOpts = append(svcOpts, vine.RegisterInterval(i*time.Second))
	}

	// set address
	if len(Address) > 0 {
		svcOpts = append(svcOpts, vine.Address(Address))
	}

	// new service
	svc := vine.NewService(svcOpts...)
	// get server id
	id := svc.Server().Options().Id

	// register the handler
	regsvcpb.RegisterRegistryHandler(svc.Server(), &handler.Registry{
		Id:        id,
		Publisher: vine.NewEvent(Topic, svc.Client()),
		Registry:  svc.Options().Registry,
		Auth:      svc.Options().Auth,
	})

	// run the service
	if err := svc.Run(); err != nil {
		log.Fatal(err)
	}
}

func Commands(options ...vine.Option) []*cli.Command {
	command := &cli.Command{
		Name:  "registry",
		Usage: "Run the service registry",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "address",
				Usage:   "Set the registry http address e.g 0.0.0.0:8000",
				EnvVars: []string{"VINE_SERVER_ADDRESS"},
			},
		},
		Action: func(ctx *cli.Context) error {
			if err := helper.UnexpectedSubcommand(ctx); err != nil {
				return err
			}
			Run(ctx, options...)
			return nil
		},
		Subcommands: rcli.RegistryCommands(),
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
