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

package cli

import (
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/imdario/mergo"
	"github.com/lack-io/cli"

	"github.com/lack-io/vine/service/config/cmd"
	"github.com/lack-io/vine/service/config/source"
)

type cliSource struct {
	opts source.Options
	ctx  *cli.Context
}

func (c *cliSource) Read() (*source.ChangeSet, error) {
	var changes map[string]interface{}

	// directly using app cli flags, to access default values of not specified options
	for _, f := range c.ctx.App.Flags {
		name := f.Names()[0]
		tmp := toEntry(name, c.ctx.Generic(name))
		mergo.Map(&changes, tmp) // need to sort error handling
	}

	b, err := c.opts.Encoder.Encode(changes)
	if err != nil {
		return nil, err
	}

	cs := &source.ChangeSet{
		Format:    c.opts.Encoder.String(),
		Data:      b,
		Timestamp: time.Now(),
		Source:    c.String(),
	}
	cs.CheckSum = cs.Sum()

	return cs, nil
}

func toEntry(name string, v interface{}) map[string]interface{} {
	n := strings.ToLower(name)
	keys := strings.FieldsFunc(n, split)
	reverse(keys)
	tmp := make(map[string]interface{})
	for i, k := range keys {
		if i == 0 {
			tmp[k] = v
			continue
		}

		tmp = map[string]interface{}{k: tmp}
	}
	return tmp
}

func reverse(ss []string) {
	for i := len(ss)/2 - 1; i >= 0; i-- {
		opp := len(ss) - 1 - i
		ss[i], ss[opp] = ss[opp], ss[i]
	}
}

func split(r rune) bool {
	return r == '-' || r == '_'
}

func (c *cliSource) Watch() (source.Watcher, error) {
	return source.NewNoopWatcher()
}

// Write is unsupported
func (c *cliSource) Write(cs *source.ChangeSet) error {
	return nil
}

func (c *cliSource) String() string {
	return "cli"
}

// NewSource returns a config source for integrating parsed flags from a vine/cli.Context.
// Hyphens are delimiters for nesting, and all keys are lowercased. The assumption is that
// command line flags have already been parsed.
//
// Example:
//      cli.StringFlag{Name: "db-host"},
//
//
//      {
//          "database": {
//              "host": "localhost"
//          }
//      }
func NewSource(opts ...source.Option) source.Source {
	options := source.NewOptions(opts...)

	var ctx *cli.Context

	if c, ok := options.Context.Value(contextKey{}).(*cli.Context); ok {
		ctx = c
	} else {
		// no context
		// get the default app/flags
		app := cmd.App()
		flags := app.Flags

		// create flagset
		set := flag.NewFlagSet(app.Name, flag.ContinueOnError)

		// apply flags to set
		for _, f := range flags {
			f.Apply(set)
		}

		// parse flags
		set.SetOutput(ioutil.Discard)
		set.Parse(os.Args[1:])

		// normalise flags
		normalizeFlags(app.Flags, set)

		// create context
		ctx = cli.NewContext(app, set, nil)
	}

	return &cliSource{
		ctx:  ctx,
		opts: options,
	}
}

// WithContext returns a new source with the context specified.
// The assumption is that Context is retrieved within an app.Action function.
func WithContext(ctx *cli.Context, opts ...source.Option) source.Source {
	return &cliSource{
		ctx:  ctx,
		opts: source.NewOptions(opts...),
	}
}
