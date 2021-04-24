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
	"path/filepath"
	"strings"

	"github.com/lack-io/cli"
	t2 "github.com/lack-io/vine/cmd/vine/app/cli/new/template"
	"github.com/lack-io/vine/cmd/vine/app/cli/tool"
)

func runProto(ctx *cli.Context) {
	atype := ctx.String("type")
	name := ctx.Args().First()

	if len(name) == 0 {
		fmt.Println("specify service name")
		return
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		fmt.Println("invalid service name: contains '/' or '\\'")
		return
	}

	dir, _ := os.Getwd()
	c := config{
		Name:    name,
		Type:    atype,
		Cluster: true,
		Dir:     strings.TrimPrefix(dir, build.Default.GOPATH+"/src/"),
		GoDir:   dir,
	}

	if _, err := os.Stat("vine.toml"); err != nil {
		fmt.Println(err)
		return
	}

	var err error
	c.Toml, err = tool.New("vine.toml")
	if err != nil {
		fmt.Printf("invalid vine project: %v\n", err)
		return
	}

	for _, item := range c.Toml.Proto {
		if item.Type == atype && item.Name == name {
			fmt.Printf("%s %s.proto exists\n", atype, name)
			return
		}
	}

	switch atype {
	case "service":
		c.Files = []file{
			{"proto/service/" + name + "/" + name + ".proto", t2.ProtoNew},
			{"vine.toml", t2.TOML},
		}
		c.Toml.Proto = append(c.Toml.Proto, tool.Proto{
			Name:    name,
			Pb:      filepath.Join(c.Dir, "proto", "service", name, name+".proto"),
			Type:    "service",
			Plugins: []string{"vine", "validator"},
		})
	case "api":
		c.Files = []file{
			{"proto/apis/" + name + ".proto", t2.ProtoType},
			{"vine.toml", t2.TOML},
		}
		c.Toml.Proto = append(c.Toml.Proto, tool.Proto{
			Name:    name,
			Pb:      filepath.Join(c.Dir, "proto", "apis", name+".proto"),
			Type:    "api",
			Plugins: []string{"validator", "dao", "deepcopy"},
		})
	}

	if err := create(c); err != nil {
		fmt.Println(err)
		return
	}

}

func CmdProto() *cli.Command {
	return &cli.Command{
		Name:  "proto",
		Usage: "Create a proto file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "type",
				Usage: "type of proto eg service, api",
				Value: "api",
			},
		},
		Action: func(c *cli.Context) error {
			runProto(c)
			return nil
		},
	}
}
