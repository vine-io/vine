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

type Insert struct {
	Table    Table
	Modifier string
}

// Name insert clause name
func (insert Insert) Name() string {
	return "INSERT"
}

// Build build insert clause
func (insert Insert) Build(builder Builder) {
	if insert.Modifier != "" {
		builder.WriteString(insert.Modifier)
		builder.WriteByte(' ')
	}

	builder.WriteString("INTO ")
	if insert.Table.Name == "" {
		builder.WriteQuoted(currentTable)
	} else {
		builder.WriteQuoted(insert.Table)
	}
}

// MergeClause merge insert clause
func (insert Insert) MergeClause(clause *Clause) {
	if v, ok := clause.Expression.(Insert); ok {
		if insert.Modifier == "" {
			insert.Modifier = v.Modifier
		}
		if insert.Table.Name == "" {
			insert.Table = v.Table
		}
	}
	clause.Expression = insert
}
