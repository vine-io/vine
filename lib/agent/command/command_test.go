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
