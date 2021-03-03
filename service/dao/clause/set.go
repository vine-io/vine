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

package clause

import "sort"

type Set []Assignment

type Assignment struct {
	Column Column
	Value  interface{}
}

func (set Set) Name() string {
	return "SET"
}

func (set Set) Build(builder Builder) {
	if len(set) > 0 {
		for idx, assignment := range set {
			if idx > 0 {
				builder.WriteByte(',')
			}
			builder.WriteQuoted(assignment.Column)
			builder.WriteByte('=')
			builder.AddVar(builder, assignment.Value)
		}
	} else {
		builder.WriteQuoted(PrimaryColumn)
		builder.WriteByte('=')
		builder.WriteQuoted(PrimaryColumn)
	}
}

// MergeClause merge assignment clauses
func (set Set) MergeClause(clause *Clause) {
	copiedAssignments := make([]Assignment, len(set))
	copy(copiedAssignments, set)
	clause.Expression = Set(copiedAssignments)
}

func Assignments(values map[string]interface{}) Set {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	assignments := make([]Assignment, len(keys))
	for idx, key := range keys {
		assignments[idx] = Assignment{Column: Column{Name: key}, Value: values[key]}
	}
	return assignments
}

func AssignmentColumns(values []string) Set {
	assignments := make([]Assignment, len(values))
	for idx, value := range values {
		assignments[idx] = Assignment{Column: Column{Name: value}, Value: Column{Table: "excluded", Name: value}}
	}
	return assignments
}
