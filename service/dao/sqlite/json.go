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
