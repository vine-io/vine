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

// Package cli implements the `vine store` subcommands
// for example:
//   vine store snapshot
//   vine store restore
//   vine store sync
package cli

import (
	"github.com/lack-io/cli"
)

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
