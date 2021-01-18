// Copyright 2020 lack
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

// Package command is an interface for defining bot commands
package command

var (
	// Commmands keyed by golang/regexp patterns
	// regexp.Match(key, input) is used to match
	Commands = map[string]Command{}
)

// Command is the interface for specific named
// commands executed via plugins or the bot.
type Command interface {
	// Executes the command with args passed in
	Exec(args ...string) ([]byte, error)
	// Usage of the command
	Usage() string
	// Description of the command
	Description() string
	// Name of the command
	String() string
}

type cmd struct {
	name        string
	usage       string
	description string
	exec        func(args ...string) ([]byte, error)
}

func (c *cmd) Description() string {
	return c.description
}

func (c *cmd) Exec(args ...string) ([]byte, error) {
	return c.exec(args...)
}

func (c *cmd) Usage() string {
	return c.usage
}

func (c *cmd) String() string {
	return c.name
}

// NewCommand helps quickly create a new command
func NewCommand(name, usage, description string, exec func(args ...string) ([]byte, error)) Command {
	return &cmd{
		name:        name,
		usage:       usage,
		description: description,
		exec:        exec,
	}
}
