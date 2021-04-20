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

package flag

import (
	"errors"
	"flag"
	"strings"
	"time"

	"github.com/imdario/mergo"

	"github.com/lack-io/vine/lib/config/source"
)

type flagsrc struct {
	opts source.Options
}

func (fs *flagsrc) Read() (*source.ChangeSet, error) {
	if !flag.Parsed() {
		return nil, errors.New("flags not parsed")
	}

	var changes map[string]interface{}

	visitFn := func(f *flag.Flag) {
		n := strings.ToLower(f.Name)
		keys := strings.FieldsFunc(n, split)
		reverse(keys)

		tmp := make(map[string]interface{})
		for i, k := range keys {
			if i == 0 {
				tmp[k] = f.Value
				continue
			}

			tmp = map[string]interface{}{k: tmp}
		}

		mergo.Map(&changes, tmp) // need to sort error handling
		return
	}

	unset, ok := fs.opts.Context.Value(includeUnsetKey{}).(bool)
	if ok && unset {
		flag.VisitAll(visitFn)
	} else {
		flag.Visit(visitFn)
	}

	b, err := fs.opts.Encoder.Encode(changes)
	if err != nil {
		return nil, err
	}

	cs := &source.ChangeSet{
		Format:    fs.opts.Encoder.String(),
		Data:      b,
		Timestamp: time.Now(),
		Source:    fs.String(),
	}
	cs.CheckSum = cs.Sum()

	return cs, nil
}

func split(r rune) bool {
	return r == '-' || r == '_'
}

func reverse(ss []string) {
	for i := len(ss)/2 - 1; i >= 0; i-- {
		opp := len(ss) - 1 - i
		ss[i], ss[opp] = ss[opp], ss[i]
	}
}

func (fs *flagsrc) Watch() (source.Watcher, error) {
	return source.NewNoopWatcher()
}

func (fs *flagsrc) Write(cs *source.ChangeSet) error {
	return nil
}

func (fs *flagsrc) String() string {
	return "flag"
}

// NewSource returns a config source for integrating parsed flags.
// Hyphens are delimiters for nesting, and all keys are lowercased.
//
// Example:
//      dbhost := flag.String("database-host", "localhost", "the db host name")
//
//      {
//          "database": {
//              "host": "localhost"
//          }
//      }
func NewSource(opts ...source.Option) source.Source {
	return &flagsrc{opts: source.NewOptions(opts...)}
}
