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
