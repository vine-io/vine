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

func protoComments(goDir, alias string) []string {
	return []string{
		"\ndownload protoc zip packages (protoc-$VERSION-$PLATFORM.zip) and install:\n",
		"visit https://github.com/protocolbuffers/protobuf/releases",
		"\ndownload protobuf for vine:\n",
		"cd " + goDir,
		"make install\n",
		"compile the proto file " + alias + ".proto:\n",
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
	Alias string
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
	FQDN string
	// github.com/vine/foo
	Dir string
	// $GOPATH/src/github.com/vine/foo
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

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	t, err := template.New("f").Funcs(fn).Parse(tmpl)
	if err != nil {
		return err
	}

	return t.Execute(f, c)
}

func create(c config) error {
	// check if dir exists
	if _, err := os.Stat(c.GoDir); !os.IsNotExist(err) {
		return fmt.Errorf("%s already exists", c.GoDir)
	}

	// just wait
	<-time.After(time.Millisecond * 250)

	fmt.Printf("Creating service %s in %s\n\n", c.FQDN, c.GoDir)

	t := treeprint.New()

	// write the files
	for _, file := range c.Files {
		f := filepath.Join(c.GoDir, file.Path)
		dir := filepath.Dir(f)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
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
	alias := ctx.String("alias")
	fqdn := ctx.String("fqdn")
	atype := ctx.String("type")
	cluster := ctx.Bool("cluster")
	dir := ctx.Args().First()
	useGoPath := ctx.Bool("gopath")
	useGoModule := os.Getenv("GO111MODULE")
	var plugins []string

	if len(dir) == 0 {
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

	// set the command
	command := "vine new"
	if len(namespace) > 0 {
		command += " --namespace=" + namespace
	}
	if len(alias) > 0 {
		command += " --alias=" + alias
	}
	if cluster {
		command += " --cluster"
	}
	if len(fqdn) > 0 {
		command += " --fqdn=" + fqdn
	}
	if len(atype) > 0 {
		command += " --type=" + atype
	}
	if plugins := ctx.StringSlice("plugin"); len(plugins) > 0 {
		command += " --plugin=" + strings.Join(plugins, ":")
	}
	command += " " + dir

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

	if len(alias) == 0 {
		// set as last part
		alias = filepath.Base(dir)
		// strip hyphens
		parts := strings.Split(alias, "-")
		alias = parts[0]
	}

	if len(fqdn) == 0 {
		fqdn = strings.Join([]string{namespace, atype, alias}, ".")
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
		Alias:     alias,
		Command:   command,
		Namespace: namespace,
		Type:      atype,
		Cluster:   cluster,
		FQDN:      fqdn,
		Dir:       dir,
		GoDir:     goDir,
		GoPath:    goPath,
		UseGoPath: useGoPath,
		Plugins:   plugins,
		Comments:  protoComments(goDir, alias),
	}

	//if kin && atype != Package {
	//	var err error
	//	f := filepath.Join(goDir, "vine.toml")
	//	c.Toml, err = tool.New(f)
	//	if err != nil {
	//		fmt.Printf("read %s failed: %v", f, err)
	//		return
	//	}
	//}
	if cluster {
		fmt.Println("todo list")
	} else {
		switch atype {
		case Service:
			// create service config
			c.Files = []file{
				{"cmd/main.go", t2.SingleCMD},
				{"pkg/mod.go", t2.SingleMod},
				{"pkg/plugin.go", t2.SinglePlugin},
				{"pkg/app.go", t2.SingleApp},
				{"pkg/default.go", t2.SingleDefault},
				{"pkg/server/" + alias + ".go", t2.SingleSRV},
				{"pkg/service/" + alias + ".go", t2.ServiceSRV},
				{"pkg/dao/" + alias + ".go", t2.DaoHandler},
				{"deploy/Dockerfile", t2.DockerSRV},
				{"deploy/" + alias + ".conf", t2.ConfSRV},
				{"deploy/" + alias + ".service", t2.SystemedSRV},
				{"deploy/Dockerfile", t2.DockerSRV},
				{"proto/apis/apis.proto", t2.ProtoType},
				{"proto/service/" + alias + "/" + alias + ".proto", t2.ProtoSRV},
				{"Makefile", t2.SingleMakefile},
				{"vine.toml", t2.TOML},
				{"README.md", t2.Readme},
				{".gitignore", t2.GitIgnore},
			}
		//case Gateway:
		//	// create api config
		//	c.Files = []file{
		//		{"main.go", t2.MainAPI},
		//		//{"plugin.go", t2.Plugin},
		//		{"client/" + alias + ".go", t2.WrapperAPI},
		//		{"handler/" + alias + ".go", t2.HandlerAPI},
		//		{"proto/" + alias + "/" + alias + ".proto", t2.ProtoAPI},
		//		{"Makefile", t2.SingleMakefile},
		//		{"Dockerfile", t2.DockerSRV},
		//		{"README.md", t2.Readme},
		//		{".gitignore", t2.GitIgnore},
		//	}
		case Web:
			// create service config
			//c.Files = []file{
			//	{"main.go", t2.MainWEB},
			//	//{"plugin.go", t2.Plugin},
			//	{"handler/handler.go", t2.HandlerWEB},
			//	{"html/index.html", t2.HTMLWEB},
			//	{"Dockerfile", t2.DockerWEB},
			//	{"Makefile", t2.SingleMakefile},
			//	{"README.md", t2.Readme},
			//	{".gitignore", t2.GitIgnore},
			//}
			//c.Comments = []string{}
			fmt.Println("todo list")
			return
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
					Name:  "fqdn",
					Usage: "FQDN of service e.g com.example.service.service (defaults to namespace.type.alias)",
				},
				&cli.StringFlag{
					Name:  "alias",
					Usage: "Alias is the short name used as part of combined name if specified",
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
