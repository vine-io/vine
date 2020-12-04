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

package file

import (
	"testing"

	"github.com/lack-io/vine/config/source"
)

func TestFormat(t *testing.T) {
	opts := source.NewOptions()
	e := opts.Encoder

	testCases := []struct {
		p string
		f string
	}{
		{"/foo/bar.json", "json"},
		{"/foo/bar.yaml", "yaml"},
		{"/foo/bar.xml", "xml"},
		{"/foo/bar.conf.ini", "ini"},
		{"conf", e.String()},
	}

	for _, d := range testCases {
		f := format(d.p, e)
		if f != d.f {
			t.Fatalf("%s: expected %s got %s", d.p, d.f, f)
		}
	}

}
