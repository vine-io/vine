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
