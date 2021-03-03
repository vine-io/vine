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

// GroupBy group by clause
type GroupBy struct {
	Columns []Column
	Having  []Expression
}

// Name from clause name
func (groupBy GroupBy) Name() string {
	return "GROUP BY"
}

// Build build group by clause
func (groupBy GroupBy) Build(builder Builder) {
	for idx, column := range groupBy.Columns {
		if idx > 0 {
			builder.WriteByte(',')
		}

		builder.WriteQuoted(column)
	}

	if len(groupBy.Having) > 0 {
		builder.WriteString(" HAVING ")
		Where{Exprs: groupBy.Having}.Build(builder)
	}
}

// MergeClause merge group by clause
func (groupBy GroupBy) MergeClause(clause *Clause) {
	if v, ok := clause.Expression.(GroupBy); ok {
		copidColumns := make([]Column, len(v.Columns))
		copy(copidColumns, v.Columns)
		groupBy.Columns = append(copidColumns, groupBy.Columns...)

		copidHaving := make([]Expression, len(v.Having))
		copy(copidHaving, v.Having)
		groupBy.Having = append(copidHaving, groupBy.Having...)
	}
	clause.Expression = groupBy
}
