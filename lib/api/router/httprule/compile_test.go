// MIT License
//
// Copyright (c) 2020 The vine Authors
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

package httprule

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
