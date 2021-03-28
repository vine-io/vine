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

package gogo

import (
	"fmt"
	"strings"

	"github.com/lack-io/vine/cmd/generator"
)

var TagString = "gen"

const (
	// message tag
	_inline = "inline"
)

type Tag struct {
	Key   string
	Value string
}

// Paths for packages used by code generated in this file,
// relative to the import_prefix of the generator.Generator.
const (
)

// validator is an implementation of the Go protocol buffer compiler's
// plugin architecture. It generates bindings for validator support.
type gogo struct {
	*generator.Generator
	generator.PluginImports
	atleastOne  bool
	localName   string
	typesPkg    generator.Single
	bitsPkg     generator.Single
	errorsPkg   generator.Single
	protoPkg    generator.Single
	sortKeysPkg generator.Single
	mathPkg     generator.Single
	binaryPkg   generator.Single
	ioPkg      generator.Single
}

func New() *gogo {
	return &gogo{}
}

// Name returns the name of this plugin, "validator".
func (g *gogo) Name() string {
	return "gogo"
}

// Init initializes the plugin.
func (g *gogo) Init(gen *generator.Generator) {
	g.Generator = gen
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (g *gogo) objectNamed(name string) generator.Object {
	g.RecordTypeUse(name)
	return g.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (g *gogo) typeName(str string) string {
	return g.TypeName(g.objectNamed(str))
}

// P forwards to g.gen.P.
//func (g *gogo) P(args ...interface{}) { g.P(args...) }

// Generate generates code for the services in the given file.
func (g *gogo) Generate(file *generator.FileDescriptor) {
	if len(file.Comments()) == 0 {
		return
	}

	g.GenerateFileDescriptor(file)
	g.GenerateSize(file)
	g.GenerateMarshal(file)
	g.GenerateUnmarshal(file)
	g.GenerateGRPC(file)
}

// GenerateImports generates the import declaration for this file.
func (g *gogo) GenerateImports(file *generator.FileDescriptor, imports map[generator.GoImportPath]generator.GoPackageName) {
	if len(file.Comments()) == 0 {
		return
	}
}

func (g *gogo) extractTags(comments []*generator.Comment) map[string]*Tag {
	if comments == nil || len(comments) == 0 {
		return nil
	}
	tags := make(map[string]*Tag, 0)
	for _, c := range comments {
		if c.Tag != TagString || len(c.Text) == 0 {
			continue
		}
		parts := strings.Split(c.Text, ";")
		for _, p := range parts {
			tag := new(Tag)
			p = strings.TrimSpace(p)
			if i := strings.Index(p, "="); i > 0 {
				tag.Key = strings.TrimSpace(p[:i])
				v := strings.TrimSpace(p[i+1:])
				if v == "" {
					g.Fail(fmt.Sprintf("tag '%s' missing value", tag.Key))
				}
				tag.Value = v
			} else {
				tag.Key = p
			}
			tags[tag.Key] = tag
		}
	}

	return tags
}
