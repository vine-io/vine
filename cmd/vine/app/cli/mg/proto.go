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
	"path/filepath"
	"strings"

	"github.com/lack-io/cli"
	"github.com/lack-io/vine/cmd/vine/version"

	t2 "github.com/lack-io/vine/cmd/vine/app/cli/mg/template"
	"github.com/lack-io/vine/cmd/vine/app/cli/util/tool"
)

func runProto(ctx *cli.Context) {
	atype := ctx.String("type")
	pv := ctx.String("proto-version")
	group := ctx.String("group")
	name := ctx.Args().First()

	if len(name) == 0 {
		fmt.Println("specify service name")
		return
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		fmt.Println("invalid service name: contains '/' or '\\'")
		return
	}

	if _, err := os.Stat("vine.toml"); err != nil {
		fmt.Println(err)
		return
	}

	cfg, err := tool.New("vine.toml")
	if err != nil {
		fmt.Printf("invalid vine project: %v\n", err)
		return
	}

	dir, _ := os.Getwd()
	goDir := dir
	dir = strings.TrimPrefix(dir, build.Default.GOPATH+"/src/")
	c := config{
		Name:    name,
		Type:    atype,
		Cluster: true,
		Dir:     dir,
		Group:   group,
		GoDir:   goDir,
		Version: pv,
		Toml:    cfg,
	}

	c.GoVersion = version.GoV()
	c.VineVersion = version.GitTag

	for _, item := range c.Toml.Proto {
		if item.Type == atype && item.Name == name && item.Version == pv {
			fmt.Printf("%s %s/%s.proto exists\n", atype, pv, name)
			return
		}
	}

	switch atype {
	case "service":
		c.Files = []file{
			{"proto/service/" + group + "/" + pv + "/" + name + ".proto", t2.ProtoNew},
			{"vine.toml", t2.TOML},
		}
		c.Toml.Proto = append(c.Toml.Proto, tool.Proto{
			Name:    name,
			Pb:      filepath.Join(c.Dir, "proto", "service", group, pv, name+".proto"),
			Group:   group,
			Version: pv,
			Type:    "service",
			Plugins: []string{"vine", "validator"},
		})
	case "api":
		c.Files = []file{
			{"proto/apis/" + group + "/" + pv + "/" + name + ".proto", t2.ProtoType},
			{"vine.toml", t2.TOML},
		}
		c.Toml.Proto = append(c.Toml.Proto, tool.Proto{
			Name:    name,
			Pb:      filepath.Join(c.Dir, "proto", "apis", group, pv, name+".proto"),
			Group:   group,
			Version: pv,
			Type:    "api",
			Plugins: []string{"validator", "dao", "deepcopy"},
		})
	}

	if err := create(c); err != nil {
		fmt.Println(err)
		return
	}
}

func cmdProto() *cli.Command {
	return &cli.Command{
		Name:  "proto",
		Usage: "Create a proto file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "type",
				Usage: "type of proto eg service, api",
				Value: "api",
			},
			&cli.StringFlag{
				Name:    "proto-version",
				Aliases: []string{"v"},
				Usage:   "specify proto version",
				Value:   "v1",
			},
			&cli.StringFlag{
				Name:  "group",
				Usage: "specify the group",
				Value: "core",
			},
		},
		Action: func(c *cli.Context) error {
			runProto(c)
			return nil
		},
	}
}
