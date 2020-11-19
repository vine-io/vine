// Copyright 2020 The vine Authors
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

package selection

type SelectorKind int32

const (
	Where SelectorKind = iota
	Or
)

type SelectorOp int32

const (
	Eq   = iota // equal
	Ne          // not equal
	Gt          // great than
	Gte         // great than and equal
	Lt          // less than
	Lte         // less than and equal
	Bw          // between ... and ..., and the value should be "$1,$2"
	Like        // like, support regular expression
	In          // in, is it the slice include the value of the field
	Not         // not
)

type Selector struct {
	Kind SelectorKind

	Field string

	Op SelectorOp

	Value interface{}
}
