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

package config

import (
	"testing"
)

func Test(t *testing.T) {
	tt := []struct {
		name   string
		values map[string]string
	}{
		{
			name: "No values",
		},
		{
			name: "Single value",
			values: map[string]string{
				"foo": "bar",
			},
		},
		{
			name: "Multiple values",
			values: map[string]string{
				"foo":   "bar",
				"apple": "tree",
			},
		},
	}

	//fp, err := filePath()
	//if err != nil {
	//	t.Fatal(err)
	//}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			//if _, err := os.Stat(fp); err != os.ErrNotExist {
			//	os.Remove(fp)
			//}

			for k, v := range tc.values {
				Set(v, k)
			}

			for k, v := range tc.values {
				val := Get(k)

				if v != val {
					t.Errorf("Got '%v' but expected '%v'", val, v)
				}
			}
		})
	}
}
