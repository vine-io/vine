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

package json

import (
	"reflect"
	"testing"

	"github.com/vine-io/vine/lib/config/source"
)

func TestValues(t *testing.T) {
	emptyStr := ""
	testData := []struct {
		csdata   []byte
		path     []string
		accepter interface{}
		value    interface{}
	}{
		{
			[]byte(`{"foo": "bar", "baz": {"bar": "cat"}}`),
			[]string{"foo"},
			emptyStr,
			"bar",
		},
		{
			[]byte(`{"foo": "bar", "baz": {"bar": "cat"}}`),
			[]string{"baz", "bar"},
			emptyStr,
			"cat",
		},
	}

	for idx, test := range testData {
		values, err := newValues(&source.ChangeSet{
			Data: test.csdata,
		})
		if err != nil {
			t.Fatal(err)
		}

		err = values.Get(test.path...).Scan(&test.accepter)
		if err != nil {
			t.Fatal(err)
		}
		if test.accepter != test.value {
			t.Fatalf("No.%d Expected %v got %v for path %v", idx, test.value, test.accepter, test.path)
		}
	}
}

func TestStructArray(t *testing.T) {
	type T struct {
		Foo string
	}

	emptyTSlice := []T{}

	testData := []struct {
		csdata   []byte
		accepter []T
		value    []T
	}{
		{
			[]byte(`[{"foo": "bar"}]`),
			emptyTSlice,
			[]T{{Foo: "bar"}},
		},
	}

	for idx, test := range testData {
		values, err := newValues(&source.ChangeSet{
			Data: test.csdata,
		})
		if err != nil {
			t.Fatal(err)
		}

		err = values.Get().Scan(&test.accepter)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(test.accepter, test.value) {
			t.Fatalf("No.%d Expected %v got %v", idx, test.value, test.accepter)
		}
	}
}
