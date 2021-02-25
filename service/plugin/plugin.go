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

// Plugin is a plugin loaded from a file
type Plugin interface {
	// Initialise a plugin with the config
	Init(c *Config) error
	// Load loads a .so plugin at the given path
	Load(path string) (*Config, error)
	// Build a .so plugin with config at the path specified
	Build(path string, c *Config) error
}

// Config is the plugin config
type Config struct {
	// Name of the plugin e.g rabbitmq
	Name string
	// Type of the plugin e.g broker
	Type string
	// Path specifies the import path
	Path string
	// NewFunc creates an instance of the plugin
	NewFunc interface{}
}

var (
	// Default plugin loader
	DefaultPlugin = NewPlugin()
)

// NewPlugin creates a new plugin interface
func NewPlugin() Plugin {
	return &plugin{}
}

func Build(path string, c *Config) error {
	return DefaultPlugin.Build(path, c)
}

func Load(path string) (*Config, error) {
	return DefaultPlugin.Load(path)
}

func Init(c *Config) error {
	return DefaultPlugin.Init(c)
}
