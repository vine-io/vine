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
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lack-io/cli"
	"github.com/lack-io/vine/cmd/vine/app/cli/util/tool"
)

func runProto(ctx *cli.Context) {
	cfg, err := tool.New("vine.toml")
	if err != nil {
		fmt.Printf("invalid vine project: %v\n", err)
		return
	}

	atype := ctx.String("type")
	group := ctx.String("group")
	version := ctx.String("proto-version")
	dir := ctx.String("dir")
	paths := ctx.StringSlice("path")
	plugins := ctx.StringSlice("plugins")
	name := ctx.Args().First()
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

	gen := func(pb *tool.Proto, root string, paths, plugins []string) {
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
		args = append(args, pb.Pb)

		fmt.Printf("protoc %s\n", strings.Join(args, " "))
		cmd := exec.Command("protoc", args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("generate protobuf: %v: %v\n", err, string(out))
			return
		}
	}

	if name != "" {
		var pb *tool.Proto
		for _, p := range cfg.Proto {
			if p.Name == name && p.Type == atype && p.Group == group && p.Version == version {
				pb = &p
				break
			}
		}

		if pb == nil {
			fmt.Printf("file %s/%s/%s.proto not found\n", group, version, name)
			return
		}

		gen(pb, dir, paths, plugins)
	} else {
		for _, p := range cfg.Proto {
			gen(&p, dir, paths, plugins)
		}
	}
}

func cmdProto() *cli.Command {
	return &cli.Command{
		Name:  "proto",
		Usage: "Generate protobuf file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "type",
				Usage: "the type of protobuf file eg api, service.",
				Value: "api",
			},
			&cli.StringFlag{
				Name:    "proto-version",
				Aliases: []string{"v"},
				Usage:   "the version of protobuf file.",
				Value:   "v1",
			},
			&cli.StringFlag{
				Name:  "group",
				Usage: "specify the group",
				Value: "core",
			},
			&cli.StringFlag{
				Name:  "dir",
				Usage: "base directory for protoc command eg $GOPATH/src.",
			},
			&cli.StringSliceFlag{
				Name:    "path",
				Aliases: []string{"I"},
				Usage:   "specify the directory in which to search for imports.",
			},
			&cli.StringSliceFlag{
				Name:    "plugins",
				Aliases: []string{"P"},
				Usage:   "specify the gRPC plugin in which to generate.",
			},
		},
		Action: func(c *cli.Context) error {
			runProto(c)
			return nil
		},
	}
}
