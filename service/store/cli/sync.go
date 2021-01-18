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

package cli

import (
	"github.com/lack-io/cli"
	"github.com/pkg/errors"
)

// Sync is the entrypoint for vine vine sync
func Sync(ctx *cli.Context) error {
	from, to, err := makeStores(ctx)
	if err != nil {
		return errors.Wrap(err, "Sync")
	}

	keys, err := from.List()
	if err != nil {
		return errors.Wrapf(err, "couldn't list from vine %s", from.String())
	}
	for _, k := range keys {
		r, err := from.Read(k)
		if err != nil {
			return errors.Wrapf(err, "couldn't read %s from vine %s", k, from.String())
		}
		if len(r) != 1 {
			return errors.Errorf("received multiple records reading %s from %s", k, from.String())
		}
		err = to.Write(r[0])
		if err != nil {
			return errors.Wrapf(err, "couldn't write %s to vine %s", k, to.String())
		}
	}
	return nil
}

// SyncFlags are the flags for vine vine sync
var SyncFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "from-backend",
		Usage:   "Backend to sync from",
		EnvVars: []string{"VINE_STORE_FROM"},
	},
	&cli.StringFlag{
		Name:    "from-nodes",
		Usage:   "Nodes to sync from",
		EnvVars: []string{"VINE_STORE_FROM_NODES"},
	},
	&cli.StringFlag{
		Name:    "from-database",
		Usage:   "Database to sync from",
		EnvVars: []string{"VINE_STORE_FROM_DATABASE"},
	},
	&cli.StringFlag{
		Name:    "from-table",
		Usage:   "Table to sync from",
		EnvVars: []string{"VINE_STORE_FROM_TABLE"},
	},
	&cli.StringFlag{
		Name:    "to-backend",
		Usage:   "Backend to sync to",
		EnvVars: []string{"VINE_STORE_TO"},
	},
	&cli.StringFlag{
		Name:    "to-nodes",
		Usage:   "Nodes to sync to",
		EnvVars: []string{"VINE_STORE_TO_NODES"},
	},
	&cli.StringFlag{
		Name:    "to-database",
		Usage:   "Database to sync to",
		EnvVars: []string{"VINE_STORE_TO_DATABASE"},
	},
	&cli.StringFlag{
		Name:    "to-table",
		Usage:   "Table to sync to",
		EnvVars: []string{"VINE_STORE_TO_TABLE"},
	},
}
