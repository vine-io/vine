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
)

// StringFlag is a flag with type string
type StringFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	FilePath    string
	Required    bool
	Hidden      bool
	TakesFile   bool
	Value       string
	DefaultText string
	Destination *string
	HasBeenSet  bool
}

// IsSet returns whether or not the flag has been set through env or file
func (f *StringFlag) IsSet() bool {
	return f.HasBeenSet
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *StringFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *StringFlag) Names() []string {
	return flagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *StringFlag) IsRequired() bool {
	return f.Required
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *StringFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *StringFlag) GetUsage() string {
	return f.Usage
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *StringFlag) GetValue() string {
	return f.Value
}

// Apply populates the flag given the flag set and environment
func (f *StringFlag) Apply(set *flag.FlagSet) error {
	if val, ok := flagFromEnvOrFile(f.EnvVars, f.FilePath); ok {
		f.Value = val
		f.HasBeenSet = true
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.StringVar(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.String(name, f.Value, f.Usage)
	}

	return nil
}

func (a *App) stringVar(p *string, name, alias string, value string, usage, env string) {
	if a.Flags == nil {
		a.Flags = make([]Flag, 0)
	}
	flag := &StringFlag{
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

// StringVar defines a string flag with specified name, default value, usage string and env string.
// The argument p points to a string variable in which to store the value of the flag.
func (a *App) StringVar(p *string, name string, value string, usage, env string) {
	a.stringVar(p, name, "", value, usage, env)
}

// StringVarP is like StringVar, but accepts a shorthand letter that can be used after a single dash.
func (a *App) StringVarP(p *string, name, alias string, value string, usage, env string) {
	a.stringVar(p, name, alias, value, usage, env)
}

// StringVar defines a string flag with specified name, default value, usage string and env string.
// The argument p points to a string variable in which to store the value of the flag.
func StringVar(p *string, name string, value string, usage, env string) {
	CommandLine.StringVar(p, name, value, usage, env)
}

// StringVarP is like StringVar, but accepts a shorthand letter that can be used after a single dash.
func StringVarP(p *string, name, alias string, value string, usage, env string) {
	CommandLine.StringVarP(p, name, alias, value, usage, env)
}

// String defines a string flag with specified name, default value, usage string and env string.
// The return value is the address of a string variable that stores the value of the flag.
func (a *App) String(name string, value string, usage, env string) *string {
	p := new(string)
	a.StringVar(p, name, value, usage, env)
	return p
}

// StringP is like String, but accepts a shorthand letter that can be used after a single dash.
func (a *App) StringP(name, alias string, value string, usage, env string) *string {
	p := new(string)
	a.StringVarP(p, name, alias, value, usage, env)
	return p
}

// String looks up the value of a local StringFlag, returns
// "" if not found
func (c *Context) String(name string) string {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupString(name, fs)
	}
	return ""
}

func lookupString(name string, set *flag.FlagSet) string {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := f.Value.String(), error(nil)
		if err != nil {
			return ""
		}
		return parsed
	}
	return ""
}
