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

package cli

import (
	"fmt"

	"github.com/lack-io/cli"
)

// Sync is the entrypoint for vine vine sync
func Sync(ctx *cli.Context) error {
	from, to, err := makeStores(ctx)
	if err != nil {
		return fmt.Errorf("Sync: %w", err)
	}

	keys, err := from.List()
	if err != nil {
		return fmt.Errorf("couldn't list from vine %s: %w", from.String(), err)
	}
	for _, k := range keys {
		r, err := from.Read(k)
		if err != nil {
			return fmt.Errorf("couldn't read %s from vine %s: %w", k, from.String(), err)
		}
		if len(r) != 1 {
			return fmt.Errorf("received multiple records reading %s from %s", k, from.String())
		}
		err = to.Write(r[0])
		if err != nil {
			return fmt.Errorf("couldn't write %s to vine %s: %w", k, to.String(), err)
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
