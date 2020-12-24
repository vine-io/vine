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
	"github.com/lack-io/cli"
)

// Options are used as part of a new plugin
type Options struct {
	Name     string
	Flags    []cli.Flag
	Commands []*cli.Command
	Handlers []Handler
	Init     func(*cli.Context) error
}

type Option func(o *Options)

// WithFlag adds flags to a plugin
func WithFlag(flag ...cli.Flag) Option {
	return func(o *Options) {
		o.Flags = append(o.Flags, flag...)
	}
}

// WithCommand adds commands to a plugin
func WithCommand(cmd ...*cli.Command) Option {
	return func(o *Options) {
		o.Commands = append(o.Commands, cmd...)
	}
}

// WithHandler adds middleware handlers to
func WithHandler(h ...Handler) Option {
	return func(o *Options) {
		o.Handlers = append(o.Handlers, h...)
	}
}

// WithName defines the name of the plugin
func WithName(n string) Option {
	return func(o *Options) {
		o.Name = n
	}
}

// WithInit sets the init function
func WithInit(fn func(*cli.Context) error) Option {
	return func(o *Options) {
		o.Init = fn
	}
}
