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
