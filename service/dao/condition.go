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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

type PredicateOp int32

const (
	OpFirst PredicateOp = iota
	OpLast
	OpOrder
	OpLimit

	OpAnd   // AND
	OpOr    // OR
	OpEQ    // ==
	OpNEQ   // <>
	OpGT    // >
	OpGTE   // >=
	OpLT    // <
	OpLTE   // <=
	OpNot   // not
	OpIn    // in
	OpLike  // like
	OpKeyEQ // map key equal
)

type Predication struct {
	Op   PredicateOp
	Expr string
	Vars []interface{}
}

type P func(*Predication)

func First() P {
	return func(p *Predication) {
		p.Op = OpFirst
	}
}

func Last() P {
	return func(p *Predication) {
		p.Op = OpLast
	}
}

func OrderBy(column string, desc bool) P {
	return func(p *Predication) {
		p.Op = OpOrder
		p.Expr = column
		if desc {
			p.Expr += " DESC"
		}
	}
}

func Limit(limit, offset int32) P {
	return func(p *Predication) {
		p.Op = OpLimit
		p.Vars = []interface{}{limit, offset}
	}
}
