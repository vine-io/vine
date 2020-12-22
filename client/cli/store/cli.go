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

// Package cli implements the `vine store` subcommands
// for example:
//   vine store snapshot
//   vine store restore
//   vine store sync
package cli

import (
	"github.com/lack-io/cli"

	"github.com/lack-io/vine/cmd"
	"github.com/lack-io/vine/internal/helper"
)

func init() {
	cmd.Register(&cli.Command{
		Name:   "store",
		Usage:  "Commands for accessing the store",
		Action: helper.UnexpectedSubcommand,
		Subcommands: []*cli.Command{
			{
				Name:      "read",
				Usage:     "read a record from the store",
				UsageText: `vine store read [options] key`,
				Action:    read,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "database",
						Aliases: []string{"d"},
						Usage:   "database to write to",
						Value:   "vine",
					},
					&cli.StringFlag{
						Name:    "table",
						Aliases: []string{"t"},
						Usage:   "table to write to",
						Value:   "vine",
					},
					&cli.BoolFlag{
						Name:    "prefix",
						Aliases: []string{"p"},
						Usage:   "read prefix",
						Value:   false,
					},
					&cli.UintFlag{
						Name:    "limit",
						Aliases: []string{"l"},
						Usage:   "list limit",
					},
					&cli.UintFlag{
						Name:    "offset",
						Aliases: []string{"o"},
						Usage:   "list offset",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "show keys and headers (only values shown by default)",
						Value:   false,
					},
					&cli.StringFlag{
						Name:  "output",
						Usage: "output format (json, table)",
						Value: "table",
					},
				},
			},
			{
				Name:      "list",
				Usage:     "list all keys from a store",
				UsageText: `vine store list [options]`,
				Action:    list,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "database",
						Aliases: []string{"d"},
						Usage:   "database to list from",
						Value:   "vine",
					},
					&cli.StringFlag{
						Name:    "table",
						Aliases: []string{"t"},
						Usage:   "table to write to",
						Value:   "vine",
					},
					&cli.StringFlag{
						Name:  "output",
						Usage: "output format (json)",
					},
					&cli.BoolFlag{
						Name:    "prefix",
						Aliases: []string{"p"},
						Usage:   "list prefix",
						Value:   false,
					},
					&cli.UintFlag{
						Name:    "limit",
						Aliases: []string{"l"},
						Usage:   "list limit",
					},
					&cli.UintFlag{
						Name:    "offset",
						Aliases: []string{"o"},
						Usage:   "list offset",
					},
				},
			},
			{
				Name:      "write",
				Usage:     "write a record to the store",
				UsageText: `vine store write [options] key value`,
				Action:    write,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "expiry",
						Aliases: []string{"e"},
						Usage:   "expiry in time.ParseDuration format",
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "database",
						Aliases: []string{"d"},
						Usage:   "database to write to",
						Value:   "vine",
					},
					&cli.StringFlag{
						Name:    "table",
						Aliases: []string{"t"},
						Usage:   "table to write to",
						Value:   "vine",
					},
				},
			},
			{
				Name:      "delete",
				Usage:     "delete a key from the store",
				UsageText: `vine store delete [options] key`,
				Action:    delete,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "database",
						Usage: "database to delete from",
						Value: "vine",
					},
					&cli.StringFlag{
						Name:  "table",
						Usage: "table to delete from",
						Value: "vine",
					},
				},
			},
			{
				Name:   "databases",
				Usage:  "List all databases known to the store service",
				Action: databases,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "store",
						Usage: "store service to call",
						Value: "store",
					},
				},
			},
			{
				Name:   "tables",
				Usage:  "List all tables in the specified database known to the store service",
				Action: tables,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "store",
						Usage: "store service to call",
						Value: "store",
					},
					&cli.StringFlag{
						Name:    "database",
						Aliases: []string{"d"},
						Usage:   "database to list tables of",
						Value:   "vine",
					},
				},
			},
			{
				Name:   "snapshot",
				Usage:  "Back up a store",
				Action: snapshot,
				Flags: append(CommonFlags,
					&cli.StringFlag{
						Name:    "destination",
						Usage:   "Backup destination",
						Value:   "file:///tmp/store-snapshot",
						EnvVars: []string{"VINE_SNAPSHOT_DESTINATION"},
					},
				),
			},
			{
				Name:   "sync",
				Usage:  "Copy all records of one store into another store",
				Action: sync,
				Flags:  SyncFlags,
			},
			{
				Name:   "restore",
				Usage:  "restore a store snapshot",
				Action: restore,
				Flags: append(CommonFlags,
					&cli.StringFlag{
						Name:  "source",
						Usage: "Backup source",
						Value: "file:///tmp/store-snapshot",
					},
				),
			},
		},
	})
}

// CommonFlags are flags common to cli commands snapshot and restore
var CommonFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "nodes",
		Usage:   "Comma separated list of Nodes to pass to the store backend",
		EnvVars: []string{"VINE_STORE_NODES"},
	},
	&cli.StringFlag{
		Name:    "database",
		Usage:   "Database option to pass to the store backend",
		EnvVars: []string{"VINE_STORE_DATABASE"},
	},
	&cli.StringFlag{
		Name:    "table",
		Usage:   "Table option to pass to the store backend",
		EnvVars: []string{"VINE_STORE_TABLE"},
	},
}
