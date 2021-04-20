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

package schema

import (
	"sort"
	"strconv"
	"strings"
)

type Index struct {
	Name    string
	Class   string // UNIQUE | FULLTEXT | SPATIAL
	Type    string // btree, hash, gist, spgist, gin, and brin
	Where   string
	Comment string
	Option  string // WITH PARSER parser_name
	Fields  []IndexOption
}

type IndexOption struct {
	*Field
	Expression string
	Sort       string // DESC, ASC
	Collate    string
	Length     int
	priority   int
}

// ParseIndexes parse schema indexes
func (schema *Schema) ParseIndexes() map[string]Index {
	var indexes = map[string]Index{}

	for _, field := range schema.Fields {
		if field.TagSettings["INDEX"] != "" || field.TagSettings["UNIQUEINDEX"] != "" {
			for _, index := range parseFieldIndexes(field) {
				idx := indexes[index.Name]
				idx.Name = index.Name
				if idx.Class == "" {
					idx.Class = index.Class
				}
				if idx.Type == "" {
					idx.Type = index.Type
				}
				if idx.Where == "" {
					idx.Where = index.Where
				}
				if idx.Comment == "" {
					idx.Comment = index.Comment
				}
				if idx.Option == "" {
					idx.Option = index.Option
				}

				idx.Fields = append(idx.Fields, index.Fields...)
				sort.Slice(idx.Fields, func(i, j int) bool {
					return idx.Fields[i].priority < idx.Fields[j].priority
				})

				indexes[index.Name] = idx
			}
		}
	}

	return indexes
}

func (schema *Schema) LookIndex(name string) *Index {
	if schema != nil {
		indexes := schema.ParseIndexes()
		for _, index := range indexes {
			if index.Name == name {
				return &index
			}

			for _, field := range index.Fields {
				if field.Name == name {
					return &index
				}
			}
		}
	}

	return nil
}

func parseFieldIndexes(field *Field) (indexes []Index) {
	for _, value := range strings.Split(field.Tag.Get("dao"), ";") {
		if value != "" {
			v := strings.Split(value, ":")
			k := strings.TrimSpace(strings.ToUpper(v[0]))
			if k == "INDEX" || k == "UNIQUEINDEX" {
				var (
					name      string
					tag       = strings.Join(v[1:], ":")
					idx       = strings.Index(tag, ",")
					settings  = ParseTagSetting(tag, ",")
					length, _ = strconv.Atoi(settings["LENGTH"])
				)

				if idx == -1 {
					idx = len(tag)
				}

				if idx != -1 {
					name = tag[0:idx]
				}

				if name == "" {
					name = field.Schema.namer.IndexName(field.Schema.Table, field.Name)
				}

				if (k == "UNIQUEINDEX") || settings["UNIQUE"] != "" {
					settings["CLASS"] = "UNIQUE"
				}

				priority, err := strconv.Atoi(settings["PRIORITY"])
				if err != nil {
					priority = 10
				}

				indexes = append(indexes, Index{
					Name:    name,
					Class:   settings["CLASS"],
					Type:    settings["TYPE"],
					Where:   settings["WHERE"],
					Comment: settings["COMMENT"],
					Option:  settings["OPTION"],
					Fields: []IndexOption{{
						Field:      field,
						Expression: settings["EXPRESSION"],
						Sort:       settings["SORT"],
						Collate:    settings["COLLATE"],
						Length:     length,
						priority:   priority,
					}},
				})
			}
		}
	}

	return
}
