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
	FirstOp PredicateOp = iota
	LastOp
	OrderOp
	LimitOp
)

type Predication struct {
	Op   PredicateOp
	Expr string
	Vars []interface{}
}

type P func(*Predication)

func First() P {
	return func(p *Predication) {
		p.Op = FirstOp
	}
}

func Last() P {
	return func(p *Predication) {
		p.Op = LastOp
	}
}

func OrderBy(column string, desc bool) P {
	return func(p *Predication) {
		p.Op = OrderOp
		p.Expr = column
		if desc {
			p.Expr += " DESC"
		}
	}
}

func Limit(limit, offset int32) P {
	return func(p *Predication) {
		p.Op = LimitOp
		p.Vars = []interface{}{limit, offset}
	}
}
