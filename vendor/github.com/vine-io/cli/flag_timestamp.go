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
	"time"
)

// Timestamp wrap to satisfy golang's flag interface.
type Timestamp struct {
	timestamp  *time.Time
	hasBeenSet bool
	layout     string
}

// Timestamp constructor
func NewTimestamp(timestamp time.Time) *Timestamp {
	return newTimestamp(timestamp, nil)
}

func newTimestamp(timestamp time.Time, p *time.Time) *Timestamp {
	t := new(Timestamp)
	if p == nil {
		p = new(time.Time)
	}
	t.timestamp = p
	*t.timestamp = timestamp
	return t
}

// Set the timestamp value directly
func (t *Timestamp) SetTimestamp(value time.Time) {
	if !t.hasBeenSet {
		t.timestamp = &value
		t.hasBeenSet = true
	}
}

// Set the timestamp string layout for future parsing
func (t *Timestamp) SetLayout(layout string) {
	t.layout = layout
}

// Parses the string value to timestamp
func (t *Timestamp) Set(value string) error {
	timestamp, err := time.Parse(t.layout, value)
	if err != nil {
		return err
	}

	t.timestamp = &timestamp
	t.hasBeenSet = true
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (t *Timestamp) String() string {
	return fmt.Sprintf("%#v", t.timestamp)
}

// Value returns the timestamp value stored in the flag
func (t *Timestamp) Value() *time.Time {
	return t.timestamp
}

// Get returns the flag structure
func (t *Timestamp) Get() interface{} {
	return *t
}

// TimestampFlag is a flag with type time
type TimestampFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	FilePath    string
	Required    bool
	Hidden      bool
	Layout      string
	Value       *Timestamp
	DefaultText string
	HasBeenSet  bool
}

// IsSet returns whether or not the flag has been set through env or file
func (f *TimestampFlag) IsSet() bool {
	return f.HasBeenSet
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *TimestampFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *TimestampFlag) Names() []string {
	return flagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *TimestampFlag) IsRequired() bool {
	return f.Required
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *TimestampFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *TimestampFlag) GetUsage() string {
	return f.Usage
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *TimestampFlag) GetValue() string {
	if f.Value != nil {
		return f.Value.timestamp.String()
	}
	return ""
}

// Apply populates the flag given the flag set and environment
func (f *TimestampFlag) Apply(set *flag.FlagSet) error {
	if f.Layout == "" {
		return fmt.Errorf("timestamp Layout is required")
	}
	f.Value = &Timestamp{}
	f.Value.SetLayout(f.Layout)

	if val, ok := flagFromEnvOrFile(f.EnvVars, f.FilePath); ok {
		if err := f.Value.Set(val); err != nil {
			return fmt.Errorf("could not parse %q as timestamp value for flag %s: %s", val, f.Name, err)
		}
		f.HasBeenSet = true
	}

	for _, name := range f.Names() {
		set.Var(f.Value, name, f.Usage)
	}
	return nil
}

func (a *App) timestampVar(p *time.Time, name, alias string, value time.Time, usage, env string) {
	if a.Flags == nil {
		a.Flags = make([]Flag, 0)
	}
	flag := &TimestampFlag{
		Name:  name,
		Usage: usage,
		Value: newTimestamp(value, p),
	}
	if alias != "" {
		flag.Aliases = []string{alias}
	}
	if env != "" {
		flag.EnvVars = []string{env}
	}
	a.Flags = append(a.Flags, flag)
}

// TimestampVar defines a time.Time flag with specified name, default value, usage string and env string.
// The argument p points to a time.Time variable in which to store the value of the flag.
func (a *App) TimestampVar(p *time.Time, name string, value time.Time, usage, env string) {
	a.timestampVar(p, name, "", value, usage, env)
}

// TimestampVarP is like TimestampVar, but accepts a shorthand letter that can be used after a single dash.
func (a *App) TimestampVarP(p *time.Time, name, alias string, value time.Time, usage, env string) {
	a.timestampVar(p, name, alias, value, usage, env)
}

// TimestampVar defines a Timestamp flag with specified name, default value, usage string and env string.
// The argument p points to a time.Time variable in which to store the value of the flag.
func TimestampVar(p *time.Time, name string, value time.Time, usage, env string) {
	CommandLine.TimestampVar(p, name, value, usage, env)
}

// TimestampVarP is like TimestampVar, but accepts a shorthand letter that can be used after a single dash.
func TimestampVarP(p *time.Time, name, alias string, value time.Time, usage, env string) {
	CommandLine.TimestampVarP(p, name, alias, value, usage, env)
}

// Timestamp defines a time.Time flag with specified name, default value, usage string and env string.
// The return value is the address of a time.Time variable that stores the value of the flag.
func (a *App) Timestamp(name string, value time.Time, usage, env string) *time.Time {
	p := new(time.Time)
	a.TimestampVar(p, name, value, usage, env)
	return p
}

// TimestampP is like Timestamp, but accepts a shorthand letter that can be used after a single dash.
func (a *App) TimestampP(name, alias string, value time.Time, usage, env string) *time.Time {
	p := new(time.Time)
	a.TimestampVarP(p, name, alias, value, usage, env)
	return p
}

// Timestamp gets the timestamp from a flag name
func (c *Context) Timestamp(name string) *time.Time {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupTimestamp(name, fs)
	}
	return nil
}

// Fetches the timestamp value from the local timestampWrap
func lookupTimestamp(name string, set *flag.FlagSet) *time.Time {
	f := set.Lookup(name)
	if f != nil {
		return (f.Value.(*Timestamp)).Value()
	}
	return nil
}
