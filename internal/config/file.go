// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// Version represents ./vine/version
type Version struct {
	Version string    `json:"version"`
	Updated time.Time `json:"updated"`
}

// WriteVersion will write a version update to a file ./vine/version.
// This indicates the current version and the last time we updated the binary.
// Its only used where self update is
func WriteVersion(v string) error {
	dir := filepath.Dir(File)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	b, err := json.Marshal(&Version{
		Version: v,
		Updated: time.Now(),
	})
	if err != nil {
		return err
	}
	f := filepath.Join(dir, "version")
	return ioutil.WriteFile(f, b, 0644)
}

// GetVersion returns the version from .vine/version. If it does not exist
func GetVersion() (*Version, error) {
	dir := filepath.Dir(File)
	f := filepath.Join(dir, "version")
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	v := new(Version)
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return v, nil
}
