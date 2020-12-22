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

// Package cli is a command line interface
package cli

import (
	"fmt"
	"os"
	osexec "os/exec"
	"strings"

	"github.com/chzyer/readline"
	"github.com/lack-io/cli"

	"github.com/lack-io/vine/client/cli/util"
	"github.com/lack-io/vine/cmd"

	_ "github.com/lack-io/vine/client/cli/auth"
	_ "github.com/lack-io/vine/client/cli/config"
	_ "github.com/lack-io/vine/client/cli/gen"
	_ "github.com/lack-io/vine/client/cli/init"
	_ "github.com/lack-io/vine/client/cli/network"
	_ "github.com/lack-io/vine/client/cli/new"
	_ "github.com/lack-io/vine/client/cli/run"
	_ "github.com/lack-io/vine/client/cli/signup"
	_ "github.com/lack-io/vine/client/cli/store"
	_ "github.com/lack-io/vine/client/cli/user"
)

var (
	prompt = "vine> "

	// TODO: only run fixed set of commands for security purposes
	commands = map[string]*command{}
)

type command struct {
	name  string
	usage string
	exec  util.Exec
}

func Run(c *cli.Context) error {
	// take the first arg as the binary
	binary := os.Args[0]

	r, err := readline.New(prompt)
	if err != nil {
		return err
	}
	defer r.Close()

	for {
		args, err := r.Readline()
		if err != nil {
			fmt.Fprint(os.Stdout, err)
			return err
		}

		args = strings.TrimSpace(args)

		// skip no args
		if len(args) == 0 {
			continue
		}

		parts := strings.Split(args, " ")
		if len(parts) == 0 {
			continue
		}

		cmd := osexec.Command(binary, parts...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(string(err.(*osexec.ExitError).Stderr))
		}
	}

	return nil
}

func init() {
	cmd.Register(
		&cli.Command{
			Name:   "cli",
			Usage:  "Run the interactive CLI",
			Action: Run,
		},
		&cli.Command{
			Name:   "call",
			Usage:  `Call a service e.g vine call greeter Say.Hello '{"name": "John"}'`,
			Action: util.Print(callService),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "address",
					Usage:   "Set the address of the service instance to call",
					EnvVars: []string{"VINE_ADDRESS"},
				},
				&cli.StringFlag{
					Name:    "output, o",
					Usage:   "Set the output format; json (default), raw",
					EnvVars: []string{"VINE_OUTPUT"},
				},
				&cli.StringSliceFlag{
					Name:    "metadata",
					Usage:   "A list of key-value pairs to be forwarded as metadata",
					EnvVars: []string{"VINE_METADATA"},
				},
				&cli.StringFlag{
					Name:  "request-timeout",
					Usage: "timeout duration",
				},
			},
		},
		&cli.Command{
			Name:   "stream",
			Usage:  `Create a service stream e.g. vine stream foo Bar.Baz '{"key": "value"}'`,
			Action: util.Print(streamService),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "output, o",
					Usage:   "Set the output format; json (default), raw",
					EnvVars: []string{"VINE_OUTPUT"},
				},
				&cli.StringSliceFlag{
					Name:    "metadata",
					Usage:   "A list of key-value pairs to be forwarded as metadata",
					EnvVars: []string{"VINE_METADATA"},
				},
			},
		},
		&cli.Command{
			Name:   "stats",
			Usage:  "Query the stats of a service",
			Action: util.Print(queryStats),
		},
		&cli.Command{
			Name:   "env",
			Usage:  "Get/set vine cli environment",
			Action: util.Print(listEnvs),
			Subcommands: []*cli.Command{
				{
					Name:   "get",
					Usage:  "Get the currently selected environment",
					Action: util.Print(getEnv),
				},
				{
					Name:   "set",
					Usage:  "Set the environment to use for subsequent commands e.g. vine env set dev",
					Action: util.Print(setEnv),
				},
				{
					Name:   "add",
					Usage:  "Add a new environment e.g. vine env add foo 127.0.0.1:8081",
					Action: util.Print(addEnv),
				},
				{
					Name:   "del",
					Usage:  "Delete an environment from your list e.g. vine env del foo",
					Action: util.Print(delEnv),
				},
			},
		},
		&cli.Command{
			Name:   "services",
			Usage:  "List services in the registry",
			Action: util.Print(listServices),
		},
	)
}
