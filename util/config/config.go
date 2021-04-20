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
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	conf "github.com/lack-io/vine/lib/config"
	"github.com/lack-io/vine/lib/config/memory"
	"github.com/lack-io/vine/lib/config/source/file"
	log "github.com/lack-io/vine/lib/logger"
)

// FileName for global vine config
const FileName = ".vine"

// config is a singleton which is required to ensure
// each function call doesn't load the .vine file
// from disk
var config = newConfig()

// Get a value from the .vine file
func Get(path ...string) (string, error) {
	tk := config.Get(path...).String("")
	return strings.TrimSpace(tk), nil
}

// Set a value in the .vine file
func Set(value string, path ...string) error {
	// get the filepath
	fp, err := filePath()
	if err != nil {
		return err
	}

	// set the value
	config.Set(value, path...)

	// write to the file
	return ioutil.WriteFile(fp, config.Bytes(), 0644)
}

func filePath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, FileName), nil
}

// newConfig returns a loaded config
func newConfig() conf.Config {
	// get the filepath
	fp, err := filePath()
	if err != nil {
		log.Error(err)
		return conf.DefaultConfig
	}

	// write the file if it does not exist
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		ioutil.WriteFile(fp, []byte{}, 0644)
	} else if err != nil {
		log.Error(err)
		return conf.DefaultConfig
	}

	// create a new config
	c := memory.NewConfig(
		conf.WithSource(
			file.NewSource(
				file.WithPath(fp),
			),
		),
	)

	if err != c.Init() {
		log.Error(err)
		return conf.DefaultConfig
	}

	// load the config
	if err := c.Load(); err != nil {
		log.Error(err)
		return conf.DefaultConfig
	}

	// return the conf
	return c
}
