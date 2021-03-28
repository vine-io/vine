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

package plugin

import (
	"fmt"
	"strings"

	"github.com/lack-io/vine/cmd/generator"
)

var TagString = "gen"

const (
	// message tag
	_ignore = "ignore"
)

type Tag struct {
	Key   string
	Value string
}

// Paths for packages used by code generated in this file,
// relative to the import_prefix of the generator.Generator.
const (
)

// gogo is an implementation of the Go protocol buffer compiler's
// plugin architecture. It generates bindings for validator support.
type gogo struct {
	gen *generator.Generator
}

func New() *gogo {
	return &gogo{}
}

// Name returns the name of this plugin, "gogo".
func (g *gogo) Name() string {
	return "gogo"
}

// The names for packages imported in the generated code.
// They may vary from the final path component of the import path
// if the name is used by other packages.
var (
	isPkg      string
	stringsPkg string
	pkgImports map[generator.GoPackageName]bool
)

// Init initializes the plugin.
func (g *gogo) Init(gen *generator.Generator) {
	g.gen = gen
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (g *gogo) objectNamed(name string) generator.Object {
	g.gen.RecordTypeUse(name)
	return g.gen.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (g *gogo) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

// P forwards to g.gen.P.
func (g *gogo) P(args ...interface{}) { g.gen.P(args...) }

// Generate generates code for the services in the given file.
func (g *gogo) Generate(file *generator.FileDescriptor) {
	if len(file.Comments()) == 0 {
		return
	}

	//isPkg = string(g.gen.AddImport(isPkgPath))
	//stringsPkg = string(g.gen.AddImport(stringsPkgPath))

	g.P("// Reference imports to suppress errors if they are not otherwise used.")
	g.P("var _ ", isPkg, ".Empty")
	g.P("var _ ", stringsPkg, ".Builder")
	g.P()
	for _, msg := range file.Messages() {
		g.generateMessage(file, msg)
	}
}

// GenerateImports generates the import declaration for this file.
func (g *gogo) GenerateImports(file *generator.FileDescriptor, imports map[generator.GoImportPath]generator.GoPackageName) {
	if len(file.Comments()) == 0 {
		return
	}

	// We need to keep track of imported packages to make sure we don't produce
	// a name collision when generating types.
	pkgImports = make(map[generator.GoPackageName]bool)
	for _, name := range imports {
		pkgImports[name] = true
	}
}

func (g *gogo) generateMessage(file *generator.FileDescriptor, msg *generator.MessageDescriptor) {
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
					g.gen.Fail(fmt.Sprintf("tag '%s' missing value", tag.Key))
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
