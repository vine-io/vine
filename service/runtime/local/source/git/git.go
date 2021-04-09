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
