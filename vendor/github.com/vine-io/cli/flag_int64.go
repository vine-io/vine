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

// Int64Flag is a flag with type int64
type Int64Flag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	FilePath    string
	Required    bool
	Hidden      bool
	Value       int64
	DefaultText string
	Destination *int64
	HasBeenSet  bool
}

// IsSet returns whether or not the flag has been set through env or file
func (f *Int64Flag) IsSet() bool {
	return f.HasBeenSet
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Int64Flag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *Int64Flag) Names() []string {
	return flagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *Int64Flag) IsRequired() bool {
	return f.Required
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *Int64Flag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *Int64Flag) GetUsage() string {
	return f.Usage
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *Int64Flag) GetValue() string {
	return fmt.Sprintf("%d", f.Value)
}

// Apply populates the flag given the flag set and environment
func (f *Int64Flag) Apply(set *flag.FlagSet) error {
	if val, ok := flagFromEnvOrFile(f.EnvVars, f.FilePath); ok {
		if val != "" {
			valInt, err := strconv.ParseInt(val, 0, 64)

			if err != nil {
				return fmt.Errorf("could not parse %q as int value for flag %s: %s", val, f.Name, err)
			}

			f.Value = valInt
			f.HasBeenSet = true
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.Int64Var(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.Int64(name, f.Value, f.Usage)
	}
	return nil
}

func (a *App) int64Var(p *int64, name, alias string, value int64, usage, env string) {
	if a.Flags == nil {
		a.Flags = make([]Flag, 0)
	}
	flag := &Int64Flag{
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

// Int64Var defines a int64 flag with specified name, default value, usage string and env string.
// The argument p points to a int64 variable in which to store the value of the flag.
func (a *App) Int64Var(p *int64, name string, value int64, usage, env string) {
	a.int64Var(p, name, "", value, usage, env)
}

// Int64VarP is like Int64Var, but accepts a shorthand letter that can be used after a single dash.
func (a *App) Int64VarP(p *int64, name, alias string, value int64, usage, env string) {
	a.int64Var(p, name, alias, value, usage, env)
}

// Int64Var defines a int64 flag with specified name, default value, usage string and env string.
// The argument p points to a int64 variable in which to store the value of the flag.
func Int64Var(p *int64, name string, value int64, usage, env string) {
	CommandLine.Int64Var(p, name, value, usage, env)
}

// Int64VarP is like Int64Var, but accepts a shorthand letter that can be used after a single dash.
func Int64VarP(p *int64, name, alias string, value int64, usage, env string) {
	CommandLine.Int64VarP(p, name, alias, value, usage, env)
}

// Int64 defines a int64 flag with specified name, default value, usage string and env string.
// The return value is the address of a int64 variable that stores the value of the flag.
func (a *App) Int64(name string, value int64, usage, env string) *int64 {
	p := new(int64)
	a.Int64Var(p, name, value, usage, env)
	return p
}

// Int64P is like Int64, but accepts a shorthand letter that can be used after a single dash.
func (a *App) Int64P(name, alias string, value int64, usage, env string) *int64 {
	p := new(int64)
	a.Int64VarP(p, name, alias, value, usage, env)
	return p
}

// Int64 looks up the value of a local Int64Flag, returns
// 0 if not found
func (c *Context) Int64(name string) int64 {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupInt64(name, fs)
	}
	return 0
}

func lookupInt64(name string, set *flag.FlagSet) int64 {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := strconv.ParseInt(f.Value.String(), 0, 64)
		if err != nil {
			return 0
		}
		return parsed
	}
	return 0
}
