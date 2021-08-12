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

// IntSlice wraps []int to satisfy flag.Value
type IntSlice struct {
	value      *[]int
	hasBeenSet bool
}

// NewIntSlice makes an *IntSlice with default values
func NewIntSlice(defaults ...int) *IntSlice {
	return newIntSlice(defaults, nil)
}

func newIntSlice(value []int, p *[]int) *IntSlice {
	slice := new(IntSlice)
	if p == nil {
		p = &[]int{}
	}
	slice.value = p
	*slice.value = value
	return slice
}

// TODO: Consistently have specific Set function for Int64 and Float64 ?
// SetInt directly adds an integer to the list of values
func (i *IntSlice) SetInt(value int) {
	if !i.hasBeenSet {
		i.value = &[]int{}
		i.hasBeenSet = true
	}

	*i.value = append(*i.value, value)
}

// Set parses the value into an integer and appends it to the list of values
func (i *IntSlice) Set(value string) error {
	if !i.hasBeenSet {
		i.value = &[]int{}
		i.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &(*i.value))
		i.hasBeenSet = true
		return nil
	}

	tmp, err := intSliceConv(value)
	if err != nil {
		return err
	}

	*i.value = append(*i.value, tmp...)

	return nil
}

func intSliceConv(val string) ([]int, error) {
	val = strings.Trim(val, "[]")
	// Empty string would cause a slice with one (empty) entry
	if len(val) == 0 {
		return []int{}, nil
	}
	ss := strings.Split(val, ",")
	out := make([]int, len(ss))
	for i, d := range ss {
		var err error
		i64, err := strconv.ParseInt(strings.TrimSpace(d), 10, 64)
		if err != nil {
			return nil, err
		}
		out[i] = int(i64)
	}
	return out, nil
}

// String returns a readable representation of this value (for usage defaults)
func (i *IntSlice) String() string {
	if i.value == nil {
		return "[]"
	}
	out := make([]string, len(*i.value))
	for i, d := range *i.value {
		out[i] = fmt.Sprintf("%d", d)
	}
	return "[" + strings.Join(out, ",") + "]"
}

// Serialize allows IntSlice to fulfill Serializer
func (i *IntSlice) Serialize() string {
	jsonBytes, _ := json.Marshal(*i.value)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of ints set by this flag
func (i *IntSlice) Value() []int {
	if i.value == nil {
		i.value = &[]int{}
	}
	return *i.value
}

// Get returns the slice of ints set by this flag
func (i *IntSlice) Get() interface{} {
	return *i
}

// IntSliceFlag is a flag with type *IntSlice
type IntSliceFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	FilePath    string
	Required    bool
	Hidden      bool
	Value       *IntSlice
	DefaultText string
	HasBeenSet  bool
}

// IsSet returns whether or not the flag has been set through env or file
func (f *IntSliceFlag) IsSet() bool {
	return f.HasBeenSet
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *IntSliceFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *IntSliceFlag) Names() []string {
	return flagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *IntSliceFlag) IsRequired() bool {
	return f.Required
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *IntSliceFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f IntSliceFlag) GetUsage() string {
	return f.Usage
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *IntSliceFlag) GetValue() string {
	if f.Value != nil {
		return f.Value.String()
	}
	return ""
}

// Apply populates the flag given the flag set and environment
func (f *IntSliceFlag) Apply(set *flag.FlagSet) error {
	if val, ok := flagFromEnvOrFile(f.EnvVars, f.FilePath); ok {
		f.Value = &IntSlice{}

		for _, s := range strings.Split(val, ",") {
			if err := f.Value.Set(strings.TrimSpace(s)); err != nil {
				return fmt.Errorf("could not parse %q as int slice value for flag %s: %s", val, f.Name, err)
			}
		}

		f.HasBeenSet = true
	}

	for _, name := range f.Names() {
		if f.Value == nil {
			f.Value = &IntSlice{}
		}
		set.Var(f.Value, name, f.Usage)
	}

	return nil
}

func (a *App) intSliceVar(p *[]int, name, alias string, value []int, usage, env string) {
	if a.Flags == nil {
		a.Flags = make([]Flag, 0)
	}
	flag := &IntSliceFlag{
		Name:  name,
		Usage: usage,
		Value: newIntSlice(value, p),
	}
	if alias != "" {
		flag.Aliases = []string{alias}
	}
	if env != "" {
		flag.EnvVars = []string{env}
	}
	a.Flags = append(a.Flags, flag)
}

// IntSliceVar defines a []int flag with specified name, default value, usage string and env string.
// The argument p points to a []int variable in which to store the value of the flag.
func (a *App) IntSliceVar(p *[]int, name string, value []int, usage, env string) {
	a.intSliceVar(p, name, "", value, usage, env)
}

// IntSliceVarP is like IntSliceVar, but accepts a shorthand letter that can be used after a single dash.
func (a *App) IntSliceVarP(p *[]int, name, alias string, value []int, usage, env string) {
	a.intSliceVar(p, name, alias, value, usage, env)
}

// IntSliceVar defines a []int flag with specified name, default value, usage string and env string.
// The argument p points to a []int variable in which to store the value of the flag.
func IntSliceVar(p *[]int, name string, value []int, usage, env string) {
	CommandLine.IntSliceVar(p, name, value, usage, env)
}

// IntSliceVarP is like IntSliceVar, but accepts a shorthand letter that can be used after a single dash.
func IntSliceVarP(p *[]int, name, alias string, value []int, usage, env string) {
	CommandLine.IntSliceVarP(p, name, alias, value, usage, env)
}

// IntSlice defines a []int flag with specified name, default value, usage string and env string.
// The return value is the address of a []int variable that stores the value of the flag.
func (a *App) IntSlice(name string, value []int, usage, env string) *[]int {
	p := new([]int)
	a.IntSliceVar(p, name, value, usage, env)
	return p
}

// IntSliceP is like IntSlice, but accepts a shorthand letter that can be used after a single dash.
func (a *App) IntSliceP(name, alias string, value []int, usage, env string) *[]int {
	p := new([]int)
	a.IntSliceVarP(p, name, alias, value, usage, env)
	return p
}

// IntSlice looks up the value of a local IntSliceFlag, returns
// nil if not found
func (c *Context) IntSlice(name string) []int {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupIntSlice(name, c.flagSet)
	}
	return nil
}

func lookupIntSlice(name string, set *flag.FlagSet) []int {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := (f.Value.(*IntSlice)).Value(), error(nil)
		if err != nil {
			return nil
		}
		return parsed
	}
	return nil
}
