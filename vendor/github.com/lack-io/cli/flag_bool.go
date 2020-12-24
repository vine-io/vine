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

package cli

import (
	"flag"
	"fmt"
	"strconv"
)

// BoolFlag is a flag with type bool
type BoolFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	FilePath    string
	Required    bool
	Hidden      bool
	Value       bool
	DefaultText string
	Destination *bool
	HasBeenSet  bool
}

// IsSet returns whether or not the flag has been set through env or file
func (f *BoolFlag) IsSet() bool {
	return f.HasBeenSet
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *BoolFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *BoolFlag) Names() []string {
	return flagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *BoolFlag) IsRequired() bool {
	return f.Required
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *BoolFlag) TakesValue() bool {
	return false
}

// GetUsage returns the usage string for the flag
func (f *BoolFlag) GetUsage() string {
	return f.Usage
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *BoolFlag) GetValue() string {
	return ""
}

// Apply populates the flag given the flag set and environment
func (f *BoolFlag) Apply(set *flag.FlagSet) error {
	if val, ok := flagFromEnvOrFile(f.EnvVars, f.FilePath); ok {
		if val != "" {
			valBool, err := strconv.ParseBool(val)

			if err != nil {
				return fmt.Errorf("could not parse %q as bool value for flag %s: %s", val, f.Name, err)
			}

			f.Value = valBool
			f.HasBeenSet = true
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.BoolVar(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.Bool(name, f.Value, f.Usage)
	}

	return nil
}

func (a *App) boolVar(p *bool, name, alias string, value bool, usage, env string) {
	if a.Flags == nil {
		a.Flags = make([]Flag, 0)
	}
	flag := &BoolFlag{
		Name:        name,
		Usage:       usage,
		Value:       value,
		Destination: p,
	}
	if alias != "" {
		flag.Aliases = []string{alias}
	}
	if env != "" {
		flag.EnvVars = []string{env}
	}
	a.Flags = append(a.Flags, flag)
}

// BoolVar defines a bool flag with specified name, default value, usage string and env string.
// The argument p points to a bool variable in which to store the value of the flag.
func (a *App) BoolVar(p *bool, name string, value bool, usage, env string) {
	a.boolVar(p, name, "", value, usage, env)
}

// BoolVarP is like BoolVar, but accepts a shorthand letter that can be used after a single dash.
func (a *App) BoolVarP(p *bool, name, alias string, value bool, usage, env string) {
	a.boolVar(p, name, alias, value, usage, env)
}

// BoolVar defines a bool flag with specified name, default value, usage string and env string.
// The argument p points to a bool variable in which to store the value of the flag.
func BoolVar(p *bool, name string, value bool, usage, env string) {
	CommandLine.BoolVar(p, name, value, usage, env)
}

// BoolVarP is like BoolVar, but accepts a shorthand letter that can be used after a single dash.
func BoolVarP(p *bool, name, alias string, value bool, usage, env string) {
	CommandLine.BoolVarP(p, name, alias, value, usage, env)
}

// Bool defines a bool flag with specified name, default value, usage string and env string.
// The return value is the address of a bool variable that stores the value of the flag.
func (a *App) Bool(name string, value bool, usage, env string) *bool {
	p := new(bool)
	a.BoolVar(p, name, value, usage, env)
	return p
}

// BoolP is like Bool, but accepts a shorthand letter that can be used after a single dash.
func (a *App) BoolP(name, alias string, value bool, usage, env string) *bool {
	p := new(bool)
	a.BoolVarP(p, name, alias, value, usage, env)
	return p
}

// Bool looks up the value of a local BoolFlag, returns
// false if not found
func (c *Context) Bool(name string) bool {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupBool(name, fs)
	}
	return false
}

func lookupBool(name string, set *flag.FlagSet) bool {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := strconv.ParseBool(f.Value.String())
		if err != nil {
			return false
		}
		return parsed
	}
	return false
}
