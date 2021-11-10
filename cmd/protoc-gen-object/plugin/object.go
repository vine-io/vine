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

package plugin

import (
	"fmt"
	"strings"

	"github.com/vine-io/vine/cmd/generator"
)

var TagString = "gen"

const (
	// message tag
	_object = "object"
)

type Tag struct {
	Key   string
	Value string
}

// object is an implementation of the Go protocol buffer compiler's
// plugin architecture. It generates bindings for deepcopy support.
type object struct {
	generator.PluginImports
	gen        *generator.Generator
	schemaPkg  generator.Single
	runtimePkg generator.Single
}

func New() *object {
	return &object{}
}

// Name returns the name of this plugin, "object".
func (g *object) Name() string {
	return "object"
}

// Init initializes the plugin.
func (g *object) Init(gen *generator.Generator) {
	g.gen = gen
	g.PluginImports = generator.NewPluginImports(g.gen)
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (g *object) objectNamed(name string) generator.Object {
	g.gen.RecordTypeUse(name)
	return g.gen.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (g *object) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

// P forwards to g.gen.P()
func (g *object) P(args ...interface{}) { g.gen.P(args...) }

// Generate generates code for the services in the given file.
func (g *object) Generate(file *generator.FileDescriptor) {
	if len(file.Comments()) == 0 {
		return
	}

	g.schemaPkg = g.NewImport("github.com/vine-io/apimachinery/schema", "schema")
	g.runtimePkg = g.NewImport("github.com/vine-io/apimachinery/runtime", "runtime")

	var goOptionPkg string
	var version string
	goPkg := file.GoPackageName()

	if index := strings.LastIndex(file.Options.GetGoPackage(), ";"); index > 1 {
		goOptionPkg = file.Options.GetGoPackage()[index+1:]
	}
	if goOptionPkg == "" || goPkg == goOptionPkg {
		version = "v1"
	} else if strings.HasPrefix(goOptionPkg, goPkg) {
		version = strings.TrimPrefix(goOptionPkg, goPkg)
	}

	goPkg = strings.TrimSuffix(goPkg, version)
	if goPkg == "core" {
		goPkg = ""
	}


	g.P("// GroupName is the group name for this API")
	g.P(fmt.Sprintf(`const GroupName = "%s"`, goPkg))
	g.P()
	g.P("// SchemeGroupVersion is group version used to register these objects")
	g.P(fmt.Sprintf(`var SchemeGroupVersion = %s.GroupVersion{Group: GroupName, Version: "%s"}`, g.schemaPkg.Use(), version))
	g.P()

	g.P("var (")
	g.P(fmt.Sprintf("SchemaBuilder = %s.NewSchemeBuilder(addKnownTypes)", g.runtimePkg.Use()))
	g.P(fmt.Sprintf(`AddToScheme = SchemaBuilder.AddToScheme`))
	g.P(")")
	g.P()
	g.P(fmt.Sprintf(`func addKnownTypes(scheme %s.Scheme) error {`, g.runtimePkg.Use()))
	g.P(`if err := scheme.AddKnownTypes(SchemeGroupVersion,`)
	for _, msg := range file.Messages() {
		g.generateSchemaTypes(file, msg)
	}
	g.P(`); err != nil {`)
	g.P(`return err`)
	g.P("}")
	g.P()
	g.P("return nil")
	g.P("}")
	g.P()

	for _, msg := range file.Messages() {
		g.generateMessage(file, msg)
	}
}

func (g *object) generateSchemaTypes(file *generator.FileDescriptor, msg *generator.MessageDescriptor) {
	if msg.Proto.File() != file {
		return
	}
	if msg.Proto.Options != nil && msg.Proto.Options.GetMapEntry() {
		return
	}

	mname := msg.Proto.GetName()
	tags := g.extractTags(msg.Comments)
	_, ok := tags[_object]
	if !ok {
		return
	}
	g.P(fmt.Sprintf("&%s{},", mname))
}

func (g *object) generateMessage(file *generator.FileDescriptor, msg *generator.MessageDescriptor) {
	if msg.Proto.File() != file {
		return
	}
	if msg.Proto.Options != nil && msg.Proto.Options.GetMapEntry() {
		return
	}

	mname := msg.Proto.GetName()
	tags := g.extractTags(msg.Comments)
	_, ok := tags[_object]
	if !ok {
		return
	}

	pkg := g.NewImport("github.com/vine-io/apimachinery/runtime", "runtime")
	g.P("// assert to App implement to runtime.Object")
	g.P(fmt.Sprintf(`var _ %s.Object = (*%s)(nil)`, pkg.Use(), mname))
	g.P()
	g.P(fmt.Sprintf(`// DeepFrom is an auto-generated deepcopy function, copying value from %s.`, mname))
	g.P(fmt.Sprintf(`func (in *%s) DeepFrom(src %s.Object) {`, mname, pkg.Use()))
	g.P(fmt.Sprintf(`o := src.(*%s)`, mname))
	g.P(`o.DeepCopyInto(in)`)
	g.P(`}`)
	g.P(fmt.Sprintf(`// DeepCopy is an auto-generated deepcopy function, copying the receiver, creating a new %s.`, mname))
	g.P(fmt.Sprintf(`func (in *%s) DeepCopy() %s.Object {`, mname, pkg.Use()))
	g.P(`if in == nil {`)
	g.P(`return nil`)
	g.P(`}`)
	g.P(fmt.Sprintf(`out := new(%s)`, mname))
	g.P(`in.DeepCopyInto(out)`)
	g.P(`return out`)
	g.P(`}`)
	g.P()
}

func (g *object) extractTags(comments []*generator.Comment) map[string]*Tag {
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
