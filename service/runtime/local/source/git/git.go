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

// Package git provides a git source
package git

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"

	"github.com/lack-io/vine/service/runtime/local/source"
)

// Source retrieves source code
// An empty struct can be used
type Source struct {
	Options source.Options
}

func (g *Source) Fetch(url string) (*source.Repository, error) {
	purl := url

	if parts := strings.Split(url, "://"); len(parts) > 1 {
		purl = parts[len(parts)-1]
	}

	name := filepath.Base(url)
	path := filepath.Join(g.Options.Path, purl)

	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL: url,
	})
	if err == nil {
		return &source.Repository{
			Name: name,
			Path: path,
			URL:  url,
		}, nil
	}

	// repo already exists
	if err != git.ErrRepositoryAlreadyExists {
		return nil, err
	}

	// open repo
	re, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	// update it
	if err := re.Fetch(nil); err != nil {
		return nil, err
	}

	return &source.Repository{
		Name: name,
		Path: path,
		URL:  url,
	}, nil
}

func (g *Source) Commit(r *source.Repository) error {
	repo := filepath.Join(r.Path)
	re, err := git.PlainOpen(repo)
	if err != nil {
		return err
	}
	return re.Push(nil)
}

func (g *Source) String() string {
	return "git"
}

func NewSource(opts ...source.Option) *Source {
	options := source.Options{
		Path: os.TempDir(),
	}
	for _, o := range opts {
		o(&options)
	}

	return &Source{
		Options: options,
	}
}
