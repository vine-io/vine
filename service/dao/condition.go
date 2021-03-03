// Copyright 2021 lack
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either columness or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

type PredicateOp int32

const (
	OpFirst PredicateOp = iota
	OpLast
	OpOrder
	OpLimit

	OpOmit

	OpEQ        // ==
	OpNEQ       // <>
	OpGT        // >
	OpGTE       // >=
	OpLT        // <
	OpLTE       // <=
	OpNot       // not
	OpIn        // in
	OpLike      // like
	OpIsNull    // IS NULL
	OpIsNotNull // IS NOT NULL
)

type PP func(query interface{}, args ...interface{}) DB

type Predicate struct {
	Op     PredicateOp
	Column string
	Vars   []interface{}
	JSON   bool
}

type Predicates []*Predicate

type P func(*Predicates)

func First() P {
	return build(OpFirst, false, "", nil)
}

func Last() P {
	return build(OpLast, false, "", nil)
}

func OrderBy(column string, desc bool) P {
	c := column
	if desc {
		c += " DESC"
	}
	return build(OpOrder, false, column, nil)
}

func Limit(limit, offset int32) P {
	return build(OpLimit, false, "", limit, offset)
}

func EQ(column string, vars ...interface{}) P {
	return build(OpEQ, false, column, vars...)
}

func NEQ(column string, vars ...interface{}) P {
	return build(OpNEQ, false, column, vars...)
}

func GT(column string, vars ...interface{}) P {
	return build(OpGT, false, column, vars...)
}

func GTE(column string, vars ...interface{}) P {
	return build(OpGTE, false, column, vars...)
}

func LT(column string, vars ...interface{}) P {
	return build(OpLT, false, column, vars...)
}

func LTE(column string, vars ...interface{}) P {
	return build(OpLTE, false, column, vars...)
}

func Not(column string, vars ...interface{}) P {
	return build(OpNot, false, column, vars...)
}

func IN(column string, vars ...interface{}) P {
	return build(OpIn, false, column, vars...)
}

func Like(column string, vars ...interface{}) P {
	return build(OpLike, false, column, vars...)
}

func IsNull(column string, vars ...interface{}) P {
	return build(OpIsNull, false, column, vars...)
}

func IsNotNull(column string, vars ...interface{}) P {
	return build(OpIsNotNull, false, column, vars...)
}

func Omit(columns ...string) P {
	vars := make([]interface{}, len(columns))
	for i := range columns {
		vars[i] = columns[i]
	}
	return build(OpOmit, false, "", vars...)
}

func JSONQuery(op PredicateOp, column string, vars ...interface{}) P {
	return build(op, true, column, vars...)
}

func build(op PredicateOp, json bool, column string, vars ...interface{}) P {
	return func(ps *Predicates) {
		*ps = append(*ps, &Predicate{
			Op:   op,
			Column: column,
			Vars: vars,
			JSON: json,
		})
	}
}
