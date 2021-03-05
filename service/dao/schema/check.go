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

package schema

import (
	"regexp"
	"strings"
)

var (
	// match English letters and midline
	regEnLetterAndmidline = regexp.MustCompile("^[A-Za-z-_]+$")
)

type Check struct {
	Name       string
	Constraint string // length(phone) >= 10
	*Field
}

// ParseCheckConstraints parse schema check constraints
func (schema *Schema) ParseCheckConstraints() map[string]Check {
	var checks = map[string]Check{}
	for _, field := range schema.FieldsByDBName {
		if chk := field.TagSettings["CHECK"]; chk != "" {
			names := strings.Split(chk, ",")
			if len(names) > 1 && regEnLetterAndmidline.MatchString(names[0]) {
				checks[names[0]] = Check{Name: names[0], Constraint: strings.Join(names[1:], ","), Field: field}
			} else {
				if names[0] == "" {
					chk = strings.Join(names[1:], ",")
				}
				name := schema.namer.CheckerName(schema.Table, field.DBName)
				checks[name] = Check{Name: name, Constraint: chk, Field: field}
			}
		}
	}
	return checks
}
