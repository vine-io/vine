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

package command

import (
	"testing"
)

func TestCommand(t *testing.T) {
	c := &cmd{
		name:        "test",
		usage:       "test usage",
		description: "test description",
		exec: func(args ...string) ([]byte, error) {
			return []byte("test"), nil
		},
	}

	if c.String() != c.name {
		t.Fatalf("expected name %s got %s", c.name, c.String())
	}

	if c.Usage() != c.usage {
		t.Fatalf("expected usage %s got %s", c.usage, c.Usage())
	}

	if c.Description() != c.description {
		t.Fatalf("expected description %s got %s", c.description, c.Description())
	}

	if r, err := c.Exec(); err != nil {
		t.Fatal(err)
	} else if string(r) != "test" {
		t.Fatalf("expected exec result test got %s", string(r))
	}
}

func TestNewCommand(t *testing.T) {
	c := &cmd{
		name:        "test",
		usage:       "test usage",
		description: "test description",
		exec: func(args ...string) ([]byte, error) {
			return []byte("test"), nil
		},
	}

	nc := NewCommand(c.name, c.usage, c.description, c.exec)

	if nc.String() != c.name {
		t.Fatalf("expected name %s got %s", c.name, nc.String())
	}

	if nc.Usage() != c.usage {
		t.Fatalf("expected usage %s got %s", c.usage, nc.Usage())
	}

	if nc.Description() != c.description {
		t.Fatalf("expected description %s got %s", c.description, nc.Description())
	}

	if r, err := nc.Exec(); err != nil {
		t.Fatal(err)
	} else if string(r) != "test" {
		t.Fatalf("expected exec result test got %s", string(r))
	}
}
