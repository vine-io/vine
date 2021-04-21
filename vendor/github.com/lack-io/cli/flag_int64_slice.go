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
	"strconv"
	"strings"
)

// Int64Slice wraps []int64 to satisfy flag.Value
type Int64Slice struct {
	value      *[]int64
	hasBeenSet bool
}

// NewInt64Slice makes an *Int64Slice with default values
func NewInt64Slice(defaults ...int64) *Int64Slice {
	return newInt64Slice(defaults, nil)
}

func newInt64Slice(value []int64, p *[]int64) *Int64Slice {
	slice := new(Int64Slice)
	if p == nil {
		p = &[]int64{}
	}
	slice.value = p
	*slice.value = value
	return slice
}

// Set parses the value into an integer and appends it to the list of values
func (i *Int64Slice) Set(value string) error {
	if !i.hasBeenSet {
		i.value = &[]int64{}
		i.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &(*i.value))
		i.hasBeenSet = true
		return nil
	}

	tmp, err := int64SliceConv(value)
	if err != nil {
		return err
	}
	*i.value = append(*i.value, tmp...)

	return nil
}

func int64SliceConv(val string) ([]int64, error) {
	val = strings.Trim(val, "[]")
	// Empty string would cause a slice with one (empty) entry
	if len(val) == 0 {
		return []int64{}, nil
	}
	ss := strings.Split(val, ",")
	out := make([]int64, len(ss))
	for i, d := range ss {
		var err error
		out[i], err = strconv.ParseInt(strings.TrimSpace(d), 10, 64)
		if err != nil {
			return nil, err
		}

	}
	return out, nil
}

// String returns a readable representation of this value (for usage defaults)
func (i *Int64Slice) String() string {
	if i.value == nil {
		return "[]"
	}
	out := make([]string, len(*i.value))
	for i, d := range *i.value {
		out[i] = fmt.Sprintf("%d", d)
	}
	return "[" + strings.Join(out, ",") + "]"
}

// Serialize allows Int64Slice to fulfill Serializer
func (i *Int64Slice) Serialize() string {
	jsonBytes, _ := json.Marshal(*i.value)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of ints set by this flag
func (i *Int64Slice) Value() []int64 {
	if i.value == nil {
		i.value = &[]int64{}
	}
	return *i.value
}

// Get returns the slice of ints set by this flag
func (i *Int64Slice) Get() interface{} {
	return *i
}

// Int64SliceFlag is a flag with type *Int64Slice
type Int64SliceFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	FilePath    string
	Required    bool
	Hidden      bool
	Value       *Int64Slice
	DefaultText string
	HasBeenSet  bool
}

// IsSet returns whether or not the flag has been set through env or file
func (f *Int64SliceFlag) IsSet() bool {
	return f.HasBeenSet
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Int64SliceFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *Int64SliceFlag) Names() []string {
	return flagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *Int64SliceFlag) IsRequired() bool {
	return f.Required
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *Int64SliceFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f Int64SliceFlag) GetUsage() string {
	return f.Usage
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *Int64SliceFlag) GetValue() string {
	if f.Value != nil {
		return f.Value.String()
	}
	return ""
}

// Apply populates the flag given the flag set and environment
func (f *Int64SliceFlag) Apply(set *flag.FlagSet) error {
	if val, ok := flagFromEnvOrFile(f.EnvVars, f.FilePath); ok {
		f.Value = &Int64Slice{}

		for _, s := range strings.Split(val, ",") {
			if err := f.Value.Set(strings.TrimSpace(s)); err != nil {
				return fmt.Errorf("could not parse %q as int64 slice value for flag %s: %s", val, f.Name, err)
			}
		}

		f.HasBeenSet = true
	}

	for _, name := range f.Names() {
		if f.Value == nil {
			f.Value = &Int64Slice{}
		}
		set.Var(f.Value, name, f.Usage)
	}

	return nil
}

func (a *App) int64SliceVar(p *[]int64, name, alias string, value []int64, usage, env string) {
	if a.Flags == nil {
		a.Flags = make([]Flag, 0)
	}
	flag := &Int64SliceFlag{
		Name:  name,
		Usage: usage,
		Value: newInt64Slice(value, p),
	}
	if alias != "" {
		flag.Aliases = []string{alias}
	}
	if env != "" {
		flag.EnvVars = []string{env}
	}
	a.Flags = append(a.Flags, flag)
}

// Int64SliceVar defines a []int64 flag with specified name, default value, usage string and env string.
// The argument p points to a []int64 variable in which to store the value of the flag.
func (a *App) Int64SliceVar(p *[]int64, name string, value []int64, usage, env string) {
	a.int64SliceVar(p, name, "", value, usage, env)
}

// Int64SliceVarP is like Int64SliceVar, but accepts a shorthand letter that can be used after a single dash.
func (a *App) Int64SliceVarP(p *[]int64, name, alias string, value []int64, usage, env string) {
	a.int64SliceVar(p, name, alias, value, usage, env)
}

// Int64SliceVar defines a []int64 flag with specified name, default value, usage string and env string.
// The argument p points to a []int64 variable in which to store the value of the flag.
func Int64SliceVar(p *[]int64, name string, value []int64, usage, env string) {
	CommandLine.Int64SliceVar(p, name, value, usage, env)
}

// Int64SliceVarP is like Int64SliceVar, but accepts a shorthand letter that can be used after a single dash.
func Int64SliceVarP(p *[]int64, name, alias string, value []int64, usage, env string) {
	CommandLine.Int64SliceVarP(p, name, alias, value, usage, env)
}

// Int64Slice defines a []int64 flag with specified name, default value, usage string and env string.
// The return value is the address of a []int64 variable that stores the value of the flag.
func (a *App) Int64Slice(name string, value []int64, usage, env string) *[]int64 {
	p := new([]int64)
	a.Int64SliceVar(p, name, value, usage, env)
	return p
}

// Int64SliceP is like Int64Slice, but accepts a shorthand letter that can be used after a single dash.
func (a *App) Int64SliceP(name, alias string, value []int64, usage, env string) *[]int64 {
	p := new([]int64)
	a.Int64SliceVarP(p, name, alias, value, usage, env)
	return p
}

// Int64Slice looks up the value of a local Int64SliceFlag, returns
// nil if not found
func (c *Context) Int64Slice(name string) []int64 {
	return lookupInt64Slice(name, c.flagSet)
}

func lookupInt64Slice(name string, set *flag.FlagSet) []int64 {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := (f.Value.(*Int64Slice)).Value(), error(nil)
		if err != nil {
			return nil
		}
		return parsed
	}
	return nil
}
