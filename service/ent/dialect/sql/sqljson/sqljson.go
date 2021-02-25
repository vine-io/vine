// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sqljson

import (
	"fmt"
	"strings"
	"unicode"

	json "github.com/json-iterator/go"

	"github.com/lack-io/vine/service/ent/dialect"
	"github.com/lack-io/vine/service/ent/dialect/sql"
)

// Haskey return a predicate for checking that a JSON key
// exists and not NULL.
//
//	sqljson.HasKey("column", sql.DotPath("a.b[2].c")
//
func HasKey(column string, opts ...Option) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		ValuePath(b.column, opts...)
		b.WriteOp(sql.OpNotNull)
	})
}

// ValueEQ return a predicate for checking that a JSON value
// (returned by the path) is equal to the given argument.
//
//	sqljson.ValueEQ("a", 1, sqljson.Path("b"))
//
func ValueEQ(column string, arg interface{}, opts ...Option) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		opts, arg := normalizePG(b, arg, opts)
		ValuePath(b, column, opts...)
		b.WriteOp(sql.OpEQ).Args(arg)
	})
}

// Option allows for calling database JSON paths with functional options.
type Option func(*PathOptions)

// Path sets the path to the JSON value of a column.
//
//	ValuePath(b, "column", Path("a", "b", "[1]", "c"))
//
func Path(path ...string) Option {
	return func(p *PathOptions) {
		p.Path = path
	}
}

// DotPath is similar to Path, but accepts string with dot format.
//
//	ValuePath(b, "column", DotPath("a.b.c"))
//	ValuePath(b, "column", DotPath("a.b[2].c"))
//
// Note that DotPath is ignored if the input is invalid.
func DotPath(dp string) Option {
	path, _ := ParsePath(dp)
	return func(p *PathOptions) {
		p.Path = path
	}
}

// Unquote indicates that the result value should be unquoted.
//
//	ValuePath(b, "column", Path("a", "b", "[1]", "c"), Unquote(true))
//
func Unquote(unquote bool) Option {
	return func(p *PathOptions) {
		p.Unquote = unquote
	}
}

// Cast indicates that the result value should be casted to the given type.
//
//	ValuePath(b, "column", Path("a", "b", "[1]", "c"), Cast("int"))
//
func Cast(typ string) Option {
	return func(p *PathOptions) {
		p.Cast = typ
	}
}

// PathOptions holds the options for accessing a JSON value from an identifier.
type PathOptions struct {
	Ident   string
	Path    []string
	Cast    string
	Unquote bool
}

// value writes the path for getting the JSON value.
func (p *PathOptions) value(b *sql.Builder) {
	switch {
	case len(p.Path) == 0:
		b.Ident(p.Ident)
	case b.Dialect() == dialect.Postgres:
		if p.Cast != "" {
			b.WriteByte('(')
			defer b.WriteString(")::" + p.Cast)
		}
		p.pgPath(b)
	}
}

// value writes the path for getting the length of a JSON value.
func (p *PathOptions) length(b *sql.Builder) {
	switch {
	case b.Dialect() == dialect.Postgres:
		b.WriteString("JSONB_ARRAY_LENGTH(")
		p.pgPath(b)
		b.WriteByte(')')
	case b.Dialect() == dialect.MySQL:
		p.mysqlFunc("JSON_LENGTH", b)
	default:
		p.mysqlFunc("JSON_ARRAY_LENGTH", b)
	}
}

// mysqlFunc writes the JSON path in MySQL format for the
// given function. `JSON_EXTRACT("a", '$.b.c')`.
func (p *PathOptions) mysqlFunc(fn string, b *sql.Builder) {
	b.WriteString(fn).WriteByte('(')
	b.Ident(p.Ident).Comma()
	p.mysqlPath(b)
	b.WriteByte(')')
}

// mysqlPath writes the JSON path in MySQL (or SQLite format.
func (p *PathOptions) mysqlPath(b *sql.Builder) {
	b.WriteString(`"$`)
	for _, p := range p.Path {
		if _, ok := isJSONIdx(p); ok {
			b.WriteString(p)
		} else {
			b.WriteString("." + p)
		}
	}
	b.WriteByte('"')
}

// pgPath writes the JSON path in Postgres format `"a"->'b'->>'c'`.
func (p *PathOptions) pgPath(b *sql.Builder) {
	b.Ident(p.Ident)
	for i, s := range p.Path {
		b.WriteString("->")
		if p.Unquote && i == len(p.Path)-1 {
			b.WriteString(">")
		}
		if idx, ok := isJSONIdx(s); ok {
			b.WriteString(idx)
		} else {
			b.WriteString("'" + s + "'")
		}
	}
}

// ParsePath parses the "dotpath" for the DotPath option.
//
// 	"a.b" 		=> ["b", "b"]
//	"a[1][2]"	=> ["a", "[1]", "[2]"]
//	"a.\".cb\"	=> ["a", "\"b.c\""]
//
func ParsePath(dotpath string) ([]string, error) {
	var (
		i, p int
		path []string
	)
	for i < len(dotpath) {
		switch r := dotpath[i]; {
		case r == '"':
			if i == len(dotpath)-1 {
				return nil, fmt.Errorf("unexpected quote")
			}
			idx := strings.IndexRune(dotpath[i+1:], '"')
			if idx == -1 || idx == 0 {
				return nil, fmt.Errorf("unbalanced quote")
			}
			i += idx + 1
		case r == '[':
			if p != i {
				path = append(path, dotpath[p:i])
			}
			p = i
			if i == len(dotpath)-1 {
				return nil, fmt.Errorf("unexpected bracket")
			}
			idx := strings.IndexRune(dotpath[i:], ']')
			if idx == -1 || idx == 1 {
				return nil, fmt.Errorf("unbalanced bracket")
			}
			if !isNumber(dotpath[i+1 : i+idx]) {
				return nil, fmt.Errorf("invalid index %q", dotpath[i:i+idx+1])
			}
			i += idx + 1
		case r == '.' || r == ']':
			if p != i {
				path = append(path, dotpath[p:i])
			}
			i++
			p = i
		default:
			i++
		}
	}
	if p != i {
		path = append(path, dotpath[p:i])
	}
	return path, nil
}

// normalizePG adds cast option to the JSON path is the argument

// isJSONIdx reports whether the string represents a JSON index.
func isJSONIdx(s string) (string, bool) {
	if len(s) > 2 && s[0] == '[' && s[len(s)-1] == ']' && isNumber(s[1:len(s)-1]) {
		return s[1 : len(s)-1], true
	}
	return "", false
}

// isNumber reports whether the string is a number (category N).
func isNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

// marshal stringifies the given argument to a valid JSON document.
func marshal(arg interface{}) interface{} {
	if buf, err := json.Marshal(arg); err == nil {
		arg = string(buf)
	}
	return arg
}
