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
