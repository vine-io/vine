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

package util

// download from https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/master/protoc-gen-grpc-gateway/httprule/compile_test.go

import (
	"reflect"
	"testing"
)

const (
	operandFiller = 0
)

func TestCompile(t *testing.T) {
	for _, spec := range []struct {
		segs []segment
		verb string

		ops    []int
		pool   []string
		fields []string
	}{
		{},
		{
			segs: []segment{
				wildcard{},
			},
			ops: []int{int(OpPush), operandFiller},
		},
		{
			segs: []segment{
				deepWildcard{},
			},
			ops: []int{int(OpPushM), operandFiller},
		},
		{
			segs: []segment{
				literal("v1"),
			},
			ops:  []int{int(OpLitPush), 0},
			pool: []string{"v1"},
		},
		{
			segs: []segment{
				literal("v1"),
			},
			verb: "LOCK",
			ops:  []int{int(OpLitPush), 0},
			pool: []string{"v1"},
		},
		{
			segs: []segment{
				variable{
					path: "name.nested",
					segments: []segment{
						wildcard{},
					},
				},
			},
			ops: []int{
				int(OpPush), operandFiller,
				int(OpConcatN), 1,
				int(OpCapture), 0,
			},
			pool:   []string{"name.nested"},
			fields: []string{"name.nested"},
		},
		{
			segs: []segment{
				literal("obj"),
				variable{
					path: "name.nested",
					segments: []segment{
						literal("a"),
						wildcard{},
						literal("b"),
					},
				},
				variable{
					path: "obj",
					segments: []segment{
						deepWildcard{},
					},
				},
			},
			ops: []int{
				int(OpLitPush), 0,
				int(OpLitPush), 1,
				int(OpPush), operandFiller,
				int(OpLitPush), 2,
				int(OpConcatN), 3,
				int(OpCapture), 3,
				int(OpPushM), operandFiller,
				int(OpConcatN), 1,
				int(OpCapture), 0,
			},
			pool:   []string{"obj", "a", "b", "name.nested"},
			fields: []string{"name.nested", "obj"},
		},
	} {
		tmpl := template{
			segments: spec.segs,
			verb:     spec.verb,
		}
		compiled := tmpl.Compile()
		if got, want := compiled.Version, opcodeVersion; got != want {
			t.Errorf("tmpl.Compile().Version = %d; want %d; segs=%#v, verb=%q", got, want, spec.segs, spec.verb)
		}
		if got, want := compiled.OpCodes, spec.ops; !reflect.DeepEqual(got, want) {
			t.Errorf("tmpl.Compile().OpCodes = %v; want %v; segs=%#v, verb=%q", got, want, spec.segs, spec.verb)
		}
		if got, want := compiled.Pool, spec.pool; !reflect.DeepEqual(got, want) {
			t.Errorf("tmpl.Compile().Pool = %q; want %q; segs=%#v, verb=%q", got, want, spec.segs, spec.verb)
		}
		if got, want := compiled.Verb, spec.verb; got != want {
			t.Errorf("tmpl.Compile().Verb = %q; want %q; segs=%#v, verb=%q", got, want, spec.segs, spec.verb)
		}
		if got, want := compiled.Fields, spec.fields; !reflect.DeepEqual(got, want) {
			t.Errorf("tmpl.Compile().Fields = %q; want %q; segs=%#v, verb=%q", got, want, spec.segs, spec.verb)
		}
	}
}
