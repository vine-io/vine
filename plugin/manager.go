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

package plugin

import (
	"fmt"
	"sync"
)

type manager struct {
	sync.Mutex
	plugins    []Plugin
	registered map[string]bool
}

var (
	// global plugin manager
	defaultManager = newManager()
)

func newManager() *manager {
	return &manager{
		registered: make(map[string]bool),
	}
}

func (m *manager) Plugins() []Plugin {
	m.Lock()
	defer m.Unlock()
	return m.plugins
}

func (m *manager) Register(plugin Plugin) error {
	m.Lock()
	defer m.Unlock()

	name := plugin.String()

	if m.registered[name] {
		return fmt.Errorf("Plugin with name %s already registered", name)
	}

	m.registered[name] = true
	m.plugins = append(m.plugins, plugin)
	return nil
}

func (m *manager) isRegistered(plugin Plugin) bool {
	m.Lock()
	defer m.Unlock()

	return m.registered[plugin.String()]
}
