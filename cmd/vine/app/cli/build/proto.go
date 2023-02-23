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

package build

import (
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vine-io/vine/cmd/vine/app/cli/util/tool"
	"github.com/vine-io/vine/util/helper"
)

func runProto(c *cobra.Command, args []string) error {
	cfg, err := tool.New("vine.toml")
	if err != nil {
		return fmt.Errorf("invalid vine project: %v\n", err)
	}

	flags := c.PersistentFlags()
	atype, _ := flags.GetString("type")
	group, _ := flags.GetString("group")
	version, _ := flags.GetString("proto-version")
	dir, _ := flags.GetString("dir")
	paths, _ := flags.GetStringSlice("path")
	plugins, _ := flags.GetStringSlice("plugins")
	name := helper.NewArgs(args).First()
	goPath := build.Default.GOPATH
	// attempt to split path if not windows
	if runtime.GOOS == "windows" {
		goPath = strings.Split(goPath, ";")[0]
	} else {
		goPath = strings.Split(goPath, ":")[0]
	}

	if dir == "" {
		dir = filepath.Join(goPath, "src")
	}

	gen := func(pb *tool.Proto, root string, paths, plugins []string) error {
		fmt.Printf("change directory %s: \n", root)

		paths = append([]string{root}, paths...)
		args := make([]string, 0)
		for _, p := range paths {
			args = append(args, "-I="+p)
		}
		if len(plugins) == 0 {
			for _, p := range pb.Plugins {
				args = append(args, "--"+p+"_out=:.")
			}
		} else {
			for _, p := range plugins {
				args = append(args, "--"+p+"_out=:.")
			}
		}
		if runtime.GOOS == "windows" {
			args = append(args, filepath.Join(root, pb.Pb))
		} else {
			args = append(args, pb.Pb)
		}

		fmt.Printf("protoc %s\n", strings.Join(args, " "))
		cmd := exec.Command("protoc", args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("generate protobuf: %v: %v\n", err, string(out))
		}

		return nil
	}

	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get pwd: %v\n", err)
	}
	paths = append(paths, filepath.Join(pwd, "vendor"))
	if name != "" {
		var pb *tool.Proto
		for _, p := range cfg.Proto {
			if p.Name == name && p.Type == atype && p.Group == group && p.Version == version {
				pb = &p
				break
			}
		}

		if pb == nil {
			return fmt.Errorf("file %s/%s/%s.proto not found\n", group, version, name)
		}

		err = gen(pb, dir, paths, plugins)
		if err != nil {
			return err
		}
	} else {
		for _, p := range cfg.Proto {
			if err = gen(&p, dir, paths, plugins); err != nil {
				return err
			}
		}
	}

	return nil
}

func cmdProto() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "proto",
		Short:         "Generate protobuf file",
		SilenceErrors: true,
		RunE:          runProto,
	}
	cmd.PersistentFlags().String("type", "api", "the type of protobuf file eg api, service.")
	cmd.PersistentFlags().String("proto-version", "v1", "the version of protobuf file.")
	cmd.PersistentFlags().String("group", "core", "specify the group.")
	cmd.PersistentFlags().String("dir", "", "base directory for protoc command eg $GOPATH/src.")
	cmd.PersistentFlags().StringP("path", "I", "", "specify the directory in which to search for imports.")
	cmd.PersistentFlags().StringP("plugins", "P", "", "specify the gRPC plugin in which to generate.")

	return cmd
}
