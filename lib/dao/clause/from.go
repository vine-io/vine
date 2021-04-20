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

// From from clause
type From struct {
	Tables []Table
	Joins  []Join
}

// Name from clause name
func (from From) Name() string {
	return "FROM"
}

// Build build from clause
func (from From) Build(builder Builder) {
	if len(from.Tables) > 0 {
		for idx, table := range from.Tables {
			if idx > 0 {
				builder.WriteByte(',')
			}

			builder.WriteQuoted(table)
		}
	} else {
		builder.WriteQuoted(currentTable)
	}

	for _, join := range from.Joins {
		builder.WriteByte(' ')
		join.Build(builder)
	}
}

// MergeClause merge from clause
func (from From) MergeClause(clause *Clause) {
	clause.Expression = from
}
