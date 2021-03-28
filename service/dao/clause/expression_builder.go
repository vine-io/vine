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

type EOp int32

const (
	EqOp EOp = iota + 1
	NeqOp
	GtOp
	GteOp
	LtOp
	LteOp
	InOp
	LikeOp
)

// condExpr the builder for interface Expression
// condExpr
type condExpr struct {
	op EOp
}

// Raw().Op(EqOp).Set("id", 1).Build()
func Cond() *condExpr {
	return &condExpr{op: EqOp}
}

func (e *condExpr) Op(op EOp) *condExpr {
	e.op = op
	return e
}

func (e *condExpr) Build(c string, v interface{}, vv ...interface{}) Expression {
	switch e.op {
	case EqOp:
		return Eq{Column: Column{Name: c}, Value: v}
	case NeqOp:
		return Neq{Column: Column{Name: c}, Value: v}
	case GtOp:
		return Gt{Column: Column{Name: c}, Value: v}
	case GteOp:
		return Gte{Column: Column{Name: c}, Value: v}
	case LtOp:
		return Lt{Column: Column{Name: c}, Value: v}
	case LteOp:
		return Lte{Column: Column{Name: c}, Value: v}
	case InOp:
		return IN{Column: Column{Name: c}, Values: append([]interface{}{v}, vv...)}
	case LikeOp:
		return Like{Column: Column{Name: c}, Value: v}
	}
	return nil
}
