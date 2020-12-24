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

package json

import (
	"reflect"
	"testing"

	"github.com/lack-io/vine/service/config/source"
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
