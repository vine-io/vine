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

// Package new generates vine service templates
package new

import (
	"fmt"
	"go/build"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/lack-io/cli"
	"github.com/xlab/treeprint"

	t2 "github.com/lack-io/vine/cmd/vine/app/cli/new/template"
	"github.com/lack-io/vine/cmd/vine/app/cli/tool"
)

func protoComments(goDir, name string) []string {
	return []string{
		"\ndownload protoc zip packages (protoc-$VERSION-$PLATFORM.zip) and install:\n",
		"visit https://github.com/protocolbuffers/protobuf/releases",
		"\ndownload protobuf for vine:\n",
		"cd " + goDir,
		"make install\n",
		"compile the proto file apis.proto and " + name + ".proto:\n",
		"cd " + goDir,
		"make proto\n",
	}
}

// specify type
const (
	Gateway string = "gateway"
	Service string = "service"
	Web     string = "web"
)

type config struct {
	// foo
	Name string
	// vine new example -type
	Command string
	// go.vine
	Namespace string
	// gateway, service, web
	// gateway only in cluster
	Type string
	// cluster, single
	Cluster bool
	// go.vine.service.foo
	Alias string
	// $GOPATH/src/github.com/vine/foo
	Dir string
	// github.com/vine/foo
	GoDir string
	// $GOPATH
	GoPath string
	// UseGoPath
	UseGoPath bool
	// Files
	Files []file
	// Comments
	Comments []string
	// Plugins registry=etcd:broker=nats
	Plugins []string

	Toml *tool.Config
}

type file struct {
	Path string
	Tmpl string
}

func write(c config, file, tmpl string) error {
	fn := template.FuncMap{
		"title": strings.Title,
	}

	var f *os.File
	var err error
	stat, _ := os.Stat(file)
	if stat == nil {
		f, err = os.Create(file)
		if err != nil {
			return err
		}
	} else {
		f, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
	}
	defer f.Close()

	t, err := template.New("f").Funcs(fn).Parse(tmpl)
	if err != nil {
		return err
	}

	return t.Execute(f, c)
}

func create(c config) error {
	fmt.Printf("Creating service %s in %s\n\n", c.Alias, c.GoDir)

	t := treeprint.New()

	// write the files
	for _, file := range c.Files {
		f := filepath.Join(c.GoDir, file.Path)
		dir := filepath.Dir(f)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			_ = os.MkdirAll(dir, 0755)
		}

		addFileToTree(t, file.Path)
		if err := write(c, f, file.Tmpl); err != nil {
			return err
		}
	}

	// print tree
	fmt.Println(t.String())

	for _, comment := range c.Comments {
		fmt.Println(comment)
	}

	// just wait
	<-time.After(time.Millisecond * 250)

	return nil
}

func addFileToTree(root treeprint.Tree, file string) {

	split := strings.Split(file, "/")
	curr := root
	for i := 0; i < len(split)-1; i++ {
		n := curr.FindByValue(split[i])
		if n != nil {
			curr = n
		} else {
			curr = curr.AddBranch(split[i])
		}
	}
	if curr.FindByValue(split[len(split)-1]) == nil {
		curr.AddNode(split[len(split)-1])
	}
}

func Run(ctx *cli.Context) {
	namespace := ctx.String("namespace")
	dir := ctx.String("dir")
	atype := ctx.String("type")
	cluster := ctx.Bool("cluster")
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

	if len(atype) == 0 {
		fmt.Println("type not defined")
		return
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		fmt.Println("invalid service name: contains '/' or '\\'")
		return
	}

	alias := strings.Join([]string{namespace, atype, name}, ".")

	// set the command
	command := "vine new"
	if len(namespace) > 0 {
		command += " --namespace=" + namespace
	}
	if len(dir) > 0 {
		command += " --dir=" + dir
	}
	if cluster {
		command += " --cluster"
	}
	if len(atype) > 0 {
		command += " --type=" + atype
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
		Cluster:   cluster,
		Alias:     alias,
		Dir:       dir,
		GoPath:    goPath,
		GoDir:     goDir,
		UseGoPath: useGoPath,
		Plugins:   plugins,
		Comments:  protoComments(dir, name),
	}

	if cluster {
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
		switch atype {
		case Service:
			// create service config
			c.Files = []file{
				{"cmd/" + name + "/main.go", t2.ClusterCMD},
				{"pkg/" + name + "/mod.go", t2.ClusterMod},
				{"pkg/" + name + "/plugin.go", t2.ClusterPlugin},
				{"pkg/" + name + "/app.go", t2.ClusterApp},
				{"pkg/" + name + "/default.go", t2.ClusterDefault},
				{"pkg/" + name + "/server/" + name + ".go", t2.ClusterSRV},
				{"pkg/" + name + "/service/" + name + ".go", t2.ServiceSRV},
				{"pkg/" + name + "/dao/" + name + ".go", t2.DaoHandler},
				{"deploy/docker/" + name + "/Dockerfile", t2.DockerSRV},
				{"deploy/config/" + name + ".ini", t2.ConfSRV},
				{"deploy/systemed/" + name + ".service", t2.SystemedSRV},
				{"proto/service/" + name + "/" + name + ".proto", t2.ProtoSRV},
				{"vine.toml", t2.TOML},
			}
		case Gateway:
			// create api config
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
		case Web:
			// create service config
			c.Files = []file{
				{"cmd/" + name + "/main.go", t2.ClusterCMD},
				{"pkg/" + name + "/mod.go", t2.ClusterMod},
				{"pkg/" + name + "/plugin.go", t2.ClusterPlugin},
				{"pkg/" + name + "/app.go", t2.WebSRV},
				{"pkg/" + name + "/default.go", t2.ClusterDefault},
				{"deploy/docker/" + name + "/Dockerfile", t2.DockerSRV},
				{"deploy/config/" + name + ".ini", t2.ConfSRV},
				{"deploy/systemed/" + name + ".service", t2.SystemedSRV},
				{"vine.toml", t2.TOML},
			}
			c.Comments = []string{}
		default:
			fmt.Printf("Unknown type %s, eg service, web\n", atype)
			return
		}

		if !exists {
			if c.Type != Gateway {
				c.Files = append(c.Files, file{"proto/apis/apis.proto", t2.ProtoType})
			}
			c.Files = append(c.Files, file{"Makefile", t2.Makefile})
			c.Files = append(c.Files, file{"README.md", t2.Readme})
			c.Files = append(c.Files, file{".gitignore", t2.GitIgnore})
		}

		// set gomodule
		if useGoModule != "off" {
			c.Files = append(c.Files, file{"go.mod", t2.Module})
		}
	} else {
		c.GoDir = filepath.Join(c.GoDir, c.Name)
		if _, err := os.Stat(c.GoDir); !os.IsNotExist(err) {
			fmt.Printf("%s already exists\n", c.GoDir)
			return
		}
		c.Toml = &tool.Config{
			Package: tool.Package{
				Kind:      "single",
				Namespace: c.Namespace,
			},
			Pkg: &tool.Mod{
				Name:    c.Name,
				Alias:   c.Alias,
				Type:    atype,
				Version: "latest",
				Dir:     filepath.Join(c.Dir, c.Name),
			},
		}
		c.Dir = filepath.Join(c.Dir, c.Name)
		switch atype {
		case Service:
			// create service config
			c.Files = []file{
				{"cmd/main.go", t2.SingleCMD},
				{"pkg/mod.go", t2.SingleMod},
				{"pkg/plugin.go", t2.SinglePlugin},
				{"pkg/app.go", t2.SingleApp},
				{"pkg/default.go", t2.SingleDefault},
				{"pkg/server/" + name + ".go", t2.SingleSRV},
				{"pkg/service/" + name + ".go", t2.ServiceSRV},
				{"pkg/dao/" + name + ".go", t2.DaoHandler},
				{"deploy/Dockerfile", t2.DockerSRV},
				{"deploy/" + name + ".ini", t2.ConfSRV},
				{"deploy/" + name + ".service", t2.SystemedSRV},
				{"proto/apis/apis.proto", t2.ProtoType},
				{"proto/service/" + name + "/" + name + ".proto", t2.ProtoSRV},
				{"Makefile", t2.Makefile},
				{"vine.toml", t2.TOML},
				{"README.md", t2.Readme},
				{".gitignore", t2.GitIgnore},
			}
		case Web:
			// create service config
			c.Files = []file{
				{"cmd/main.go", t2.SingleCMD},
				{"pkg/mod.go", t2.SingleMod},
				{"pkg/plugin.go", t2.SinglePlugin},
				{"pkg/app.go", t2.WebSRV},
				{"pkg/default.go", t2.SingleDefault},
				{"deploy/Dockerfile", t2.DockerSRV},
				{"deploy/" + name + ".ini", t2.ConfSRV},
				{"deploy/" + name + ".service", t2.SystemedSRV},
				{"Makefile", t2.Makefile},
				{"vine.toml", t2.TOML},
				{"README.md", t2.Readme},
				{".gitignore", t2.GitIgnore},
			}
			c.Comments = []string{}
		default:
			fmt.Printf("Unknown type %s, eg service, web\n", atype)
			return
		}

		// set gomodule
		if useGoModule != "off" {
			c.Files = append(c.Files, file{"go.mod", t2.Module})
		}
	}

	if err := create(c); err != nil {
		fmt.Println(err)
		return
	}

}

func Commands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "new",
			Usage: "Create a project template",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "namespace",
					Usage: "Namespace for the project e.g com.example",
					Value: "go.vine",
				},
				&cli.StringFlag{
					Name:  "type",
					Usage: "Type of service e.g gateway, service, web",
					Value: "service",
				},
				&cli.BoolFlag{
					Name:  "cluster",
					Usage: "create cluster package.",
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
				Run(c)
				return nil
			},
		},
	}
}
