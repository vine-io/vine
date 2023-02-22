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
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vine-io/vine/cmd/vine/version"
	"github.com/vine-io/vine/util/helper"

	t2 "github.com/vine-io/vine/cmd/vine/app/cli/mg/template"
	"github.com/vine-io/vine/cmd/vine/app/cli/util/tool"
)

func runProto(cmd *cobra.Command, args []string) error {

	flags := cmd.PersistentFlags()
	atype, _ := flags.GetString("type")
	pv, _ := flags.GetString("proto-version")
	group, _ := flags.GetString("group")
	name := helper.NewArgs(args).First()

	if len(name) == 0 {
		return fmt.Errorf("specify service name")
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("invalid service name: contains '/' or '\\'")
	}

	if _, err := os.Stat("vine.toml"); err != nil {
		return err
	}

	cfg, err := tool.New("vine.toml")
	if err != nil {
		return fmt.Errorf("invalid vine project: %v\n", err)
	}

	dir, _ := os.Getwd()
	goDir := dir
	if runtime.GOOS == "windows" {
		dir = strings.TrimPrefix(dir, build.Default.GOPATH+"\\src\\")
	} else {
		dir = strings.TrimPrefix(dir, build.Default.GOPATH+"/src/")
	}
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
		if item.Type == atype && item.Group == group && item.Name == name && item.Version == pv {
			return fmt.Errorf("%s %s/%s/%s.proto exists\n", atype, group, pv, name)
		}
	}

	switch atype {
	case "service":
		c.Files = []file{
			{"api/services/" + group + "/" + pv + "/" + name + ".proto", t2.ProtoNew},
			{"vine.toml", t2.TOML},
		}
		c.Toml.Proto = append(c.Toml.Proto, tool.Proto{
			Name:    name,
			Pb:      filepath.Join(c.Dir, "api", "services", group, pv, name+".proto"),
			Group:   group,
			Version: pv,
			Type:    "service",
			Plugins: []string{"vine", "validator"},
		})
	case "api":
		c.Files = []file{
			{"api/types/" + group + "/" + pv + "/" + name + ".proto", t2.ProtoType},
			{"vine.toml", t2.TOML},
		}
		c.Toml.Proto = append(c.Toml.Proto, tool.Proto{
			Name:    name,
			Pb:      filepath.Join(c.Dir, "api", "types", group, pv, name+".proto"),
			Group:   group,
			Version: pv,
			Type:    "api",
			Plugins: []string{"validator", "dao", "deepcopy"},
		})

		registerFile := filepath.Join(c.Dir, "api", "types", group, pv, "register.go")
		_, err := os.Stat(registerFile)
		if os.IsNotExist(err) {
			c.Files = append(c.Files, file{
				Path: "api/types/" + group + "/" + pv + "/" + "register.go",
				Tmpl: t2.Register,
			})
		}
	}

	return create(c)
}

func cmdProto() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proto",
		Short: "Create a proto file",
		RunE:  runProto,
	}
	cmd.PersistentFlags().String("type", "api", "type of proto eg service, api")
	cmd.PersistentFlags().String("proto-version", "v1", "specify proto version")
	cmd.PersistentFlags().String("group", "core", "specify the group")

	return cmd
}
