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
	"encoding/json"
	"flag"
	"fmt"
	"strings"
)

// StringSlice wraps a []string to satisfy flag.Value
type StringSlice struct {
	value      *[]string
	hasBeenSet bool
}

// NewStringSlice creates a *StringSlice with default values
func NewStringSlice(defaults ...string) *StringSlice {
	return newStringSlice(defaults, nil)
}

func newStringSlice(value []string, p *[]string) *StringSlice {
	slice := new(StringSlice)
	if p == nil {
		p = &[]string{}
	}
	slice.value = p
	*slice.value = value
	return slice
}

// Set appends the string value to the list of values
func (s *StringSlice) Set(value string) error {
	if !s.hasBeenSet {
		s.value = &[]string{}
		s.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &(*s.value))
		s.hasBeenSet = true
		return nil
	}

	tmp, err := stringSliceConv(value)
	if err != nil {
		return err
	}

	*s.value = append(*s.value, tmp...)

	return nil
}

func stringSliceConv(val string) ([]string, error) {
	val = strings.Trim(val, "[]")
	// Empty string would cause a slice with one (empty) entry
	if len(val) == 0 {
		return []string{}, nil
	}
	ss := strings.Split(val, ",")
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = strings.TrimSpace(s)
	}
	return out, nil
}

// String returns a readable representation of this value (for usage defaults)
func (s *StringSlice) String() string {
	if s.value == nil {
		return "[]"
	}
	out := make([]string, len(*s.value))
	for i, s := range *s.value {
		out[i] = strings.TrimSpace(s)
	}
	return "[" + strings.Join(out, ",") + "]"
}

// Serialize allows StringSlice to fulfill Serializer
func (s *StringSlice) Serialize() string {
	jsonBytes, _ := json.Marshal(*s.value)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of strings set by this flag
func (s *StringSlice) Value() []string {
	if s.value == nil {
		s.value = &[]string{}
	}
	return *s.value
}

// Get returns the slice of strings set by this flag
func (s *StringSlice) Get() interface{} {
	return *s
}

// StringSliceFlag is a flag with type *StringSlice
type StringSliceFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	FilePath    string
	Required    bool
	Hidden      bool
	TakesFile   bool
	Value       *StringSlice
	DefaultText string
	HasBeenSet  bool
}

// IsSet returns whether or not the flag has been set through env or file
func (f *StringSliceFlag) IsSet() bool {
	return f.HasBeenSet
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *StringSliceFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *StringSliceFlag) Names() []string {
	return flagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *StringSliceFlag) IsRequired() bool {
	return f.Required
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *StringSliceFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *StringSliceFlag) GetUsage() string {
	return f.Usage
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *StringSliceFlag) GetValue() string {
	if f.Value != nil {
		return f.Value.String()
	}
	return ""
}

// Apply populates the flag given the flag set and environment
func (f *StringSliceFlag) Apply(set *flag.FlagSet) error {
	if val, ok := flagFromEnvOrFile(f.EnvVars, f.FilePath); ok {
		f.Value = &StringSlice{}

		for _, s := range strings.Split(val, ",") {
			if err := f.Value.Set(strings.TrimSpace(s)); err != nil {
				return fmt.Errorf("could not parse %q as string value for flag %s: %s", val, f.Name, err)
			}
		}

		f.HasBeenSet = true
	}

	for _, name := range f.Names() {
		if f.Value == nil {
			f.Value = &StringSlice{}
		}
		set.Var(f.Value, name, f.Usage)
	}

	return nil
}

func (a *App) stringSliceVar(p *[]string, name, alias string, value []string, usage, env string) {
	if a.Flags == nil {
		a.Flags = make([]Flag, 0)
	}
	flag := &StringSliceFlag{
		Name:  name,
		Usage: usage,
		Value: newStringSlice(value, p),
	}
	if alias != "" {
		flag.Aliases = []string{alias}
	}
	if env != "" {
		flag.EnvVars = []string{env}
	}
	a.Flags = append(a.Flags, flag)
}

// StringSliceVar defines a []string flag with specified name, default value, usage string and env string.
// The argument p points to a []string variable in which to store the value of the flag.
func (a *App) StringSliceVar(p *[]string, name string, value []string, usage, env string) {
	a.stringSliceVar(p, name, "", value, usage, env)
}

// StringSliceVarP is like StringSliceVar, but accepts a shorthand letter that can be used after a single dash.
func (a *App) StringSliceVarP(p *[]string, name, alias string, value []string, usage, env string) {
	a.stringSliceVar(p, name, alias, value, usage, env)
}

// StringSliceVar defines a string flag with specified name, default value, usage string and env string.
// The argument p points to a string variable in which to store the value of the flag.
func StringSliceVar(p *[]string, name string, value []string, usage, env string) {
	CommandLine.StringSliceVar(p, name, value, usage, env)
}

// StringSliceVarP is like StringSliceVar, but accepts a shorthand letter that can be used after a single dash.
func StringSliceVarP(p *[]string, name, alias string, value []string, usage, env string) {
	CommandLine.StringSliceVarP(p, name, alias, value, usage, env)
}

// StringSlice defines a string flag with specified name, default value, usage string and env string.
// The return value is the address of a string variable that stores the value of the flag.
func (a *App) StringSlice(name string, value []string, usage, env string) *[]string {
	p := new([]string)
	a.StringSliceVar(p, name, value, usage, env)
	return p
}

// StringSliceP is like StringSlice, but accepts a shorthand letter that can be used after a single dash.
func (a *App) StringSliceP(name, alias string, value []string, usage, env string) *[]string {
	p := new([]string)
	a.StringSliceVarP(p, name, alias, value, usage, env)
	return p
}

// StringSlice looks up the value of a local StringSliceFlag, returns
// nil if not found
func (c *Context) StringSlice(name string) []string {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupStringSlice(name, fs)
	}
	return nil
}

func lookupStringSlice(name string, set *flag.FlagSet) []string {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := (f.Value.(*StringSlice)).Value(), error(nil)
		if err != nil {
			return nil
		}
		return parsed
	}
	return nil
}
