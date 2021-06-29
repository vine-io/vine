// MIT License
//
// Copyright (c) 2021 Lack
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

package mg

import (
	"fmt"
	"go/build"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lack-io/cli"
	"github.com/lack-io/vine/cmd/vine/version"

	t2 "github.com/lack-io/vine/cmd/vine/app/cli/mg/template"
	"github.com/lack-io/vine/cmd/vine/app/cli/util/tool"
)

func runWeb(ctx *cli.Context) {
	cfg, err := tool.New("vine.toml")
	if err != nil {
		fmt.Printf("invalid vine project: %v\n", err)
		return
	}

	var plugins []string
	atype := "web"
	dir, _ := os.Getwd()
	namespace := cfg.Package.Namespace
	cluster := cfg.Package.Kind == "cluster"

	var name string
	if cluster {
		name = ctx.Args().First()

		if len(name) == 0 {
			fmt.Println("specify service name")
			return
		}

		if strings.Contains(name, "/") || strings.Contains(name, "\\") {
			fmt.Println("invalid service name: contains '/' or '\\'")
			return
		}
	} else {
		name = filepath.Base(dir)
	}

	alias := strings.Join([]string{namespace, atype, name}, ".")

	// set the command
	command := "vine new"
	if len(namespace) > 0 {
		command += " --namespace=" + namespace
	}
	if plugins := ctx.StringSlice("plugin"); len(plugins) > 0 {
		command += " --plugin=" + strings.Join(plugins, ":")
	}
	command += " " + name

	goPath := build.Default.GOPATH
	// attempt to split path if not windows
	if runtime.GOOS == "windows" {
		goPath = strings.Split(goPath, ";")[0]
	} else {
		goPath = strings.Split(goPath, ":")[0]
	}

	for _, plugin := range ctx.StringSlice("plugin") {
		// registry=etcd:broker=nats
		for _, p := range strings.Split(plugin, ":") {
			// registry=etcd
			parts := strings.Split(p, "=")
			if len(parts) < 2 {
				continue
			}
			plugins = append(plugins, path.Join(parts...))
		}
	}

	goDir := dir
	dir = strings.TrimPrefix(dir, goPath+"/src/")
	if runtime.GOOS == "windows" {
		dir = strings.TrimPrefix(dir, goPath+"\\src\\")
	} else {
		dir = strings.TrimPrefix(dir, goPath+"/src/")
	}
	c := config{
		Name:      name,
		Command:   command,
		Namespace: namespace,
		Type:      atype,
		Cluster:   cluster,
		Alias:     alias,
		Dir:       dir,
		GoDir:     goDir,
		GoPath:    goPath,
		Plugins:   plugins,
		Comments:  protoComments(dir, name),
		Toml:      cfg,
	}

	c.GoVersion = version.GoV()
	c.VineVersion = version.GitTag

	if cluster {
		if c.Toml.Mod != nil {
			for _, item := range *c.Toml.Mod {
				if item.Name == name {
					fmt.Printf("web %s already exists\n", name)
					return
				}
			}
		} else {
			c.Toml.Mod = &tool.Mods{}
		}
		*c.Toml.Mod = append(*c.Toml.Mod, tool.Mod{
			Name:    c.Name,
			Alias:   c.Alias,
			Type:    atype,
			Version: "latest",
			Main:    filepath.Join("cmd", name, "main.go"),
			Dir:     filepath.Join(c.Dir, c.Name),
			Flags:   defaultFlag,
		})
		// create web config
		c.Files = []file{
			{"cmd/" + name + "/main.go", t2.ClusterCMD},
			{"pkg/runtime/doc.go", t2.Doc},
			{"pkg/" + name + "/plugin.go", t2.ClusterPlugin},
			{"pkg/" + name + "/app.go", t2.ClusterWebSRV},
			{"deploy/docker/" + name + "/Dockerfile", t2.DockerSRV},
			{"deploy/config/" + name + ".ini", t2.ConfSRV},
			{"deploy/systemd/" + name + ".service", t2.SystemedSRV},
			{"vine.toml", t2.TOML},
		}
	} else {
		if c.Toml.Pkg != nil {
			fmt.Printf("web %s already exists", name)
			return
		}
		c.Toml.Pkg = &tool.Mod{
			Name:    c.Name,
			Alias:   c.Alias,
			Type:    atype,
			Version: "latest",
			Main:    filepath.Join("cmd", "main.go"),
			Dir:     filepath.Join(c.Dir, c.Name),
			Flags:   defaultFlag,
		}
		// create web config
		c.Files = []file{
			{"cmd/main.go", t2.SingleCMD},
			{"pkg/runtime/doc.go", t2.Doc},
			{"pkg/plugin.go", t2.SinglePlugin},
			{"pkg/app.go", t2.SingleWebSRV},
			{"deploy/Dockerfile", t2.DockerSRV},
			{"deploy/" + name + ".ini", t2.ConfSRV},
			{"deploy/" + name + ".service", t2.SystemedSRV},
			{"vine.toml", t2.TOML},
			{"README.md", t2.Readme},
			{".gitignore", t2.GitIgnore},
		}
	}

	if err := create(c); err != nil {
		fmt.Println(err)
		return
	}
}

func cmdWeb() *cli.Command {
	return &cli.Command{
		Name:  "web",
		Usage: "Create a web template",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "plugin",
				Usage: "Specify plugins e.g --plugin=registry=etcd:broker=nats or use flag multiple times",
			},
		},
		Action: func(c *cli.Context) error {
			runWeb(c)
			return nil
		},
	}
}
