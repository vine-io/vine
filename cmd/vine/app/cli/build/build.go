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
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/cmd/vine/app/cli/tool"
)

func runBuild(c *cli.Context) {
	cfg, err := tool.New("vine.toml")
	if err != nil && os.IsNotExist(err) {
		fmt.Printf("invalid vine project: %v\n", err)
		return
	}

	var exists bool
	name := c.Args().First()
	wireEnable := c.Bool("wire")

	switch cfg.Package.Kind {
	case "single":
		exists = cfg.Pkg.Name == name
	case "cluster":
		for _, mod := range *cfg.Mod {
			if mod.Name == name {
				exists = true
				break
			}
		}
	}
	if !exists {
		fmt.Printf("package %s not exists\n", name)
		return
	}

	if wireEnable {
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			return
		}
		root := filepath.Join(pwd, "pkg")
		err = filepath.Walk(root, func(p string, _ fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			dir := path.Dir(p)
			base := path.Base(p)
			if base == "inject.go" {
				cmd := exec.Command("wire", "gen")
				cmd.Dir = dir
				out, err := cmd.CombinedOutput()
				if err != nil {
					return fmt.Errorf("generate wire code: %v: %v", err, strings.TrimSuffix(string(out), "\n"))
				}
			}

			return nil
		})
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	exec.Command("make", "build-"+name)
}

func Commands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "build",
			Usage: "build vine project",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "wire",
					Usage: "generate wire code before building vine project.",
					Value: true,
				},
			},
			Action: func(c *cli.Context) error {
				runBuild(c)
				return nil
			},
		},
	}
}
