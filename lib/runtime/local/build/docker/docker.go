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

// Package docker builds docker images
package docker

import (
	"archive/tar"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"

	"github.com/lack-io/vine/lib/logger"
	"github.com/lack-io/vine/lib/runtime/local/build"
)

type Builder struct {
	Options build.Options
	Client  *docker.Client
}

func (d *Builder) Build(s *build.Source) (*build.Package, error) {
	image := filepath.Join(s.Repository.Path, s.Repository.Name)

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	dockerFile := "Dockerfile"

	// open docker file
	f, err := os.Open(filepath.Join(s.Repository.Path, s.Repository.Name, dockerFile))
	if err != nil {
		return nil, err
	}
	// read docker file
	by, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	tarHeader := &tar.Header{
		Name: dockerFile,
		Size: int64(len(by)),
	}
	err = tw.WriteHeader(tarHeader)
	if err != nil {
		return nil, err
	}
	_, err = tw.Write(by)
	if err != nil {
		return nil, err
	}
	tr := bytes.NewReader(buf.Bytes())

	err = d.Client.BuildImage(docker.BuildImageOptions{
		Name:           image,
		Dockerfile:     dockerFile,
		InputStream:    tr,
		OutputStream:   ioutil.Discard,
		RmTmpContainer: true,
		SuppressOutput: true,
	})
	if err != nil {
		return nil, err
	}
	return &build.Package{
		Name:   image,
		Path:   image,
		Type:   "docker",
		Source: s,
	}, nil
}

func (d *Builder) Clean(b *build.Package) error {
	image := filepath.Join(b.Path, b.Name)
	return d.Client.RemoveImage(image)
}

func NewBuilder(opts ...build.Option) build.Builder {
	options := build.Options{}
	for _, o := range opts {
		o(&options)
	}
	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	if err != nil {
		logger.Fatal(err)
	}
	return &Builder{
		Options: options,
		Client:  client,
	}
}
