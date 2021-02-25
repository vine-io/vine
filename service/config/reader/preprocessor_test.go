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

package reader

import (
	"os"
	"strings"
	"testing"
)

func TestReplaceEnvVars(t *testing.T) {
	os.Setenv("myBar", "cat")
	os.Setenv("MYBAR", "cat")
	os.Setenv("my_Bar", "cat")
	os.Setenv("myBar_", "cat")

	testData := []struct {
		expected string
		data     []byte
	}{
		// Right use cases
		{
			`{"foo": "bar", "baz": {"bar": "cat"}}`,
			[]byte(`{"foo": "bar", "baz": {"bar": "${myBar}"}}`),
		},
		{
			`{"foo": "bar", "baz": {"bar": "cat"}}`,
			[]byte(`{"foo": "bar", "baz": {"bar": "${MYBAR}"}}`),
		},
		{
			`{"foo": "bar", "baz": {"bar": "cat"}}`,
			[]byte(`{"foo": "bar", "baz": {"bar": "${my_Bar}"}}`),
		},
		{
			`{"foo": "bar", "baz": {"bar": "cat"}}`,
			[]byte(`{"foo": "bar", "baz": {"bar": "${myBar_}"}}`),
		},
		// Wrong use cases
		{
			`{"foo": "bar", "baz": {"bar": "${myBar-}"}}`,
			[]byte(`{"foo": "bar", "baz": {"bar": "${myBar-}"}}`),
		},
		{
			`{"foo": "bar", "baz": {"bar": "${}"}}`,
			[]byte(`{"foo": "bar", "baz": {"bar": "${}"}}`),
		},
		{
			`{"foo": "bar", "baz": {"bar": "$sss}"}}`,
			[]byte(`{"foo": "bar", "baz": {"bar": "$sss}"}}`),
		},
		{
			`{"foo": "bar", "baz": {"bar": "${sss"}}`,
			[]byte(`{"foo": "bar", "baz": {"bar": "${sss"}}`),
		},
		{
			`{"foo": "bar", "baz": {"bar": "{something}"}}`,
			[]byte(`{"foo": "bar", "baz": {"bar": "{something}"}}`),
		},
		// Use cases without replace env vars
		{
			`{"foo": "bar", "baz": {"bar": "cat"}}`,
			[]byte(`{"foo": "bar", "baz": {"bar": "cat"}}`),
		},
	}

	for _, test := range testData {
		res, err := ReplaceEnvVars(test.data)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Compare(test.expected, string(res)) != 0 {
			t.Fatalf("Expected %s got %s", test.expected, res)
		}
	}
}
