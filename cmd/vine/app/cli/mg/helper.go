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
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/xlab/treeprint"

	"github.com/lack-io/vine/cmd/vine/app/cli/util/tool"
)

var defaultFlag = []string{
	"-installsuffix",
	"cgo",
	"-ldflags \"-s -W\"",
}

func protoComments(goDir, name string) []string {
	return []string{
		"\ndownload protoc zip packages (protoc-$VERSION-$PLATFORM.zip) and install:\n",
		"visit https://github.com/protocolbuffers/protobuf/releases",
		"\ndownload protobuf for vine:\n",
		"cd " + goDir,
		"\ninstall dependencies:",
		"\tgo get github.com/gogo/protobuf",
		"\tgo get github.com/lack-io/vine/cmd/protoc-gen-gogo",
		"\tgo get github.com/lack-io/vine/cmd/protoc-gen-vine",
		"\tgo get github.com/lack-io/vine/cmd/protoc-gen-validator",
		"\tgo get github.com/lack-io/vine/cmd/protoc-gen-deepcopy",
		"\tgo get github.com/lack-io/vine/cmd/protoc-gen-dao\n",
		"cd " + goDir,
		"\tvine build " + name,
	}
}

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
		"quota": func(s string) string {
			return strings.ReplaceAll(s, `"`, `\"`)
		},
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
	fmt.Printf("Creating resource %s in %s\n\n", c.Name, c.GoDir)

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
