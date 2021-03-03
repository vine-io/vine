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

type Locking struct {
	Strength string
	Table    Table
	Options  string
}

// Name where clause name
func (locking Locking) Name() string {
	return "FOR"
}

// Build build where clause
func (locking Locking) Build(builder Builder) {
	builder.WriteString(locking.Strength)
	if locking.Table.Name != "" {
		builder.WriteString(" OF ")
		builder.WriteQuoted(locking.Table)
	}

	if locking.Options != "" {
		builder.WriteByte(' ')
		builder.WriteString(locking.Options)
	}
}

// MergeClause merge order by clauses
func (locking Locking) MergeClause(clause *Clause) {
	clause.Expression = locking
}
