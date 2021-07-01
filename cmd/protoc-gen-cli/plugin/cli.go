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

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

	"github.com/lack-io/vine/cmd/generator"
)

var TagString = "gen"

const (
	// message tag
	_ignore = "ignore"

	// field common tag
	_inline   = "inline"
	_required = "required"
	_default  = "default"
	_in       = "in"
	_enum     = "enum"
	_notIn    = "not_in"

	// string tag
	_minLen   = "min_len"
	_maxLen   = "max_len"
	_prefix   = "prefix"
	_suffix   = "suffix"
	_contains = "contains"
	_number   = "number"
	_email    = "email"
	_ip       = "ip"
	_ipv4     = "ipv4"
	_ipv6     = "ipv6"
	_crontab  = "crontab"
	_uuid     = "uuid"
	_uri      = "uri"
	_domain   = "domain"
	_pattern  = "pattern"

	// int32, int64, uint32, uint64, float32, float64 tag
	_ne  = "ne"
	_eq  = "eq"
	_lt  = "lt"
	_lte = "lte"
	_gt  = "gt"
	_gte = "gte"

	// bytes tag
	_maxBytes = "max_bytes"
	_minBytes = "min_bytes"

	// repeated tag: required, min_len, max_len
	// message tag: required
)

type Tag struct {
	Key   string
	Value string
}

// cli is an implementation of the Go protocol buffer compiler's
// plugin architecture. It generates bindings for cli support.
type cli struct {
	generator.PluginImports
	gen        *generator.Generator
	isPkg      generator.Single
	stringsPkg generator.Single
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

	g.isPkg = g.NewImport("github.com/lack-io/vine/util/is", "is")
	g.stringsPkg = g.NewImport("strings", "")

	for i, msg := range file.Messages() {
		g.generateMessage(file, msg, i)
	}
}

func (g *cli) generateMessage(file *generator.FileDescriptor, msg *generator.MessageDescriptor, index int) {
	if msg.Proto.Options != nil && msg.Proto.Options.GetMapEntry() {
		return
	}
	if msg.Proto.File() != file {
		return
	}
	g.P("func (m *", msg.Proto.Name, ") Validate() error {")
	g.P(`return m.ValidateE("")`)
	g.P("}")
	g.P()
	g.P("func (m *", msg.Proto.Name, ") ValidateE(prefix string) error {")
	if g.ignoredMessage(msg) {
		g.P("return nil")
	} else {
		g.P("errs := make([]error, 0)")
		for _, field := range msg.Fields {
			if field.Proto.IsRepeated() {
				g.generateRepeatedField(field)
				continue
			}
			if field.Proto.IsMessage() {
				g.generateMessageField(field)
				continue
			}
			switch field.Proto.GetType() {
			case descriptor.FieldDescriptorProto_TYPE_DOUBLE,
				descriptor.FieldDescriptorProto_TYPE_FLOAT,
				descriptor.FieldDescriptorProto_TYPE_FIXED32,
				descriptor.FieldDescriptorProto_TYPE_FIXED64,
				descriptor.FieldDescriptorProto_TYPE_INT32,
				descriptor.FieldDescriptorProto_TYPE_INT64:
				g.generateNumberField(field)
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				g.generateStringField(field)
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				g.generateBytesField(field)
			case descriptor.FieldDescriptorProto_TYPE_ENUM:
				g.generateEnumField(field)
			}
		}
		g.P(fmt.Sprintf("return %s.MargeErr(errs...)", g.isPkg.Use()))
	}
	g.P("}")
	g.P()
}

func (g *cli) generateNumberField(field *generator.FieldDescriptor) {
	fieldName := generator.CamelCase(*field.Proto.Name)
	tags := g.extractTags(field.Comments)
	if len(tags) == 0 {
		return
	}
	if _, ok := tags[_required]; ok {
		g.P("if int64(m.", fieldName, ") == 0 {")
		g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' is required\", prefix))", field.Proto.GetJsonName()))
		if len(tags) > 1 {
			g.P("} else {")
		}
	} else {
		if tag, ok := tags[_default]; ok {
			g.P("if int64(m.", fieldName, ") == 0 {")
			g.P("m.", fieldName, " = ", tag.Value)
			g.P("}")
		}
		g.P("if int64(m.", fieldName, ") != 0 {")
	}
	for _, tag := range tags {
		switch tag.Key {
		case _enum, _in:
			value := strings.TrimPrefix(tag.Value, "[")
			value = strings.TrimSuffix(value, "]")
			g.P(fmt.Sprintf("if !%s.In([]float64{%s}, float64(m.%s)) {", g.isPkg.Use(), value, fieldName))
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' must in '%s'\", prefix))", field.Proto.GetJsonName(), tag.Value))
			g.P("}")
		case _notIn:
			value := strings.TrimPrefix(tag.Value, "[")
			value = strings.TrimSuffix(value, "]")
			g.P(fmt.Sprintf("if !%s.NotIn([]float64{%s}, float64(m.%s)) {", g.isPkg.Use(), value, fieldName))
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' must not in '%s'\", prefix))", field.Proto.GetJsonName(), tag.Value))
			g.P("}")
		case _eq:
			g.P("if !(m.", fieldName, " == ", tag.Value, ") {")
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' must equal to '%s'\", prefix))", field.Proto.GetJsonName(), tag.Value))
			g.P("}")
		case _ne:
			g.P("if !(m.", fieldName, " != ", tag.Value, ") {")
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' must not equal to '%s'\", prefix))", field.Proto.GetJsonName(), tag.Value))
			g.P("}")
		case _lt:
			g.P("if !(m.", fieldName, " < ", tag.Value, ") {")
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' must less than '%s'\", prefix))", field.Proto.GetJsonName(), tag.Value))
			g.P("}")
		case _lte:
			g.P("if !(m.", fieldName, " <= ", tag.Value, ") {")
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' must less than or equal to '%s'\", prefix))", field.Proto.GetJsonName(), tag.Value))
			g.P("}")
		case _gt:
			g.P("if !(m.", fieldName, " > ", tag.Value, ") {")
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' must great than '%s'\", prefix))", field.Proto.GetJsonName(), tag.Value))
			g.P("}")
		case _gte:
			g.P("if !(m.", fieldName, " >= ", tag.Value, ") {")
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' must great than or equal to '%s'\", prefix))", field.Proto.GetJsonName(), tag.Value))
			g.P("}")
		}
	}
	g.P("}")
}

func (g *cli) generateEnumField(field *generator.FieldDescriptor) {
	enum := g.gen.ExtractEnum(field.Proto.GetTypeName())
	fieldName := generator.CamelCase(field.Proto.GetName())
	tags := g.extractTags(field.Comments)
	if _, ok := tags[_required]; ok {
		g.P("if int32(m.", fieldName, ") == 0 {")
		g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' is required\", prefix))", field.Proto.GetJsonName()))
		if len(tags) > 1 {
			g.P("} else {")
		}
	} else {
		if tag, ok := tags[_default]; ok {
			g.P("if int32(m.", fieldName, ") == 0 {")
			g.P("m.", fieldName, " = ", tag.Value)
			g.P("}")
		}
		g.P("if int32(m.", fieldName, ") != 0 {")
	}

	val := make([]string, 0)
	for _, item := range enum.Value {
		val = append(val, fmt.Sprintf("%d", item.GetNumber()))
	}
	g.P(fmt.Sprintf("if !%s.In([]int32{%s}, int32(m.%s)) {", g.isPkg.Use(), strings.Join(val, ","), fieldName))
	g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' must in '[%s]'\", prefix))", field.Proto.GetJsonName(), strings.Join(val, ", ")))
	g.P("}")
	g.P("}")
}

func (g *cli) generateStringField(field *generator.FieldDescriptor) {
	fieldName := generator.CamelCase(*field.Proto.Name)
	tags := g.extractTags(field.Comments)
	if len(tags) == 0 {
		return
	}
	if _, ok := tags[_required]; ok {
		g.P("if len(m.", fieldName, ") == 0 {")
		g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' is required\", prefix))", field.Proto.GetJsonName()))
		if len(tags) > 1 {
			g.P("} else {")
		}
	} else {
		if tag, ok := tags[_default]; ok {
			g.P("if len(m.", fieldName, ") == 0 {")
			g.P("m.", fieldName, " = ", tag.Value)
			g.P("}")
		}
		g.P("if len(m.", fieldName, ") != 0 {")
	}
	for _, tag := range tags {
		fieldName := generator.CamelCase(*field.Proto.Name)
		switch tag.Key {
		case _enum, _in:
			value := fullStringSlice(tag.Value)
			g.P(fmt.Sprintf("if !%s.In([]string{%s}, string(m.%s)) {", g.isPkg.Use(), value, fieldName))
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' must in '[%s]'\", prefix))", field.Proto.GetJsonName(), strings.ReplaceAll(value, "\"", "")))
			g.P("}")
		case _notIn:
			value := fullStringSlice(tag.Value)
			g.P(fmt.Sprintf("if !%s.NotIn([]string{%s}, string(m.%s)) {", g.isPkg.Use(), value, fieldName))
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' must not in '[%s]'\", prefix))", field.Proto.GetJsonName(), strings.ReplaceAll(value, "\"", "")))
			g.P("}")
		case _minLen:
			g.P("if !(len(m.", fieldName, ") >= ", tag.Value, ") {")
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' length must less than '%s'\", prefix))", field.Proto.GetJsonName(), tag.Value))
			g.P("}")
		case _maxLen:
			g.P("if !(len(m.", fieldName, ") <= ", tag.Value, ") {")
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' length must great than '%s'\", prefix))", field.Proto.GetJsonName(), tag.Value))
			g.P("}")
		case _prefix:
			value := TrimString(tag.Value, "\"")
			g.P("if !strings.HasPrefix(m.", fieldName, ", \"", value, "\") {")
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' must start with '%s'\", prefix))", field.Proto.GetJsonName(), value))
			g.P("}")
		case _suffix:
			value := TrimString(tag.Value, "\"")
			g.P("if !strings.HasSuffix(m.", fieldName, ", \"", value, "\") {")
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' must end with '%s'\", prefix))", field.Proto.GetJsonName(), value))
			g.P("}")
		case _contains:
			value := TrimString(tag.Value, "\"")
			g.P("if !strings.Contains(m.", fieldName, ", \"", value, "\") {")
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' must contain '%s'\", prefix))", field.Proto.GetJsonName(), value))
			g.P("}")
		case _number:
			g.P(fmt.Sprintf("if !%s.Number(m.%s) {", g.isPkg.Use(), fieldName))
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' is not a valid number\", prefix))", field.Proto.GetJsonName()))
			g.P("}")
		case _email:
			g.P(fmt.Sprintf("if !%s.Email(m.%s) {", g.isPkg.Use(), fieldName))
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' is not a valid email\", prefix))", field.Proto.GetJsonName()))
			g.P("}")
		case _ip:
			g.P(fmt.Sprintf("if !%s.IP(m.%s) {", g.isPkg.Use(), fieldName))
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' is not a valid ip\", prefix))", field.Proto.GetJsonName()))
			g.P("}")
		case _ipv4:
			g.P(fmt.Sprintf("if !%s.IPv4(m.%s) {", g.isPkg.Use(), fieldName))
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' is not a valid ipv4\", prefix))", field.Proto.GetJsonName()))
			g.P("}")
		case _ipv6:
			g.P(fmt.Sprintf("if !%s.IPv6(m.%s) {", g.isPkg.Use(), fieldName))
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' is not a valid ipv6\", prefix))", field.Proto.GetJsonName()))
			g.P("}")
		case _crontab:
			g.P(fmt.Sprintf("if !%s.Crontab(m.%s) {", g.isPkg.Use(), fieldName))
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' is not a valid crontab\", prefix))", field.Proto.GetJsonName()))
			g.P("}")
		case _uuid:
			g.P(fmt.Sprintf("if !%s.Uuid(m.%s) {", g.isPkg.Use(), fieldName))
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' is not a valid uuid\", prefix))", field.Proto.GetJsonName()))
			g.P("}")
		case _uri:
			g.P(fmt.Sprintf("if !%s.URL(m.%s) {", g.isPkg.Use(), fieldName))
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' is not a valid url\", prefix))", field.Proto.GetJsonName()))
			g.P("}")
		case _domain:
			g.P(fmt.Sprintf("if !%s.Domain(m.%s) {", g.isPkg.Use(), fieldName))
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' is not a valid domain\", prefix))", field.Proto.GetJsonName()))
			g.P("}")
		case _pattern:
			value := TrimString(tag.Value, "`")
			g.P(fmt.Sprintf("if !%s.Re(`%s`, m.%s) {", g.isPkg.Use(), value, fieldName))
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(`field '%%s%s' is not a valid pattern '%s'`, prefix))", field.Proto.GetJsonName(), value))
			g.P("}")
		}
	}
	g.P("}")
}

func (g *cli) generateBytesField(field *generator.FieldDescriptor) {
	fieldName := generator.CamelCase(*field.Proto.Name)
	tags := g.extractTags(field.Comments)
	if len(tags) == 0 {
		return
	}
	if _, ok := tags[_required]; ok {
		g.P("if len(m.", fieldName, ") == 0 {")
		g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' is required\", prefix))", field.Proto.GetJsonName()))
		if len(tags) > 1 {
			g.P("} else {")
		}
	} else {
		g.P("if len(m.", fieldName, ") != 0 {")
	}
	for _, tag := range tags {
		fieldName := generator.CamelCase(field.Proto.GetName())
		switch tag.Key {
		case _minBytes:
			g.P("if !(len(m.", fieldName, ") <= ", tag.Value, ") {")
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' length must less than '%s'\", prefix))", field.Proto.GetJsonName(), tag.Value))
			g.P("}")
		case _maxBytes:
			g.P("if !(len(m.", fieldName, ") >= ", tag.Value, ") {")
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' length must great than '%s'\", prefix))", field.Proto.GetJsonName(), tag.Value))
			g.P("}")
		}
	}
	g.P("}")
}

func (g *cli) generateRepeatedField(field *generator.FieldDescriptor) {
	fieldName := generator.CamelCase(*field.Proto.Name)
	tags := g.extractTags(field.Comments)
	if len(tags) == 0 {
		return
	}
	if _, ok := tags[_required]; ok {
		g.P("if len(m.", fieldName, ") == 0 {")
		g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' is required\", prefix))", field.Proto.GetJsonName()))
		if len(tags) > 1 {
			g.P("} else {")
		}
	} else {
		g.P("if len(m.", fieldName, ") != 0 {")
	}
	for _, tag := range tags {
		fieldName := generator.CamelCase(*field.Proto.Name)
		switch tag.Key {
		case _minLen:
			g.P("if !(len(m.", fieldName, ") <= ", tag.Value, ") {")
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' length must less than '%s'\", prefix))", field.Proto.GetJsonName(), tag.Value))
			g.P("}")
		case _maxLen:
			g.P("if !(len(m.", fieldName, ") >= ", tag.Value, ") {")
			g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' length must great than '%s'\", prefix))", field.Proto.GetJsonName(), tag.Value))
			g.P("}")
		}
	}
	g.P("}")
}

func (g *cli) generateMessageField(field *generator.FieldDescriptor) {
	tags := g.extractTags(field.Comments)
	if len(tags) == 0 {
		return
	}
	if _, ok := tags[_inline]; ok {
		smsg := g.gen.ExtractMessage(field.Proto.GetTypeName())
		for _, field := range smsg.Fields {
			if field.Proto.IsRepeated() {
				g.generateRepeatedField(field)
				continue
			}
			if field.Proto.IsMessage() {
				g.generateMessageField(field)
				continue
			}
			switch *field.Proto.Type {
			case descriptor.FieldDescriptorProto_TYPE_DOUBLE,
				descriptor.FieldDescriptorProto_TYPE_FLOAT,
				descriptor.FieldDescriptorProto_TYPE_FIXED32,
				descriptor.FieldDescriptorProto_TYPE_FIXED64,
				descriptor.FieldDescriptorProto_TYPE_INT32,
				descriptor.FieldDescriptorProto_TYPE_INT64:
				g.generateNumberField(field)
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				g.generateStringField(field)
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				g.generateBytesField(field)
			}
		}
		return
	}
	if _, ok := tags[_required]; ok {
		fname := generator.CamelCase(*field.Proto.Name)
		g.P("if m.", fname, " == nil {")
		g.P(fmt.Sprintf("errs = append(errs, fmt.Errorf(\"field '%%s%s' is required\", prefix))", field.Proto.GetJsonName()))
		g.P("} else {")
		g.P(fmt.Sprintf("errs = append(errs, m.%s.ValidateE(prefix+\"%s.\"))", fname, field.Proto.GetJsonName()))
		g.P("}")
	}
}

func (g *cli) ignoredMessage(msg *generator.MessageDescriptor) bool {
	tags := g.extractTags(msg.Comments)
	for _, c := range tags {
		if c.Key == _ignore {
			return true
		}
	}
	return false
}

func TrimString(s string, c string) string {
	s = strings.TrimPrefix(s, c)
	s = strings.TrimSuffix(s, c)
	return s
}

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

func (g *cli) extractTags(comments []*generator.Comment) map[string]*Tag {
	if comments == nil || len(comments) == 0 {
		return nil
	}
	tags := make(map[string]*Tag, 0)
	for _, c := range comments {
		if c.Tag != TagString || len(c.Text) == 0 {
			continue
		}
		if strings.HasPrefix(c.Text, _pattern) {
			if i := strings.Index(c.Text, "="); i == -1 {
				g.gen.Fail("invalid pattern format")
			} else {
				pa := string(c.Text[i+1])
				pe := string(c.Text[len(c.Text)-1])
				if pa != "`" || pe != "`" {
					g.gen.Fail(fmt.Sprintf("invalid pattern value, pa=%s, pe=%s", pa, pe))
				}
				key := strings.TrimSpace(c.Text[:i])
				value := strings.TrimSpace(c.Text[i+1:])
				if len(value) == 0 {
					g.gen.Fail(fmt.Sprintf("tag '%s' missing value", key))
				}
				tags[key] = &Tag{
					Key:   key,
					Value: value,
				}
			}
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
