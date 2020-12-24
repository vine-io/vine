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

package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/lack-io/vine/service/plugin"
)

func buildSo(soPath string, parts []string) error {
	// check if .so file exists
	if _, err := os.Stat(soPath); os.IsExist(err) {
		return nil
	}

	path := filepath.Join(append([]string{"github.com/vine/plugins"}, parts...)...)
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName}, path)
	if err != nil {
		return err
	}

	// name and things
	name := pkgs[0].Name
	// type of plugin
	typ := parts[0]
	// new func signature
	newfn := fmt.Sprintf("New%s", strings.Title(typ))

	// vine has NewPlugin type def
	if typ == "vine" {
		newfn = "NewPlugin"
	}

	// now build the plugin
	if err := plugin.Build(soPath, &plugin.Config{
		Name:    name,
		Type:    typ,
		Path:    path,
		NewFunc: newfn,
	}); err != nil {
		return fmt.Errorf("Failed to build plugin %s: %v", name, err)
	}

	return nil
}

func load(p string) error {
	p = strings.TrimSpace(p)

	if len(p) == 0 {
		return nil
	}

	parts := strings.Split(p, "/")

	// 1 part means local plugin
	// plugin/foobar
	if len(parts) == 1 {
		return fmt.Errorf("Unknown plugin %s", p)
	}

	// set soPath to specified path
	soPath := p

	// build on the fly if not .so
	if !strings.HasSuffix(p, ".so") {
		// set new so path
		soPath = filepath.Join("plugin", p+".so")

		// build new .so
		if err := buildSo(soPath, parts); err != nil {
			return err
		}
	}

	// load the plugin
	pl, err := plugin.Load(soPath)
	if err != nil {
		return fmt.Errorf("Failed to load plugin %s: %v", soPath, err)
	}

	// Initialise the plugin
	return plugin.Init(pl)
}
