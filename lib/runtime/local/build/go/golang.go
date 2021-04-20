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

// Package golang is a go package manager
package golang

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/lack-io/vine/lib/runtime/local/build"
)

type Builder struct {
	Options build.Options
	Cmd     string
	Path    string
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

func (g *Builder) Build(s *build.Source) (*build.Package, error) {
	binary := filepath.Join(g.Path, s.Repository.Name)
	source := filepath.Join(s.Repository.Path, s.Repository.Name)

	cmd := exec.Command(g.Cmd, "build", "-o", binary, source)
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return &build.Package{
		Name:   s.Repository.Name,
		Path:   binary,
		Type:   "go",
		Source: s,
	}, nil
}

func (g *Builder) Clean(b *build.Package) error {
	binary := filepath.Join(b.Path, b.Name)
	return os.Remove(binary)
}

func NewBuild(opts ...build.Option) build.Builder {
	options := build.Options{
		Path: os.TempDir(),
	}
	for _, o := range opts {
		o(&options)
	}
	return &Builder{
		Options: options,
		Cmd:     whichGo(),
		Path:    options.Path,
	}
}
