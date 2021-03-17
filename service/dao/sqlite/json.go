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

type jsonOp int32

const (
	eq jsonOp = iota
	hasKeys
)

type jsonQueryExpression struct {
	op          jsonOp
	column      string
	keys        []string
	equalsValue interface{}
}

func JSONQuery(column string) *jsonQueryExpression {
	return &jsonQueryExpression{column: column}
}

func (j *jsonQueryExpression) HasKeys(keys ...string) dao.JSONQuery {
	j.op = hasKeys
	j.keys = keys
	return j
}

func (j *jsonQueryExpression) Equals(value interface{}, keys ...string) dao.JSONQuery {
	j.op = eq
	j.keys = keys
	j.equalsValue = value
	return j
}

func (j *jsonQueryExpression) Build(builder clause.Builder) {
	if stmt, ok := builder.(*dao.Statement); ok {
		switch j.op {
		case hasKeys:
			if len(j.keys) > 0 {
				builder.WriteString(fmt.Sprintf("JSON_EXTRACT(%s, '$.%s') IS NOT NULL", stmt.Quote(j.column), strings.Join(j.keys, ".")))
			}
		case eq:
			if len(j.keys) > 0 {
				builder.WriteString(fmt.Sprintf("JSON_EXTRACT(%s, '$.%s') = ", stmt.Quote(j.column), strings.Join(j.keys, ".")))
				stmt.AddVar(builder, j.equalsValue)
			}
		}
	}
}
