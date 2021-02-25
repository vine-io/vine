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
	"time"

	"github.com/lack-io/vine/service/config/reader"
)

type value struct{}

func newValue() reader.Value {
	return new(value)
}

func (v *value) Bool(def bool) bool {
	return false
}

func (v *value) Int(def int64) int64 {
	return 0
}

func (v *value) String(def string) string {
	return ""
}

func (v *value) Float64(def float64) float64 {
	return 0.0
}

func (v *value) Duration(def time.Duration) time.Duration {
	return time.Duration(0)
}

func (v *value) StringSlice(def []string) []string {
	return []string{}
}

func (v *value) StringMap(def map[string]string) map[string]string {
	return map[string]string{}
}

func (v *value) Scan(val interface{}) error {
	return nil
}

func (v *value) Bytes() []byte {
	return nil
}
