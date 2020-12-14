// Copyright 2020 The vine Authors
//
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

// Package docker builds docker images
package docker

import (
	"archive/tar"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"

	"github.com/lack-io/vine/log"
	"github.com/lack-io/vine/runtime/local/build"
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
		log.Fatal(err)
	}
	return &Builder{
		Options: options,
		Client:  client,
	}
}
