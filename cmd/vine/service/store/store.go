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

package store

import (
	"fmt"

	"github.com/lack-io/cli"
	"github.com/lack-io/vine"
	cli3 "github.com/lack-io/vine/cmd/vine/app/cli"
	"github.com/lack-io/vine/cmd/vine/service/store/handler"
	log "github.com/lack-io/vine/lib/logger"
	"github.com/lack-io/vine/lib/store"
	pb "github.com/lack-io/vine/proto/services/store"
	"github.com/lack-io/vine/util/helper"
)

var (
	// Name of the store service
	Name = "go.vine.store"
	// Address is the store address
	Address = ":8002"
)

// Run runs the vine server
func Run(ctx *cli.Context, svcOpts ...vine.Option) {

	if len(ctx.String("server-name")) > 0 {
		Name = ctx.String("server-name")
	}
	if len(ctx.String("address")) > 0 {
		Address = ctx.String("address")
	}

	// Initialise service
	svc := vine.NewService(
		vine.Name(Name),
	)

	// the store handler
	storeHandler := &handler.Store{
		Default: svc.Options().Store,
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
			return nil, fmt.Errorf("vine store couldn't store new database in internal table: %w", err)
		}
		if err := storeHandler.Default.Write(&store.Record{
			Key:   "tables/" + database + "/" + table,
			Value: []byte{},
		}, store.WriteTo("vine", "internal")); err != nil {
			return nil, fmt.Errorf("vine store couldn't store new table in internal table: %w", err)
		}

		return storeHandler.Default, nil
	}

	pb.RegisterStoreHandler(svc.Server(), storeHandler)

	// start the service
	if err := svc.Run(); err != nil {
		log.Fatal(err)
	}
}

// Commands is the cli interface for the store service
func Commands(options ...vine.Option) []*cli.Command {
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
		Subcommands: cli3.StoreCommands(),
	}

	return []*cli.Command{command}
}
