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

type Returning struct {
	Columns []Column
}

// Name where clause name
func (returning Returning) Name() string {
	return "RETURNING"
}

// Build build where clause
func (returning Returning) Build(builder Builder) {
	for idx, column := range returning.Columns {
		if idx > 0{
			builder.WriteByte(',')
		}

		builder.WriteQuoted(column)
	}
}

// MergeClause merge order by clauses
func (returning Returning) MergeClause(clause *Clause) {
	if v, ok := clause.Expression.(Returning); ok {
		returning.Columns = append(v.Columns, returning.Columns...)
	}

	clause.Expression = returning
}