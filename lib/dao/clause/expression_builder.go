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

import "strings"

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

func ParseOp(v interface{}) (op EOp) {
	switch v.(type) {
	case string:
		vv := v.(string)
		if strings.HasPrefix(vv, "%") || strings.HasSuffix(vv, "%") {
			op = LikeOp
		} else {
			op = EqOp
		}
	default:
		op = EqOp
	}
	return
}

// condExpr the builder for interface Expression
// condExpr
type condExpr struct {
	op EOp
}

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
