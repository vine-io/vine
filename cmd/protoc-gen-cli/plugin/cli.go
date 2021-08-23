// MIT License
//
// Copyright (c) 2021 Lack
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
	"bytes"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/vine-io/vine/cmd/generator"
)

var TagString = "gen"

const (
	// Service
	_cli = "cli"

	_inline = "inline"
)

type Tag struct {
	Key   string
	Value string
}

// cli is an implementation of the Go protocol buffer compiler's
// plugin architecture. It generates bindings for cli support.
type cli struct {
	generator.PluginImports
	gen       *generator.Generator
	sourcePkg string
	traits    []Trait

	ctxPkg    generator.Single
	clientPkg generator.Single
}

func New() *cli {
	return &cli{}
}

// Name returns the name of this plugin, "cli".
func (g *cli) Name() string {
	return "cli"
}

// Init initializes the plugin.
func (g *cli) Init(gen *generator.Generator) {
	g.gen = gen
	g.PluginImports = generator.NewPluginImports(g.gen)
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (g *cli) objectNamed(name string) generator.Object {
	g.gen.RecordTypeUse(name)
	return g.gen.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (g *cli) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

// P forwards to g.gen.P.
func (g *cli) P(args ...interface{}) { g.gen.P(args...) }

// Generate generates code for the services in the given file.
func (g *cli) Generate(file *generator.FileDescriptor) {
	if len(file.Comments()) == 0 {
		return
	}

	g.ctxPkg = g.NewImport("context", "context")
	g.clientPkg = g.NewImport("github.com/vine-io/vine/core/client", "vclient")

	if g.gen.OutPut.Load {
		g.sourcePkg = string(g.gen.AddImport(generator.GoImportPath(g.gen.OutPut.SourcePkgPath)))
	}

	g.traits = make([]Trait, 0)
	for _, service := range file.TagServices() {
		g.generateImpl(file, service)
	}

	for _, trait := range g.traits {
		g.generateTrait(file, trait)
	}
}

func (g *cli) generateTrait(file *generator.FileDescriptor, trait Trait) {
	g.P(fmt.Sprintf(`type Client interface {`))
	for _, meth := range trait.methods {
		buf := bytes.NewBuffer([]byte(""))
		//g.P(fmt.Sprintf(`%s()`))
		buf.WriteString(fmt.Sprintf(`%s(`, meth.name))
		args := make([]string, 0)
		for _, arg := range meth.args {
			args = append(args, arg.name)
		}
		buf.WriteString(strings.Join(args, ", "))
		buf.WriteString(") ")
		if len(meth.results) == 1 {
			buf.WriteString(meth.results[0].name)
		} else {
			buf.WriteByte('(')
			results := make([]string, 0)
			for _, result := range meth.results {
				results = append(results, result.name)
			}
			buf.WriteString(strings.Join(results, ", "))
			buf.WriteByte(')')
		}
		g.P(buf.String())
	}
	g.P(`}`)
	g.P()

	g.P(`type SimpleClient struct {`)
	sname := trait.name
	if g.gen.OutPut.Load {
		sname = fmt.Sprintf(`%s.%s`, g.sourcePkg, trait.name)
	}
	g.P(fmt.Sprintf(`cc %sService`, sname))
	g.P(`}`)
	g.P()

	//for _, meth := range trait.methods {
	//
	//}
}

// generateService generates all the code for the named service.
func (g *cli) generateImpl(file *generator.FileDescriptor, service *generator.ServiceDescriptor) {
	svcTags := g.extractTags(service.Comments)
	if _, ok := svcTags[_cli]; !ok {
		return
	}

	trait := Trait{
		name:    service.Proto.GetName(),
		methods: make([]Method, 0),
	}

	for _, item := range service.Methods {
		method := Method{
			name:         item.Proto.GetName(),
			args:         []Value{},
			results:      []Value{},
			clientStream: item.Proto.GetClientStreaming(),
			serverStream: item.Proto.GetServerStreaming(),
		}

		ctxArg := Value{
			name:  fmt.Sprintf(`%s.Context`, g.ctxPkg.Use()),
			alias: "ctx",
		}
		method.args = append(method.args, ctxArg)

		input := g.extractMessage(item.Proto.GetInputType())
		if input == nil {
			g.gen.Fail(item.Proto.GetInputType(), "not found")
			return
		}
		fields := make([]*generator.FieldDescriptor, 0)
		g.buildField(input, &fields)
		for _, field := range fields {
			method.args = append(method.args, g.parseField(file, input, field))
		}
		method.args = append(method.args, Value{
			name:  fmt.Sprintf(`...%s.CallOption`, g.clientPkg.Use()),
			alias: "opts",
		})

		output := g.extractMessage(item.Proto.GetOutputType())
		if output == nil {
			g.gen.Fail(item.Proto.GetOutputType(), "not found")
			return
		}
		fields = make([]*generator.FieldDescriptor, 0)
		g.buildField(output, &fields)
		for _, field := range fields {
			method.results = append(method.results, g.parseField(file, output, field))
		}

		errResult := Value{
			name:  "error",
			alias: "err",
		}
		method.results = append(method.results, errResult)
		trait.methods = append(trait.methods, method)
	}

	g.traits = append(g.traits, trait)

	return
}

//func (g *cli) wrapImplements(file *generator.FileDescriptor, service )

func (g *cli) buildField(msg *generator.MessageDescriptor, out *[]*generator.FieldDescriptor) {
	for _, field := range msg.Fields {
		if !field.Proto.IsMessage() {
			*out = append(*out, field)
			//continue
		}
		if field.Proto.IsMessage() {
			_, isInline := g.extractTags(field.Comments)[_inline]
			if isInline {
				subMsg := g.gen.ExtractMessage(field.Proto.GetTypeName())
				g.buildField(subMsg, out)
			} else {
				*out = append(*out, field)
			}
		}
	}
}

func (g *cli) parseField(file *generator.FileDescriptor, msg *generator.MessageDescriptor, field *generator.FieldDescriptor) Value {
	value := Value{alias: field.Proto.GetJsonName()}

	if field.Proto.IsRepeated() {
		if strings.HasSuffix(field.Proto.GetTypeName(), "Entry") {
			for _, nest := range msg.Proto.GetNestedType() {
				if strings.HasSuffix(field.Proto.GetTypeName(), nest.GetName()) {
					key, v := nest.GetMapFields()
					value.name = fmt.Sprintf(`map[%s]%s`, g.getType(file, msg, key), g.getType(file, msg, v))
				}
			}
		} else {
			value.name = "[]" + g.getType(file, msg, field.Proto)
		}
	} else {
		value.name = g.getType(file, msg, field.Proto)
	}

	return value
}

func (g *cli) getType(file *generator.FileDescriptor, msg *generator.MessageDescriptor, fd *descriptor.FieldDescriptorProto) string {
	switch fd.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return "float64"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return "float32"
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		return "uint64"
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		return "uint32"
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_INT64:
		return "int64"
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_ENUM:
		return "int32"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return "[]byte"
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		var typ string
		subMsg := g.extractMessage(fd.GetTypeName())
		if subMsg.Proto.File() == file {
			if g.gen.OutPut.Load {
				typ = g.sourcePkg + "." + subMsg.Proto.GetName()
			}
			typ = subMsg.Proto.GetName()
		} else {
			obj := g.gen.ObjectNamed(fd.GetTypeName())
			v, ok := g.gen.ImportMap[obj.GoImportPath().String()]
			if !ok {
				v = string(g.gen.AddImport(obj.GoImportPath()))
			}
			typ = fmt.Sprintf("*%s.%s", v, subMsg.Proto.GetName())
		}
		return typ
	}

	return ""
}

// extractMessage extract MessageDescriptor by name
func (g *cli) extractMessage(name string) *generator.MessageDescriptor {
	obj := g.gen.ObjectNamed(name)
	for _, f := range g.gen.AllFiles() {
		for _, m := range f.Messages() {
			if m.Proto.GoImportPath() == obj.GoImportPath() {
				for _, item := range obj.TypeName() {
					if item == m.Proto.GetName() {
						return m
					}
				}
			}
		}
	}
	return nil
}

// extractTags extracts the maps of *Tag from []*generator.Comment
func (g *cli) extractTags(comments []*generator.Comment) map[string]*Tag {
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

// extractDesc extracts descriptions from []*generator.Comment
func extractDesc(comments []*generator.Comment) []string {
	if comments == nil || len(comments) == 0 {
		return nil
	}
	desc := make([]string, 0)
	for _, c := range comments {
		if c.Tag == "" {
			text := strings.TrimSpace(c.Text)
			if len(text) == 0 {
				continue
			}
			desc = append(desc, text)
		}
	}
	return desc
}

func TrimString(s string, c string) string {
	s = strings.TrimPrefix(s, c)
	s = strings.TrimSuffix(s, c)
	return s
}

// fullStringSlice converts [1,2,3] => `"1", "2", "3"`
func fullStringSlice(s string) string {
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	parts := strings.Split(s, ",")
	out := make([]string, 0)
	for _, a := range parts {
		a = strings.TrimSpace(a)
		if len(a) == 0 {
			continue
		}
		if !strings.HasPrefix(a, "\"") {
			a = "\"" + a
		}
		if !strings.HasSuffix(a, "\"") {
			a = a + "\""
		}
		out = append(out, a)
	}
	return strings.Join(out, ",")
}
