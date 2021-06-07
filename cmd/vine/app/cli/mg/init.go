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
	"runtime"
	"strings"

	"github.com/lack-io/cli"
	"github.com/lack-io/vine/cmd/vine/version"

	t2 "github.com/lack-io/vine/cmd/vine/app/cli/mg/template"
	"github.com/lack-io/vine/cmd/vine/app/cli/util/tool"
)

func runInit(ctx *cli.Context) {
	cluster := ctx.Bool("cluster")
	namespace := ctx.String("namespace")
	useGoModule := os.Getenv("GO111MODULE")

	goPath := build.Default.GOPATH
	// attempt to split path if not windows
	if runtime.GOOS == "windows" {
		goPath = strings.Split(goPath, ";")[0]
	} else {
		goPath = strings.Split(goPath, ":")[0]
	}

	dir, _ := os.Getwd()
	goDir := dir
	if runtime.GOOS == "windows" {
		goDir = strings.TrimPrefix(goDir, goPath+"\\src\\")
	} else {
		goDir = strings.TrimPrefix(goDir, goPath+"/src/")
	}
	c := config{
		Namespace: namespace,
		Cluster:   cluster,
		Dir:       goDir,
		GoDir:     dir,
	}

	c.GoVersion = version.GoV()
	c.VineVersion = version.GitTag

	if _, err := os.Stat("vine.toml"); !os.IsNotExist(err) {
		fmt.Println("vine.toml already exists")
		return
	}
	c.Toml = &tool.Config{
		Package: tool.Package{
			Namespace: namespace,
		},
	}
	if cluster {
		c.Toml.Package.Kind = "cluster"
	} else {
		c.Toml.Package.Kind = "single"
	}

	c.Files = []file{
		{"vine.toml", t2.TOML},
		{"README.md", t2.Readme},
		{".gitignore", t2.GitIgnore},
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
