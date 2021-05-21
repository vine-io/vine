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

package gogo

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/gogoproto"

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

// generatedCodeVersion indicates a version of the generated code.
// It is incremented whenever an incompatibility between the generated code and
// the grpc package is introduced; the generated code references
// a constant, grpc.SupportPackageIsVersionN (where N is generatedCodeVersion).
const generatedCodeVersion = 4

// validator is an implementation of the Go protocol buffer compiler's
// plugin architecture. It generates bindings for validator support.
type gogo struct {
	*generator.Generator
	generator.PluginImports
	atleastOne  bool
	localName   string
	contextPkg  generator.Single
	grpcPkg     generator.Single
	codePkg     generator.Single
	statusPkg   generator.Single
	fmtPkg      generator.Single
	typesPkg    generator.Single
	bitsPkg     generator.Single
	errorsPkg   generator.Single
	protoPkg    generator.Single
	sortKeysPkg generator.Single
	mathPkg     generator.Single
	binaryPkg   generator.Single
	ioPkg       generator.Single
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
	g.PluginImports = generator.NewPluginImports(g.Generator)
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
	g.contextPkg = g.NewImport("context", "context")
	g.grpcPkg = g.NewImport("google.golang.org/grpc", "grpc")
	g.codePkg = g.NewImport("google.golang.org/grpc/codes", "codes")
	g.statusPkg = g.NewImport("google.golang.org/grpc/status", "status")

	g.atleastOne = false
	g.localName = generator.FileName(file)

	g.mathPkg = g.NewImport("math", "")
	g.sortKeysPkg = g.NewImport("github.com/gogo/protobuf/sortkeys", "sortkeys")
	g.protoPkg = g.NewImport("github.com/gogo/protobuf/proto", "proto")
	if !gogoproto.ImportsGoGoProto(file.FileDescriptorProto) {
		g.protoPkg = g.NewImport("github.com/golang/protobuf/proto", "proto")
	}
	g.errorsPkg = g.NewImport("errors", "errors")
	g.binaryPkg = g.NewImport("encoding/binary", "ebinary")
	g.typesPkg = g.NewImport("github.com/gogo/protobuf/types", "types")

	g.ioPkg = g.NewImport("io", "io")
	g.bitsPkg = g.NewImport("math/bits", "bits")
	g.fmtPkg = g.NewImport("fmt", "fmt")

	g.P("var _ =", g.binaryPkg.Use(), ".BigEndian")
	g.P()

	// Assert version compatibility.
	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the grpc package it is being compiled against.")
	g.P("const _ = ", g.grpcPkg.Use(), ".SupportPackageIsVersion", generatedCodeVersion)
	g.P()

	g.GenerateFileDescriptor(file)
	g.GenerateSize(file)
	g.GenerateMarshal(file)
	g.GenerateUnmarshal(file)
	g.GenerateGRPC(file)
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
