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

// Float64Flag is a flag with type float64
type Float64Flag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	FilePath    string
	Required    bool
	Hidden      bool
	Value       float64
	DefaultText string
	Destination *float64
	HasBeenSet  bool
}

// IsSet returns whether or not the flag has been set through env or file
func (f *Float64Flag) IsSet() bool {
	return f.HasBeenSet
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Float64Flag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *Float64Flag) Names() []string {
	return flagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *Float64Flag) IsRequired() bool {
	return f.Required
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *Float64Flag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *Float64Flag) GetUsage() string {
	return f.Usage
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *Float64Flag) GetValue() string {
	return fmt.Sprintf("%f", f.Value)
}

// Apply populates the flag given the flag set and environment
func (f *Float64Flag) Apply(set *flag.FlagSet) error {
	if val, ok := flagFromEnvOrFile(f.EnvVars, f.FilePath); ok {
		if val != "" {
			valFloat, err := strconv.ParseFloat(val, 10)

			if err != nil {
				return fmt.Errorf("could not parse %q as float64 value for flag %s: %s", val, f.Name, err)
			}

			f.Value = valFloat
			f.HasBeenSet = true
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.Float64Var(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.Float64(name, f.Value, f.Usage)
	}

	return nil
}

func (a *App) float64Var(p *float64, name, alias string, value float64, usage, env string) {
	if a.Flags == nil {
		a.Flags = make([]Flag, 0)
	}
	flag := &Float64Flag{
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

// Float64Var defines a float64 flag with specified name, default value, usage string and env string.
// The argument p points to a float64 variable in which to store the value of the flag.
func (a *App) Float64Var(p *float64, name string, value float64, usage, env string) {
	a.float64Var(p, name, "", value, usage, env)
}

// Float64VarP is like Float64Var, but accepts a shorthand letter that can be used after a single dash.
func (a *App) Float64VarP(p *float64, name, alias string, value float64, usage, env string) {
	a.float64Var(p, name, alias, value, usage, env)
}

// Float64Var defines a float64 flag with specified name, default value, usage string and env string.
// The argument p points to a float64 variable in which to store the value of the flag.
func Float64Var(p *float64, name string, value float64, usage, env string) {
	CommandLine.Float64Var(p, name, value, usage, env)
}

// Float64VarP is like Float64Var, but accepts a shorthand letter that can be used after a single dash.
func Float64VarP(p *float64, name, alias string, value float64, usage, env string) {
	CommandLine.Float64VarP(p, name, alias, value, usage, env)
}

// Float64 defines a float64 flag with specified name, default value, usage string and env string.
// The return value is the address of a float64 variable that stores the value of the flag.
func (a *App) Float64(name string, value float64, usage, env string) *float64 {
	p := new(float64)
	a.Float64Var(p, name, value, usage, env)
	return p
}

// Float64P is like Float64, but accepts a shorthand letter that can be used after a single dash.
func (a *App) Float64P(name, alias string, value float64, usage, env string) *float64 {
	p := new(float64)
	a.Float64VarP(p, name, alias, value, usage, env)
	return p
}

// Float64 looks up the value of a local Float64Flag, returns
// 0 if not found
func (c *Context) Float64(name string) float64 {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupFloat64(name, fs)
	}
	return 0
}

func lookupFloat64(name string, set *flag.FlagSet) float64 {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := strconv.ParseFloat(f.Value.String(), 64)
		if err != nil {
			return 0
		}
		return parsed
	}
	return 0
}
