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

type OnConflict struct {
	Columns      []Column
	Where        Where
	OnConstraint string
	DoNothing    bool
	DoUpdates    Set
	UpdateAll    bool
}

func (OnConflict) Name() string {
	return "ON CONFLICT"
}

// Build build onConflict clause
func (onConflict OnConflict) Build(builder Builder) {
	if len(onConflict.Columns) > 0 {
		builder.WriteByte('(')
		for idx, column := range onConflict.Columns {
			if idx > 0 {
				builder.WriteByte(',')
			}
			builder.WriteQuoted(column)
		}
		builder.WriteString(`) `)
	}

	if onConflict.OnConstraint != "" {
		builder.WriteString("ON CONSTRAINT ")
		builder.WriteString(onConflict.OnConstraint)
		builder.WriteByte(' ')
	}

	if onConflict.DoNothing {
		builder.WriteString("DO NOTHING")
	} else {
		builder.WriteString("DO UPDATE SET ")
		onConflict.DoUpdates.Build(builder)
	}

	if len(onConflict.Where.Exprs) > 0 {
		builder.WriteString("WHERE ")
		onConflict.Where.Build(builder)
		builder.WriteByte(' ')
	}
}

// MergeClause merge onConflict clauses
func (onConflict OnConflict) MergeClause(clause *Clause) {
	clause.Expression = onConflict
}
