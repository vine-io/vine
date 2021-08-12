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

// Float64Slice wraps []float64 to satisfy flag.Value
type Float64Slice struct {
	val        *[]float64
	hasBeenSet bool
}

// NewFloat64Slice makes a *Float64Slice with default values
func NewFloat64Slice(defaults ...float64) *Float64Slice {
	return newFloat64Slice(defaults, nil)
}

func newFloat64Slice(val []float64, p *[]float64) *Float64Slice {
	slice := new(Float64Slice)
	if p == nil {
		p = &[]float64{}
	}
	slice.val = p
	*slice.val = val
	return slice
}

// Set parses the value into a float64 and appends it to the list of values
func (f *Float64Slice) Set(value string) error {
	if !f.hasBeenSet {
		f.val = &[]float64{}
		f.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &(*f.val))
		f.hasBeenSet = true
		return nil
	}

	tmp, err := float64SliceConv(value)
	if err != nil {
		return err
	}

	*f.val = append(*f.val, tmp...)
	return nil
}

func float64SliceConv(val string) ([]float64, error) {
	val = strings.Trim(val, "[]")
	// Empty string would cause a slice with one (empty) entry
	if len(val) == 0 {
		return []float64{}, nil
	}
	ss := strings.Split(val, ",")
	out := make([]float64, len(ss))
	for i, d := range ss {
		var err error
		out[i], err = strconv.ParseFloat(strings.TrimSpace(d), 64)
		if err != nil {
			return nil, err
		}

	}
	return out, nil
}

// String returns a readable representation of this value (for usage defaults)
func (f *Float64Slice) String() string {
	if f.val == nil {
		return "[]"
	}
	out := make([]string, len(*f.val))
	for i, d := range *f.val {
		out[i] = fmt.Sprintf("%f", d)
	}
	return "[" + strings.Join(out, ",") + "]"
}

// Serialize allows Float64Slice to fulfill Serializer
func (f *Float64Slice) Serialize() string {
	jsonBytes, _ := json.Marshal(*f.val)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of float64s set by this flag
func (f *Float64Slice) Value() []float64 {
	if f.val == nil {
		f.val = &[]float64{}
	}
	return *f.val
}

// Get returns the slice of float64s set by this flag
func (f *Float64Slice) Get() interface{} {
	return *f
}

// Float64SliceFlag is a flag with type *Float64Slice
type Float64SliceFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	FilePath    string
	Required    bool
	Hidden      bool
	Value       *Float64Slice
	DefaultText string
	HasBeenSet  bool
}

// IsSet returns whether or not the flag has been set through env or file
func (f *Float64SliceFlag) IsSet() bool {
	return f.HasBeenSet
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Float64SliceFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *Float64SliceFlag) Names() []string {
	return flagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *Float64SliceFlag) IsRequired() bool {
	return f.Required
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *Float64SliceFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *Float64SliceFlag) GetUsage() string {
	return f.Usage
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *Float64SliceFlag) GetValue() string {
	if f.Value != nil {
		return f.Value.String()
	}
	return ""
}

// Apply populates the flag given the flag set and environment
func (f *Float64SliceFlag) Apply(set *flag.FlagSet) error {
	if val, ok := flagFromEnvOrFile(f.EnvVars, f.FilePath); ok {
		if val != "" {
			f.Value = &Float64Slice{}

			for _, s := range strings.Split(val, ",") {
				if err := f.Value.Set(strings.TrimSpace(s)); err != nil {
					return fmt.Errorf("could not parse %q as float64 slice value for flag %s: %s", f.Value, f.Name, err)
				}
			}

			f.HasBeenSet = true
		}
	}

	for _, name := range f.Names() {
		if f.Value == nil {
			f.Value = &Float64Slice{}
		}
		set.Var(f.Value, name, f.Usage)
	}

	return nil
}

func (a *App) float64SliceVar(p *[]float64, name, alias string, value []float64, usage, env string) {
	if a.Flags == nil {
		a.Flags = make([]Flag, 0)
	}
	flag := &Float64SliceFlag{
		Name:  name,
		Usage: usage,
		Value: newFloat64Slice(value, p),
	}
	if alias != "" {
		flag.Aliases = []string{alias}
	}
	if env != "" {
		flag.EnvVars = []string{env}
	}
	a.Flags = append(a.Flags, flag)
}

// Float64SliceVar defines a []float64 flag with specified name, default value, usage string and env string.
// The argument p points to a []float64 variable in which to store the value of the flag.
func (a *App) Float64SliceVar(p *[]float64, name string, value []float64, usage, env string) {
	a.float64SliceVar(p, name, "", value, usage, env)
}

// Float64SliceVarP is like Float64SliceVar, but accepts a shorthand letter that can be used after a single dash.
func (a *App) Float64SliceVarP(p *[]float64, name, alias string, value []float64, usage, env string) {
	a.float64SliceVar(p, name, alias, value, usage, env)
}

// Float64SliceVar defines a []float64 flag with specified name, default value, usage string and env string.
// The argument p points to a []float64 variable in which to store the value of the flag.
func Float64SliceVar(p *[]float64, name string, value []float64, usage, env string) {
	CommandLine.Float64SliceVar(p, name, value, usage, env)
}

// Float64SliceVarP is like Float64SliceVar, but accepts a shorthand letter that can be used after a single dash.
func Float64SliceVarP(p *[]float64, name, alias string, value []float64, usage, env string) {
	CommandLine.Float64SliceVarP(p, name, alias, value, usage, env)
}

// Float64Slice defines a []float64 flag with specified name, default value, usage string and env string.
// The return value is the address of a []float64 variable that stores the value of the flag.
func (a *App) Float64Slice(name string, value []float64, usage, env string) *[]float64 {
	p := new([]float64)
	a.Float64SliceVar(p, name, value, usage, env)
	return p
}

// Float64SliceP is like Float64Slice, but accepts a shorthand letter that can be used after a single dash.
func (a *App) Float64SliceP(name, alias string, value []float64, usage, env string) *[]float64 {
	p := new([]float64)
	a.Float64SliceVarP(p, name, alias, value, usage, env)
	return p
}

// Float64Slice looks up the value of a local Float64SliceFlag, returns
// nil if not found
func (c *Context) Float64Slice(name string) []float64 {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupFloat64Slice(name, fs)
	}
	return nil
}

func lookupFloat64Slice(name string, set *flag.FlagSet) []float64 {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := (f.Value.(*Float64Slice)).Value(), error(nil)
		if err != nil {
			return nil
		}
		return parsed
	}
	return nil
}
