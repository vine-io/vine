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

// UintFlag is a flag with type uint
type UintFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	FilePath    string
	Required    bool
	Hidden      bool
	Value       uint
	DefaultText string
	Destination *uint
	HasBeenSet  bool
}

// IsSet returns whether or not the flag has been set through env or file
func (f *UintFlag) IsSet() bool {
	return f.HasBeenSet
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *UintFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *UintFlag) Names() []string {
	return flagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *UintFlag) IsRequired() bool {
	return f.Required
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *UintFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *UintFlag) GetUsage() string {
	return f.Usage
}

// Apply populates the flag given the flag set and environment
func (f *UintFlag) Apply(set *flag.FlagSet) error {
	if val, ok := flagFromEnvOrFile(f.EnvVars, f.FilePath); ok {
		if val != "" {
			valInt, err := strconv.ParseUint(val, 0, 64)
			if err != nil {
				return fmt.Errorf("could not parse %q as uint value for flag %s: %s", val, f.Name, err)
			}

			f.Value = uint(valInt)
			f.HasBeenSet = true
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.UintVar(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.Uint(name, f.Value, f.Usage)
	}

	return nil
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *UintFlag) GetValue() string {
	return fmt.Sprintf("%d", f.Value)
}

func (a *App) uintVar(p *uint, name, alias string, value uint, usage, env string) {
	if a.Flags == nil {
		a.Flags = make([]Flag, 0)
	}
	flag := &UintFlag{
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

// UintVar defines a uint flag with specified name, default value, usage string and env string.
// The argument p points to a uint variable in which to store the value of the flag.
func (a *App) UintVar(p *uint, name string, value uint, usage, env string) {
	a.uintVar(p, name, "", value, usage, env)
}

// UintVarP is like UintVar, but accepts a shorthand letter that can be used after a single dash.
func (a *App) UintVarP(p *uint, name, alias string, value uint, usage, env string) {
	a.uintVar(p, name, alias, value, usage, env)
}

// UintVar defines a uint flag with specified name, default value, usage string and env string.
// The argument p points to a uint variable in which to store the value of the flag.
func UintVar(p *uint, name string, value uint, usage, env string) {
	CommandLine.UintVar(p, name, value, usage, env)
}

// UintVarP is like UintVar, but accepts a shorthand letter that can be used after a single dash.
func UintVarP(p *uint, name, alias string, value uint, usage, env string) {
	CommandLine.UintVarP(p, name, alias, value, usage, env)
}

// Uint defines a uint flag with specified name, default value, usage string and env string.
// The return value is the address of a uint variable that stores the value of the flag.
func (a *App) Uint(name string, value uint, usage, env string) *uint {
	p := new(uint)
	a.UintVar(p, name, value, usage, env)
	return p
}

// UintP is like Uint, but accepts a shorthand letter that can be used after a single dash.
func (a *App) UintP(name, alias string, value uint, usage, env string) *uint {
	p := new(uint)
	a.UintVarP(p, name, alias, value, usage, env)
	return p
}

// Uint looks up the value of a local UintFlag, returns
// 0 if not found
func (c *Context) Uint(name string) uint {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupUint(name, fs)
	}
	return 0
}

func lookupUint(name string, set *flag.FlagSet) uint {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := strconv.ParseUint(f.Value.String(), 0, 64)
		if err != nil {
			return 0
		}
		return uint(parsed)
	}
	return 0
}
