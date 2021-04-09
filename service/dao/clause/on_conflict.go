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
