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

// Uint64Flag is a flag with type uint64
type Uint64Flag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	FilePath    string
	Required    bool
	Hidden      bool
	Value       uint64
	DefaultText string
	Destination *uint64
	HasBeenSet  bool
}

// IsSet returns whether or not the flag has been set through env or file
func (f *Uint64Flag) IsSet() bool {
	return f.HasBeenSet
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Uint64Flag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *Uint64Flag) Names() []string {
	return flagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *Uint64Flag) IsRequired() bool {
	return f.Required
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *Uint64Flag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *Uint64Flag) GetUsage() string {
	return f.Usage
}

// Apply populates the flag given the flag set and environment
func (f *Uint64Flag) Apply(set *flag.FlagSet) error {
	if val, ok := flagFromEnvOrFile(f.EnvVars, f.FilePath); ok {
		if val != "" {
			valInt, err := strconv.ParseUint(val, 0, 64)
			if err != nil {
				return fmt.Errorf("could not parse %q as uint64 value for flag %s: %s", val, f.Name, err)
			}

			f.Value = valInt
			f.HasBeenSet = true
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.Uint64Var(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.Uint64(name, f.Value, f.Usage)
	}

	return nil
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *Uint64Flag) GetValue() string {
	return fmt.Sprintf("%d", f.Value)
}

func (a *App) uint64Var(p *uint64, name, alias string, value uint64, usage, env string) {
	if a.Flags == nil {
		a.Flags = make([]Flag, 0)
	}
	flag := &Uint64Flag{
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

// Uint64Var defines a uint64 flag with specified name, default value, usage string and env string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func (a *App) Uint64Var(p *uint64, name string, value uint64, usage, env string) {
	a.uint64Var(p, name, "", value, usage, env)
}

// Uint64VarP is like Uint64Var, but accepts a shorthand letter that can be used after a single dash.
func (a *App) Uint64VarP(p *uint64, name, alias string, value uint64, usage, env string) {
	a.uint64Var(p, name, alias, value, usage, env)
}

// Uint64Var defines a uint64 flag with specified name, default value, usage string and env string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func Uint64Var(p *uint64, name string, value uint64, usage, env string) {
	CommandLine.Uint64Var(p, name, value, usage, env)
}

// Uint64VarP is like Uint64Var, but accepts a shorthand letter that can be used after a single dash.
func Uint64VarP(p *uint64, name, alias string, value uint64, usage, env string) {
	CommandLine.Uint64VarP(p, name, alias, value, usage, env)
}

// Uint64 defines a uint64 flag with specified name, default value, usage string and env string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func (a *App) Uint64(name string, value uint64, usage, env string) *uint64 {
	p := new(uint64)
	a.Uint64Var(p, name, value, usage, env)
	return p
}

// Uint64P is like Uint64, but accepts a shorthand letter that can be used after a single dash.
func (a *App) Uint64P(name, alias string, value uint64, usage, env string) *uint64 {
	p := new(uint64)
	a.Uint64VarP(p, name, alias, value, usage, env)
	return p
}

// Uint64 looks up the value of a local Uint64Flag, returns
// 0 if not found
func (c *Context) Uint64(name string) uint64 {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupUint64(name, fs)
	}
	return 0
}

func lookupUint64(name string, set *flag.FlagSet) uint64 {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := strconv.ParseUint(f.Value.String(), 0, 64)
		if err != nil {
			return 0
		}
		return parsed
	}
	return 0
}
