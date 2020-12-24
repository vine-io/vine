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

package config

import (
	"fmt"

	"github.com/lack-io/cli"
)

var (
	// Commands defines a set of commands for local config
	Commands = []*cli.Command{
		{
			Name:   "get",
			Usage:  "Get a value by specifying [key] as an arg",
			Action: get,
		},
		{
			Name:   "set",
			Usage:  "Set a key-val using [key] [value] as args",
			Action: set,
		},
		{
			Name:   "delete",
			Usage:  "Delete a value using [key] as an arg",
			Action: del,
		},
	}
)

func get(ctx *cli.Context) error {
	args := ctx.Args()
	key := args.Get(0)
	val := args.Get(1)

	val, err := Get(key)
	if err != nil {
		return err
	}

	fmt.Println(val)
	return nil
}

func set(ctx *cli.Context) error {
	args := ctx.Args()
	key := args.Get(0)
	val := args.Get(1)

	return Set(key, val)
}

func del(ctx *cli.Context) error {
	args := ctx.Args()
	key := args.Get(0)

	if len(key) == 0 {
		return fmt.Errorf("key cannot be blank")
	}

	// TODO: actually delete the key also
	return Set(key, "")
}
