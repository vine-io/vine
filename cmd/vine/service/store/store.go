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

package store

import (
	"github.com/lack-io/cli"
	"github.com/pkg/errors"

	mcli "github.com/lack-io/vine/client/cli"
	"github.com/lack-io/vine/cmd/vine/service/store/handler"
	pb "github.com/lack-io/vine/proto/store"
	"github.com/lack-io/vine/service"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/store"
	"github.com/lack-io/vine/util/helper"
)

var (
	// Name of the store service
	Name = "go.vine.store"
	// Address is the store address
	Address = ":8002"
)

// Run runs the vine server
func Run(ctx *cli.Context, srvOpts ...service.Option) {

	// Init plugins
	for _, p := range Plugins() {
		p.Init(ctx)
	}

	if len(ctx.String("server-name")) > 0 {
		Name = ctx.String("server-name")
	}
	if len(ctx.String("address")) > 0 {
		Address = ctx.String("address")
	}

	// Initialise service
	srv := service.NewService(
		service.Name(Name),
	)

	// the store handler
	storeHandler := &handler.Store{
		Default: srv.Options().Store,
		Stores:  make(map[string]bool),
	}

	table := "store"
	if v := ctx.String("store-table"); len(v) > 0 {
		table = v
	}

	// set to store table
	storeHandler.Default.Init(
		store.Table(table),
	)

	backend := storeHandler.Default.String()
	options := storeHandler.Default.Options()

	log.Infof("Initialising the [%s] store with opts: %+v", backend, options)

	// set the new store initialiser
	storeHandler.New = func(database string, table string) (store.Store, error) {
		// Record the new database and table in the internal store
		if err := storeHandler.Default.Write(&store.Record{
			Key:   "databases/" + database,
			Value: []byte{},
		}, store.WriteTo("vine", "internal")); err != nil {
			return nil, errors.Wrap(err, "vine store couldn't store new database in internal table")
		}
		if err := storeHandler.Default.Write(&store.Record{
			Key:   "tables/" + database + "/" + table,
			Value: []byte{},
		}, store.WriteTo("vine", "internal")); err != nil {
			return nil, errors.Wrap(err, "vine store couldn't store new table in internal table")
		}

		return storeHandler.Default, nil
	}

	pb.RegisterStoreHandler(srv.Server(), storeHandler)

	// start the service
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}

// Commands is the cli interface for the store service
func Commands(options ...service.Option) []*cli.Command {
	command := &cli.Command{
		Name:  "store",
		Usage: "Run the vine store service",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "address",
				Usage:   "Set the vine tunnel address :8002",
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
		Subcommands: mcli.StoreCommands(),
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
