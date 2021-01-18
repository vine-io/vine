// Copyright 2020 lack
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

package config

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	conf "github.com/lack-io/vine/service/config"
	"github.com/lack-io/vine/service/config/source/file"
	log "github.com/lack-io/vine/service/logger"
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
	c, err := conf.NewConfig(
		conf.WithSource(
			file.NewSource(
				file.WithPath(fp),
			),
		),
	)
	if err != nil {
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
