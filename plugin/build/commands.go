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

package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/plugin"
	log "github.com/lack-io/vine/service/logger"
	goplugin "github.com/lack-io/vine/service/plugin"
)

func build(ctx *cli.Context) error {
	name := ctx.String("name")
	path := ctx.String("path")
	newfn := ctx.String("func")
	typ := ctx.String("type")
	out := ctx.String("output")

	if len(name) == 0 {
		// TODO return err
		fmt.Println("specify --name of plugin")
		os.Exit(1)
	}

	if len(typ) == 0 {
		// TODO return err
		fmt.Println("specify --type of plugin")
		os.Exit(1)
	}

	// set the path
	if len(path) == 0 {
		// github.com/vine/plugins/broker/rabbitmq
		// github.com/vine/plugins/vine/basic_auth
		path = filepath.Join("github.com/vine/plugins", typ, name)
	}

	// set the newfn
	if len(newfn) == 0 {
		if typ == "vine" {
			newfn = "NewPlugin"
		} else {
			newfn = "New" + strings.Title(typ)
		}
	}

	if len(out) == 0 {
		out = "./"
	}

	// create a .so file
	if !strings.HasSuffix(out, ".so") {
		out = filepath.Join(out, name+".so")
	}

	if err := goplugin.Build(out, &goplugin.Config{
		Name:    name,
		Type:    typ,
		Path:    path,
		NewFunc: newfn,
	}); err != nil {
		// TODO return err
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Plugin %s generated at %s\n", name, out)
	return nil
}

func pluginCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:   "build",
			Usage:  "Build a vine plugin",
			Action: build,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "name",
					Usage: "Name of the plugin e.g rabbitmq",
				},
				&cli.StringFlag{
					Name:  "type",
					Usage: "Type of the plugin e.g broker",
				},
				&cli.StringFlag{
					Name:  "path",
					Usage: "Import path of the plugin",
				},
				&cli.StringFlag{
					Name:  "func",
					Usage: "New plugin function creator name e.g NewBroker",
				},
				&cli.StringFlag{
					Name:  "output, o",
					Usage: "Output dir or file for the plugin",
				},
			},
		},
	}
}

// Commands returns license commands
func Commands() []*cli.Command {
	return []*cli.Command{{
		Name:        "plugin",
		Usage:       "Plugin commands",
		Subcommands: pluginCommands(),
	}}
}

// returns a vine plugin which loads plugins
func Flags() plugin.Plugin {
	return plugin.NewPlugin(
		plugin.WithName("plugin"),
		plugin.WithFlag(
			&cli.StringSliceFlag{
				Name:    "plugin",
				EnvVars: []string{"VINE_PLUGIN"},
				Usage:   "Comma separated list of plugins e.g broker/rabbitmq, registry/etcd, vine/basic_auth, /path/to/plugin.so",
			},
		),
		plugin.WithInit(func(ctx *cli.Context) error {
			plugins := ctx.StringSlice("plugin")
			if len(plugins) == 0 {
				return nil
			}

			for _, p := range plugins {
				if err := load(p); err != nil {
					log.Errorf("Error loading plugin %s: %v", p, err)
					return err
				}
				log.Infof("Loaded plugin %s\n", p)
			}

			return nil
		}),
	)
}
