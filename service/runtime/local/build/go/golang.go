// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package golang is a go package manager
package golang

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/lack-io/vine/service/runtime/local/build"
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
