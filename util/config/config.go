// MIT License
//
// Copyright (c) 2020 Lack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config

import (
	"io"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// FileName for global vine config, format yaml
var FileName = "vine.yml"

var excludes = []string{}

func isExclude(in string) bool {
	for _, e := range excludes {
		if strings.TrimSpace(in) == e {
			return true
		}
	}

	return false
}

func AddExclude(items ...string) {
	excludes = append(excludes, items...)
}

// config is a singleton which is required to ensure
// each function call doesn't load the .vine file
// from disk
var config = newConfig()

// SetConfigFile explicitly defines the path, name and extension of the config file.
func SetConfigFile(in string) {
	config.SetConfigName(in)
}

// SetConfigName sets name for the config file.
// Does not include extension.
func SetConfigName(in string) { config.SetConfigName(in) }

// SetConfigType sets the type of the configuration returned by the
// remote source, e.g. "json".
func SetConfigType(in string) { config.SetConfigType(in) }

// AddConfigPath adds a path to search for the config file in.
// Can be called multiple times to define multiple search paths.
func AddConfigPath(in string) { config.AddConfigPath(in) }

// ReadConfig reads a new configuration with an existing config.
func ReadConfig(in io.Reader) error { return config.ReadConfig(in) }

// ReadInConfig will discover and load the configuration file from disk
// and key/value stores, searching in one of the defined paths.
func ReadInConfig() error {
	return config.ReadInConfig()
}

// Get a value from the config
func Get(path ...string) interface{} {
	return config.Get(strings.Join(path, "."))
}

// GetString a string value from the config
func GetString(path ...string) string {
	return config.GetString(strings.Join(path, "."))
}

// GetInt a int64 value from the config
func GetInt(path ...string) int {
	return config.GetInt(strings.Join(path, "."))
}

// GetInt32 a int32 value from the config
func GetInt32(path ...string) int32 {
	return config.GetInt32(strings.Join(path, "."))
}

// GetInt64 a int64 value from the config
func GetInt64(path ...string) int64 {
	return config.GetInt64(strings.Join(path, "."))
}

// GetFloat a float value from the config
func GetFloat(path ...string) float64 {
	return config.GetFloat64(strings.Join(path, "."))
}

// GetBool a boolean value from the config
func GetBool(path ...string) bool {
	return config.GetBool(strings.Join(path, "."))
}

// GetStringSlice a string slice value from the config
func GetStringSlice(path ...string) []string {
	return config.GetStringSlice(strings.Join(path, "."))
}

// GetIntSlice a int slice value from the config
func GetIntSlice(path ...string) []int {
	return config.GetIntSlice(strings.Join(path, "."))
}

// GetStringMap a string map value from the config
func GetStringMap(path ...string) map[string]interface{} {
	return config.GetStringMap(strings.Join(path, "."))
}

// GetDuration a duration value from the config
func GetDuration(path ...string) time.Duration {
	return config.GetDuration(strings.Join(path, "."))
}

// Set a value in the config
func Set(value interface{}, path ...string) {
	config.Set(strings.Join(path, "."), value)
}

// UnmarshalKey takes a single key and unmarshals it into a Struct.
func UnmarshalKey(rawVal interface{}, path ...string) error {
	return config.UnmarshalKey(strings.Join(path, "."), rawVal)
}

// Unmarshal unmarshals the config into a Struct. Make sure that the tags
// on the fields of the structure are properly set.
func Unmarshal(rawVal interface{}) error {
	return config.Unmarshal(rawVal)
}

// Sync apply value to .vine config
func Sync() error {
	return config.WriteConfig()
}

// SyncAs writes current configuration to a given filename.
func SyncAs(filename string) error {
	return config.SafeWriteConfigAs(filename)
}

// BindPFlags binds a full flag set to the configuration, using each flag's long
// name as the config key.
func BindPFlags(flags *pflag.FlagSet) error {
	var err error
	flags.VisitAll(func(flag *pflag.Flag) {
		name := replace(flag.Name)
		if isExclude(name) {
			return
		}
		err = config.BindPFlag(name, flag)
		if err != nil {
			return
		}
	})
	return err
}

// newConfig returns a *viper.Viper
func newConfig() *viper.Viper {
	c := viper.New()
	c.SetConfigFile(FileName)
	c.SetConfigType("yaml")
	// return the conf
	return c
}

func replace(s string) string {
	return strings.ReplaceAll(s, "_", ".")
}
