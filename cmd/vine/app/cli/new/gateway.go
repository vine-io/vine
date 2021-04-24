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

package new

import (
	"fmt"
	"go/build"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lack-io/cli"
	t2 "github.com/lack-io/vine/cmd/vine/app/cli/new/template"
	"github.com/lack-io/vine/cmd/vine/app/cli/tool"
)

func runGateway(ctx *cli.Context) {
	namespace := ctx.String("namespace")
	dir := ctx.String("dir")
	name := ctx.Args().First()
	useGoPath := ctx.Bool("gopath")
	useGoModule := os.Getenv("GO111MODULE")
	var plugins []string

	if len(name) == 0 {
		fmt.Println("specify service name")
		return
	}

	if len(namespace) == 0 {
		fmt.Println("namespace not defined")
		return
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		fmt.Println("invalid service name: contains '/' or '\\'")
		return
	}

	atype := "gateway"
	alias := strings.Join([]string{namespace, atype, name}, ".")

	// set the command
	command := "vine new"
	if len(namespace) > 0 {
		command += " --namespace=" + namespace
	}
	if len(dir) > 0 {
		command += " --dir=" + dir
	}
	if plugins := ctx.StringSlice("plugin"); len(plugins) > 0 {
		command += " --plugin=" + strings.Join(plugins, ":")
	}
	command += " " + name

	// check if the path is absolute, we don't want this
	// we want to a relative path so we can install in GOPATH
	if path.IsAbs(dir) {
		fmt.Println("require relative path as service will be installed in GOPATH")
		return
	}

	var goPath string
	var goDir string
	// only set gopath if told to use it
	if useGoPath {
		goPath = build.Default.GOPATH

		// don't know GOPATH, runaway....
		if len(goPath) == 0 {
			fmt.Println("unknown GOPATH")
			return
		}

		// attempt to split path if not windows
		if runtime.GOOS == "windows" {
			goPath = strings.Split(goPath, ";")[0]
		} else {
			goPath = strings.Split(goPath, ":")[0]
		}
		goDir = filepath.Join(goPath, "src", path.Clean(dir))
	} else {
		goDir = path.Clean(dir)
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

	c := config{
		Name:      name,
		Command:   command,
		Namespace: namespace,
		Type:      atype,
		Cluster:   true,
		Alias:     alias,
		Dir:       dir,
		GoPath:    goPath,
		GoDir:     goDir,
		UseGoPath: useGoPath,
		Plugins:   plugins,
		Comments:  protoComments(dir, name),
	}

	var exists bool
	if _, err := os.Stat(c.Dir); !os.IsNotExist(err) {
		var err error
		c.Toml, err = tool.New(filepath.Join(c.GoDir, "vine.toml"))
		if err != nil {
			fmt.Printf("invalid vine project: %v\n", err)
			return
		}
		*c.Toml.Mod = append(*c.Toml.Mod, tool.Mod{
			Name:    c.Name,
			Alias:   c.Alias,
			Type:    atype,
			Version: "latest",
			Dir:     filepath.Join(c.Dir, c.Name),
		})
		exists = true
	} else {
		c.Toml = &tool.Config{
			Package: tool.Package{
				Kind:      "cluster",
				Namespace: c.Namespace,
			},
			Mod: &tool.Mods{tool.Mod{
				Name:    c.Name,
				Alias:   c.Alias,
				Type:    atype,
				Version: "latest",
				Dir:     filepath.Join(c.Dir, c.Name),
			}},
		}
		exists = false
	}
	// create gateway config
	c.Files = []file{
		{"cmd/" + name + "/main.go", t2.ClusterCMD},
		{"pkg/" + name + "/mod.go", t2.ClusterMod},
		{"pkg/" + name + "/plugin.go", t2.ClusterPlugin},
		{"pkg/" + name + "/app.go", t2.GatewayApp},
		{"pkg/" + name + "/default.go", t2.ClusterDefault},
		{"deploy/docker/" + name + "/Dockerfile", t2.DockerSRV},
		{"deploy/config/" + name + ".ini", t2.ConfSRV},
		{"deploy/systemed/" + name + ".service", t2.SystemedSRV},
		{"vine.toml", t2.TOML},
	}

	if !exists {
		c.Files = append(c.Files, file{"README.md", t2.Readme})
		c.Files = append(c.Files, file{".gitignore", t2.GitIgnore})
	}

	// set gomodule
	if useGoModule != "off" {
		c.Files = append(c.Files, file{"go.mod", t2.Module})
	}

	if err := create(c); err != nil {
		fmt.Println(err)
		return
	}

}

func CmdGateway() *cli.Command {
	return &cli.Command{
		Name:  "gateway",
		Usage: "Create a gateway template",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "namespace",
				Usage: "Namespace for the project e.g com.example",
				Value: "go.vine",
			},
			&cli.StringFlag{
				Name:  "dir",
				Usage: "base dir of service e.g github.com/lack-io",
			},
			&cli.StringSliceFlag{
				Name:  "plugin",
				Usage: "Specify plugins e.g --plugin=registry=etcd:broker=nats or use flag multiple times",
			},
			&cli.BoolFlag{
				Name:  "gopath",
				Usage: "Create the service in the gopath.",
				Value: true,
			},
		},
		Action: func(c *cli.Context) error {
			runGateway(c)
			return nil
		},
	}
}
