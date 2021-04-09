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

// Package golang is a source for Go
package golang

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lack-io/vine/service/runtime/local/source"
)

type Source struct {
	Options source.Options
	// Go Command
	Cmd  string
	Path string
}

func (g *Source) Fetch(url string) (*source.Repository, error) {
	purl := url

	if parts := strings.Split(url, "://"); len(parts) > 1 {
		purl = parts[len(parts)-1]
	}

	// name of repo
	name := filepath.Base(url)
	// local path of repo
	path := filepath.Join(g.Path, purl)
	args := []string{"get", "-d", url, path}

	cmd := exec.Command(g.Cmd, args...)
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return &source.Repository{
		Name: name,
		Path: path,
		URL:  url,
	}, nil
}

// Commit is not yet supported
func (g *Source) Commit(r *source.Repository) error {
	return nil
}

func (g *Source) String() string {
	return "golang"
}

// whichGo locates the go command
func whichGo() string {
	// check GOROOT
	if gr := os.Getenv("GOROOT"); len(gr) > 0 {
		return filepath.Join(gr, "bin", "go")
	}

	// check path
	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		bin := filepath.Join(p, "go")
		if _, err := os.Stat(bin); err == nil {
			return bin
		}
	}

	// best effort
	return "go"
}

func NewSource(opts ...source.Option) source.Source {
	options := source.Options{
		Path: os.TempDir(),
	}
	for _, o := range opts {
		o(&options)
	}

	cmd := whichGo()
	path := options.Path

	// point of no return
	if len(cmd) == 0 {
		panic("Could not find Go executable")
	}

	return &Source{
		Options: options,
		Cmd:     cmd,
		Path:    path,
	}
}
