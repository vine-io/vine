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

package sqlite

import (
	"fmt"
	"strings"

	"github.com/lack-io/vine/service/dao"
	"github.com/lack-io/vine/service/dao/clause"
)

type jsonQueryExpression struct {
	tx          *dao.DB
	op          dao.JSONOp
	column      string
	hasKeys     bool
	keys        []string
	equalsValue interface{}
}

func JSONQuery(column string) *jsonQueryExpression {
	return &jsonQueryExpression{column: column}
}

func (j *jsonQueryExpression) Tx(tx *dao.DB) dao.JSONQuery {
	j.tx = tx
	return j
}

func (j *jsonQueryExpression) Op(op dao.JSONOp, value interface{}, keys ...string) dao.JSONQuery {
	j.op = op
	j.keys = keys
	j.equalsValue = value
	return j
}

// SELECT * FROM users INNER JOIN JSON_EACH(`comments`) ON JSON_EXTRACT(JSON_EACH.value, '$.content') = 'aaa';
//
// SELECT * FROM users INNER JOIN JSON_EACH(`following`) ON JSON_EACH.value = 'OY';
func (j *jsonQueryExpression) Contains(op dao.JSONOp, value interface{}, keys ...string) dao.JSONQuery {
	if j.tx != nil {
		j.tx.Statement.Join(fmt.Sprintf("INNER JOIN JSON_EACH(%s)", j.tx.Statement.Quote(j.column)))
	}
	j.hasKeys = true
	j.op = op
	j.keys = keys
	j.equalsValue = value
	return j
}

func (j *jsonQueryExpression) Build(builder clause.Builder) {
	if stmt, ok := builder.(*dao.Statement); ok {
		if j.hasKeys {
			if len(j.keys) == 0 {
				builder.WriteString(fmt.Sprintf("JSON_EACH.value %s ", j.op.String()))
			} else {
				builder.WriteString(fmt.Sprintf("JSON_EXTRACT(JSON_EACH.value, '$.%s') %s ", strings.Join(j.keys, "."), j.op.String()))
			}
			stmt.AddVar(builder, j.equalsValue)
		} else {
			if len(j.keys) > 0 {
				builder.WriteString(fmt.Sprintf("JSON_EXTRACT(%s, '$.%s') %s ", stmt.Quote(j.column), strings.Join(j.keys, "."), j.op.String()))
				if j.op != dao.JSONHasKey {
					stmt.AddVar(builder, j.equalsValue)
				}
			}
		}
	}
}
