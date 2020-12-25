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

// Package plugin provides the ability to load plugins
package plugin

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	pg "plugin"
	"strings"
	"text/template"

	"github.com/lack-io/vine/service/broker"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/client/selector"
	"github.com/lack-io/vine/service/config/cmd"
	"github.com/lack-io/vine/service/network/transport"
	"github.com/lack-io/vine/service/registry"
	"github.com/lack-io/vine/service/server"
)

type plugin struct{}

// Init sets up the plugin
func (p *plugin) Init(c *Config) error {
	switch c.Type {
	case "broker":
		pg, ok := c.NewFunc.(func(...broker.Option) broker.Broker)
		if !ok {
			return fmt.Errorf("invalid plugin %s", c.Name)
		}
		cmd.DefaultBrokers[c.Name] = pg
	case "client":
		pg, ok := c.NewFunc.(func(...client.Option) client.Client)
		if !ok {
			return fmt.Errorf("invalid plugin %s", c.Name)
		}
		cmd.DefaultClients[c.Name] = pg
	case "registry":
		pg, ok := c.NewFunc.(func(...registry.Option) registry.Registry)
		if !ok {
			return fmt.Errorf("invalid plugin %s", c.Name)
		}
		cmd.DefaultRegistries[c.Name] = pg

	case "selector":
		pg, ok := c.NewFunc.(func(...selector.Option) selector.Selector)
		if !ok {
			return fmt.Errorf("invalid plugin %s", c.Name)
		}
		cmd.DefaultSelectors[c.Name] = pg
	case "server":
		pg, ok := c.NewFunc.(func(...server.Option) server.Server)
		if !ok {
			return fmt.Errorf("invalid plugin %s", c.Name)
		}
		cmd.DefaultServers[c.Name] = pg
	case "transport":
		pg, ok := c.NewFunc.(func(...transport.Option) transport.Transport)
		if !ok {
			return fmt.Errorf("invalid plugin %s", c.Name)
		}
		cmd.DefaultTransports[c.Name] = pg
	default:
		return fmt.Errorf("unknown plugin type: %s for %s", c.Type, c.Name)
	}

	return nil
}

// Load loads a plugin created with `go build -buildmode=plugin`
func (p *plugin) Load(path string) (*Config, error) {
	plugin, err := pg.Open(path)
	if err != nil {
		return nil, err
	}
	s, err := plugin.Lookup("Plugin")
	if err != nil {
		return nil, err
	}
	pl, ok := s.(*Config)
	if !ok {
		return nil, errors.New("could not cast Plugin object")
	}
	return pl, nil
}

// Generate creates a go file at the specified path.
// You must use `go build -buildmode=plugin`to build it.
func (p *plugin) Generate(path string, c *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	t, err := template.New(c.Name).Parse(tmpl)
	if err != nil {
		return err
	}
	return t.Execute(f, c)
}

// Build generates a dso plugin using the go command `go build -buildmode=plugin`
func (p *plugin) Build(path string, c *Config) error {
	path = strings.TrimSuffix(path, ".so")

	// create go file in tmp path
	temp := os.TempDir()
	base := filepath.Base(path)
	goFile := filepath.Join(temp, base+".go")

	// generate .go file
	if err := p.Generate(goFile, c); err != nil {
		return err
	}
	// remove .go file
	defer os.Remove(goFile)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create dir %s: %v", filepath.Dir(path), err)
	}
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", path+".so", goFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
