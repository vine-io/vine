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

// Package runtime is the vine runtime
package runtime

import (
	"github.com/lack-io/cli"

	"github.com/lack-io/vine/cmd"
)

// flags is shared flags so we don't have to continually re-add
var flags = []cli.Flag{
	&cli.StringFlag{
		Name:  "name",
		Usage: "Set the name of the service. Otherwise defaults to directory name",
	},
	&cli.StringFlag{
		Name:  "source",
		Usage: "Set the source url of the service e.g github.com/vine/services",
	},
	&cli.StringFlag{
		Name:  "image",
		Usage: "Set the image to use for the container",
	},
	&cli.StringFlag{
		Name:  "command",
		Usage: "Command to exec",
	},
	&cli.StringFlag{
		Name:  "args",
		Usage: "Command args",
	},
	&cli.StringFlag{
		Name:  "type",
		Usage: "The type of service operate on",
	},
	&cli.StringSliceFlag{
		Name:  "env-vars",
		Usage: "Set the environment variables e.g. foo=bar",
	},
}

func init() {
	cmd.Register(
		&cli.Command{
			// In future we'll also have `vine run [x]` hence `vine run service` requiring "service"
			Name:  "run",
			Usage: RunUsage,
			Description: `Examples:
			vine run github.com/lack-io/services/helloworld
			vine run .  # deploy local folder to your local vine server
			vine run ../path/to/folder # deploy local folder to your local vine server
			vine run helloworld # deploy latest version, translates to vine run github.com/vine/services/helloworld
			vine run helloworld@9342934e6180 # deploy certain version
			vine run helloworld@branchname	# deploy certain branch`,
			Flags:  flags,
			Action: runService,
		},
		&cli.Command{
			Name:  "update",
			Usage: UpdateUsage,
			Description: `Examples:
			vine update github.com/lack-io/services/helloworld
			vine update .  # deploy local folder to your local vine server
			vine update ../path/to/folder # deploy local folder to your local vine server
			vine update helloworld # deploy master branch, translates to vine update github.com/vine/services/helloworld
			vine update helloworld@branchname	# deploy certain branch`,
			Flags:  flags,
			Action: updateService,
		},
		&cli.Command{
			Name:  "kill",
			Usage: KillUsage,
			Flags: flags,
			Description: `Examples:
			vine kill github.com/lack-io/services/helloworld
			vine kill .  # kill service deployed from local folder
			vine kill ../path/to/folder # kill service deployed from local folder
			vine kill helloworld # kill serviced deployed from master branch, translates to vine kill github.com/vine/services/helloworld
			vine kill helloworld@branchname	# kill service deployed from certain branch`,
			Action: killService,
		},
		&cli.Command{
			Name:   "status",
			Usage:  GetUsage,
			Flags:  flags,
			Action: getService,
		},
		&cli.Command{
			Name:   "logs",
			Usage:  "Get logs for a service e.g. vine logs helloworld",
			Action: getLogs,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "version",
					Usage: "Set the version of the service to debug",
				},
				&cli.StringFlag{
					Name:    "output",
					Aliases: []string{"o"},
					Usage:   "Set the output format e.g json, text",
				},
				&cli.BoolFlag{
					Name:    "follow",
					Aliases: []string{"f"},
					Usage:   "Set to stream logs continuously (default: true)",
				},
				&cli.StringFlag{
					Name:  "since",
					Usage: "Set to the relative time from which to show the logs for e.g. 1h",
				},
				&cli.IntFlag{
					Name:    "lines",
					Aliases: []string{"n"},
					Usage:   "Set to query the last number of log events",
				},
			},
		},
	)
}
