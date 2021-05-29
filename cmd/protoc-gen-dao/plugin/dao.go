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
	"errors"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

	"github.com/lack-io/vine/cmd/generator"
	"github.com/lack-io/vine/lib/dao/schema"
)

var TagString = "gen"

const (
	// message tags
	// dao generate flag
	_dao   = "dao"
	_table = "table"

	// field tags
	// inline
	_inline = "inline"
	// dao primary key
	_pk = "pk"
)

type Tag struct {
	Key   string
	Value string
}

// dao is an implementation of the Go protocol buffer compiler's
// plugin architecture. It generates bindings for dao support.
type dao struct {
	generator.PluginImports
	gen *generator.Generator

	schemas     []*Schema
	aliasTypes  map[string]string
	aliasFields map[string]*Field

	sourcePkg string
	ctxPkg    generator.Single
	timePkg   generator.Single
	stringPkg generator.Single
	errPkg    generator.Single
	DriverPkg generator.Single
	jsonPkg   generator.Single
	daoPkg    generator.Single
	clausePkg generator.Single
}

func New() *dao {
	return &dao{
		schemas:     []*Schema{},
		aliasTypes:  map[string]string{},
		aliasFields: map[string]*Field{},
	}
}

// Name returns the name of this plugin, "dao".
func (g *dao) Name() string {
	return "dao"
}

// Init initializes the plugin.
func (g *dao) Init(gen *generator.Generator) {
	g.gen = gen
	g.PluginImports = generator.NewPluginImports(g.gen)
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (g *dao) objectNamed(name string) generator.Object {
	g.gen.RecordTypeUse(name)
	return g.gen.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (g *dao) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

// P forwards to g.gen.P.
func (g *dao) P(args ...interface{}) { g.gen.P(args...) }

// Generate generates code for the services in the given file.
func (g *dao) Generate(file *generator.FileDescriptor) {
	if len(file.Comments()) == 0 {
		return
	}

	g.ctxPkg = g.NewImport("context", "context")
	g.timePkg = g.NewImport("time", "time")
	g.stringPkg = g.NewImport("strings", "strings")
	g.DriverPkg = g.NewImport("database/sql/driver", "driver")
	g.jsonPkg = g.NewImport("github.com/json-iterator/go", "json")
	g.errPkg = g.NewImport("errors", "errors")
	g.daoPkg = g.NewImport("github.com/lack-io/vine/lib/dao", "dao")
	g.clausePkg = g.NewImport("github.com/lack-io/vine/lib/dao/clause", "clause")
	if g.gen.OutPut.Load {
		g.sourcePkg = string(g.gen.AddImport(generator.GoImportPath(g.gen.OutPut.SourcePkgPath)))
	}

	for _, msg := range file.Messages() {
		g.wrapSchemas(file, msg)
	}

	aFields := make([]*Field, 0)
	for _, value := range g.aliasFields {
		if value.File.GetName() != file.GetName() {
			continue
		}
		aFields = append(aFields, value)
	}
	sort.Slice(aFields, func(i, j int) bool { return aFields[i].Num < aFields[j].Num })
	for _, value := range aFields {
		f := strings.TrimSuffix(filepath.Clean(file.GetName()), ".proto")
		// ignore unique alias
		if g.isContains(filepath.Join(build.Default.GOPATH, "src", g.gen.OutPut.Out), f, value.Alias) {
			continue
		}
		g.generateAliasField(file, value)
	}

	for _, item := range g.schemas {
		if item.Desc.Proto.File().GetName() != file.GetName() {
			continue
		}
		g.generateSchema(file, item)
	}
}

func (g *dao) wrapSchemas(file *generator.FileDescriptor, msg *generator.MessageDescriptor) {
	if msg.Proto.Options != nil && msg.Proto.Options.GetMapEntry() {
		return
	}

	tags := g.extractTags(msg.Comments)
	if _, ok := tags[_dao]; !ok {
		return
	}

	table := toTableName(msg.Proto.GetName())
	if v, ok := tags[_table]; ok {
		table = v.Value
	}

	s := &Schema{
		Name:    msg.Proto.GetName() + "S",
		Desc:    msg,
		MFields: map[string]*Field{},
		Table:   table,
	}
	n := 0
	g.buildFields(file, msg, s, &n)
	if s.PK == nil {
		g.gen.Fail(fmt.Sprintf(`Message:%s missing primary key`, msg.Proto.GetName()))
	}

	s.Fields = make([]*Field, 0)
	for _, item := range s.MFields {
		s.Fields = append(s.Fields, item)
	}
	sort.Slice(s.Fields, func(i, j int) bool { return s.Fields[i].Num < s.Fields[j].Num })
	g.schemas = append(g.schemas, s)
}

func (g *dao) buildFields(file *generator.FileDescriptor, m *generator.MessageDescriptor, s *Schema, n *int) {
	for _, item := range m.Fields {
		fTags := g.extractTags(item.Comments)

		_, isInline := fTags[_inline]
		if isInline && item.Proto.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			subMsg := g.gen.ExtractMessage(item.Proto.GetTypeName())
			g.buildFields(file, subMsg, s, n)
			continue
		}

		field := &Field{
			Name: generator.CamelCase(item.Proto.GetName()),
			Tags: []*FieldTag{},
			Desc: item,
			File: file,
			Num:  *n,
		}
		if item.Proto.IsRepeated() {
			alias := generator.CamelCaseSlice([]string{m.Proto.GetName(), item.Proto.GetName()})
			if strings.HasSuffix(item.Proto.GetTypeName(), "Entry") {
				field.Type = _map
				for _, nest := range m.Proto.GetNestedType() {
					if strings.HasSuffix(item.Proto.GetTypeName(), nest.GetName()) {
						field.Map = &MapFields{}
						field.Map.Key, field.Map.Value = nest.GetMapFields()
						break
					}
				}
			} else {
				field.Type = _slice
				field.Slice = field.Desc.Proto
				field.IsRepeated = true
			}
			field.Alias = alias
			field.Tags = append(field.Tags, g.buildJSONTags(item), g.buildDaoTags(item, false))
			g.aliasFields[alias] = field
		} else {
			typ, tags, err := g.buildFieldTypeAndTags(item)
			if err != nil {
				continue
			}
			field.Type = typ
			field.Tags = tags
			if item.Proto.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
				alias := generator.CamelCaseSlice([]string{m.Proto.GetName(), item.Proto.GetName()})
				field.Alias = alias
				g.aliasFields[alias] = field
			}

			_, pkExists := fTags[_pk]
			if pkExists && s.PK == nil && (strings.ToLower(field.Name) == "id" || strings.ToLower(field.Name) == "uuid") {
				s.PK = field
				for _, tag := range field.Tags {
					if tag.Key == "dao" {
						tag.Values = append(tag.Values, "primaryKey")
					}
				}
			}
		}

		*n += 1
		s.MFields[field.Name] = field
	}
}

func (g *dao) generateAliasField(file *generator.FileDescriptor, field *Field) {
	alias := field.Alias
	if field.IsRepeated {
		// slice, array type
		typ, err := g.buildFieldGoType(file, field.Desc.Proto)
		if err != nil {
			return
		}
		g.P(fmt.Sprintf(`// %s the alias of []%s`, alias, typ))
		g.P(fmt.Sprintf(`type %s []%s`, alias, typ))

	}
	if field.Type == _map {
		if field.Map == nil {
			return
		}

		key, value := field.Map.Key, field.Map.Value
		keyString, _ := g.buildFieldGoType(file, key)
		valueString, _ := g.buildFieldGoType(file, value)
		g.P(fmt.Sprintf(`// %s the alias of map[%s]%s`, alias, keyString, valueString))
		g.P(fmt.Sprintf(`type %s map[%s]%s`, alias, keyString, valueString))
	}
	if field.Type == _point {
		var typ string
		subMsg := g.extractMessage(field.Desc.Proto.GetTypeName())
		if subMsg.Proto.File() == file {
			typ = g.wrapPkg(subMsg.Proto.GetName())
		} else {
			obj := g.gen.ObjectNamed(field.Desc.Proto.GetTypeName())
			v, ok := g.gen.ImportMap[obj.GoImportPath().String()]
			if !ok {
				v = string(g.gen.AddImport(obj.GoImportPath()))
			}
			typ = fmt.Sprintf("%s.%s", v, subMsg.Proto.GetName())
		}
		g.P(fmt.Sprintf(`// %s the alias of %s`, alias, typ))
		g.P(fmt.Sprintf("type %s %s", alias, typ))
	}
	g.P()

	g.P("// Value return json value, implement driver.Valuer interface")
	switch field.Type {
	case _slice:
		g.P(fmt.Sprintf(`func (m %s) Value() (%s.Value, error) {`, alias, g.DriverPkg.Use()))
		g.P("if len(m) == 0 {")
		g.P("return nil, nil")
		g.P("}")
	case _map:
		g.P(fmt.Sprintf(`func (m %s) Value() (%s.Value, error) {`, alias, g.DriverPkg.Use()))
		g.P("if m == nil {")
		g.P("return nil, nil")
		g.P("}")
	default:
		g.P(fmt.Sprintf(`func (m *%s) Value() (%s.Value, error) {`, alias, g.DriverPkg.Use()))
		g.P("if m == nil {")
		g.P("return nil, nil")
		g.P("}")
	}
	g.P(fmt.Sprintf(`b, err := %s.Marshal(m)`, g.jsonPkg.Use()))
	g.P(`return string(b), err`)
	g.P("}")
	g.P()

	g.P("// Scan scan value into Jsonb, implements sql.Scanner interface")
	g.P(fmt.Sprintf(`func (m *%s) Scan(value interface{}) error {`, alias))
	g.P("var bytes []byte")
	g.P("switch v := value.(type) {")
	g.P("case []byte:")
	g.P("bytes = v")
	g.P("case string:")
	g.P("bytes = []byte(v)")
	g.P("default:")
	g.P(fmt.Sprintf(`return %s.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))`, g.errPkg.Use()))
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`return %s.Unmarshal(bytes, &m)`, g.jsonPkg.Use()))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) DaoDataType() string {`, alias))
	g.P(fmt.Sprintf(`return %s.DefaultDialect.JSONDataType()`, g.daoPkg.Use()))
	g.P("}")
	g.P()
}

func (g *dao) generateSchema(file *generator.FileDescriptor, schema *Schema) {

	g.generateSchemaFields(file, schema)

	g.generateSchemaIOMethods(file, schema)

	g.generateSchemaUtilMethods(file, schema)

	g.generateSchemaCURDMethods(file, schema)
}

func (g *dao) generateSchemaFields(file *generator.FileDescriptor, schema *Schema) {
	g.P(fmt.Sprintf(`// %s the Schema for %s`, schema.Name, schema.Desc.Proto.GetName()))
	g.P("type ", schema.Name, " struct {")
	g.P(fmt.Sprintf(`tx *dao.DB %s`, toQuoted(`json:"-" dao:"-"`)))
	g.P(fmt.Sprintf(`exprs []%s.Expression %s`, g.clausePkg.Use(), toQuoted(`json:"-" dao:"-"`)))
	g.P()
	for _, field := range schema.Fields {
		switch field.Type {
		case _point:
			g.P(fmt.Sprintf(`%s *%s %s`, field.Name, field.Alias, MargeTags(field.Tags...)))
		case _slice, _map:
			g.P(fmt.Sprintf(`%s %s %s`, field.Name, field.Alias, MargeTags(field.Tags...)))
		default:
			g.P(fmt.Sprintf(`%s %s %s`, field.Name, field.Type, MargeTags(field.Tags...)))
		}
	}
	g.P(fmt.Sprintf(`DeletionTimestamp int64 %s`, toQuoted(`json:"deletionTimestamp,omitempty" dao:"column:deletion_timestamp"`)))
	g.P("}")
	g.P()
}

func (g *dao) generateSchemaIOMethods(file *generator.FileDescriptor, schema *Schema) {
	sname := schema.Name
	pname := schema.Desc.Proto.GetName()

	g.P(fmt.Sprintf(`func Registry%s() error {`, pname))
	g.P(fmt.Sprintf(`return dao.DefaultDialect.Migrator().AutoMigrate(&%s{})`, sname))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func %sBuilder() *%s {`, sname, sname))
	g.P(fmt.Sprintf(`return &%s{tx: dao.DefaultDialect.NewTx()}`, sname))
	g.P("}")
	g.P()

	for _, field := range schema.Fields {
		switch field.Type {
		case _float32, _float64, _int32, _int64, _uint32, _uint64, _string:
			g.P(fmt.Sprintf(`func (m *%s) Set%s(in %s) *%s {`, sname, field.Name, field.Type, sname))
			g.P(fmt.Sprintf(`m.%s = in`, field.Name))
			g.P(`return m`)
			g.P("}")
			g.P()
		case _map:
			key, value := field.Map.Key, field.Map.Value
			keyString, _ := g.buildFieldGoType(file, key)
			valueString, _ := g.buildFieldGoType(file, value)
			g.P(fmt.Sprintf(`func (m *%s) Set%s(in map[%s]%s) *%s {`, sname, field.Name, keyString, valueString, sname))
			g.P(fmt.Sprintf(`m.%s = in`, field.Name))
			g.P(`return m`)
			g.P("}")
			g.P()
			g.P(fmt.Sprintf(`func (m *%s) Put%s(k %s, v %s) *%s {`, sname, field.Name, keyString, valueString, sname))
			g.P(fmt.Sprintf(`if len(m.%s) == 0 {`, field.Name))
			g.P(fmt.Sprintf(`m.%s = map[%s]%s{}`, field.Name, keyString, valueString))
			g.P(`}`)
			g.P(fmt.Sprintf(`m.%s[k] = v`, field.Name))
			g.P(`return m`)
			g.P("}")
			g.P()
			g.P(fmt.Sprintf(`func (m *%s) Remove%s(k %s) *%s {`, sname, field.Name, keyString, sname))
			g.P(fmt.Sprintf(`if len(m.%s) == 0 {`, field.Name))
			g.P(`return m`)
			g.P(`}`)
			g.P(fmt.Sprintf(`delete(m.%s, k)`, field.Name))
			g.P(`return m`)
			g.P("}")
			g.P()
		case _slice:
			typ, _ := g.buildFieldGoType(file, field.Slice)
			g.P(fmt.Sprintf(`func (m *%s) Set%s(in []%s) *%s {`, sname, field.Name, typ, sname))
			g.P(fmt.Sprintf(`m.%s = in`, field.Name))
			g.P(`return m`)
			g.P("}")
			g.P()
			g.P(fmt.Sprintf(`func (m *%s) Add%s(in %s) *%s {`, sname, field.Name, typ, sname))
			g.P(fmt.Sprintf(`if len(m.%s) == 0 {`, field.Name))
			g.P(fmt.Sprintf(`m.%s = []%s{}`, field.Name, typ))
			g.P(`}`)
			g.P(fmt.Sprintf(`m.%s = append(m.%s, in)`, field.Name, field.Name))
			g.P(`return m`)
			g.P("}")
			g.P()
			g.P(fmt.Sprintf(`func (m *%s) Filter%s(fn func(item %s) bool) *%s {`, sname, field.Name, typ, sname))
			g.P(fmt.Sprintf(`out := make([]%s, 0)`, typ))
			g.P(fmt.Sprintf(`for _, item := range m.%s {`, field.Name))
			g.P(`if !fn(item) {`)
			g.P(`out = append(out, item)`)
			g.P(`}`)
			g.P(`}`)
			g.P(fmt.Sprintf(`m.%s = out`, field.Name))
			g.P(`return m`)
			g.P(`}`)
			g.P()
			g.P(fmt.Sprintf(`func (m *%s) Remove%s(index int) *%s {`, sname, field.Name, sname))
			g.P(fmt.Sprintf(`if index < 0 || index >= len(m.%s) {`, field.Name))
			g.P(`return m`)
			g.P(`}`)
			g.P(fmt.Sprintf(`if index == len(m.%s)-1 {`, field.Name))
			g.P(fmt.Sprintf(`m.%s = m.%s[:index-1]`, field.Name, field.Name))
			g.P(`} else {`)
			g.P(fmt.Sprintf(`m.%s = append(m.%s[0:index], m.%s[index+1:]...)`, field.Name, field.Name, field.Name))
			g.P(`}`)
			g.P(`return m`)
			g.P("}")
			g.P()
		case _point:
			var typ string
			subMsg := g.extractMessage(field.Desc.Proto.GetTypeName())
			if subMsg.Proto.File() == file {
				typ = g.wrapPkg(subMsg.Proto.GetName())
			} else {
				obj := g.gen.ObjectNamed(field.Desc.Proto.GetTypeName())
				v, ok := g.gen.ImportMap[obj.GoImportPath().String()]
				if !ok {
					v = string(g.gen.AddImport(obj.GoImportPath()))
				}
				typ = fmt.Sprintf(`%s.%s`, v, subMsg.Proto.GetName())
			}
			g.P(fmt.Sprintf(`func (m *%s) Set%s(in *%s) *%s {`, sname, field.Name, typ, sname))
			g.P(fmt.Sprintf(`m.%s = (*%s)(in)`, field.Name, field.Alias))
			g.P(`return m`)
			g.P("}")
			g.P()
		}
	}

	g.P(fmt.Sprintf(`func From%s(in *%s) *%s {`, pname, g.wrapPkg(pname), sname))
	g.P(fmt.Sprintf(`out := new(%s)`, sname))
	g.P(`out.tx = dao.DefaultDialect.NewTx()`)
	for _, field := range schema.Fields {
		switch field.Type {
		case _float32, _float64, _int32, _int64, _uint32, _uint64:
			g.P(fmt.Sprintf(`if in.%s != 0 {`, field.Name))
			if field.Desc.Proto.GetType() == descriptor.FieldDescriptorProto_TYPE_ENUM {
				g.P(fmt.Sprintf(`out.%s = int32(in.%s)`, field.Name, field.Name))
			} else {
				g.P(fmt.Sprintf(`out.%s = in.%s`, field.Name, field.Name))
			}
			g.P("}")
		case _map:
			g.P(fmt.Sprintf(`if in.%s != nil {`, field.Name))
			g.P(fmt.Sprintf(`out.%s = in.%s`, field.Name, field.Name))
			g.P("}")
		case _string:
			if field.Desc.Proto.GetType() == descriptor.FieldDescriptorProto_TYPE_BYTES {
				g.P(fmt.Sprintf(`if in.%s != nil {`, field.Name))
				g.P(fmt.Sprintf(`out.%s = string(in.%s)`, field.Name, field.Name))
			} else {
				g.P(fmt.Sprintf(`if in.%s != "" {`, field.Name))
				g.P(fmt.Sprintf(`out.%s = in.%s`, field.Name, field.Name))
			}
			g.P("}")
		case _slice:
			g.P(fmt.Sprintf(`if len(in.%s) != 0 {`, field.Name))
			g.P(fmt.Sprintf(`out.%s = in.%s`, field.Name, field.Name))
			g.P("}")
		case _point:
			g.P(fmt.Sprintf(`if in.%s != nil {`, field.Name))
			g.P(fmt.Sprintf(`out.%s = (*%s)(in.%s)`, field.Name, field.Alias, field.Name))
			g.P("}")
		}
	}
	g.P("return out")
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) To%s() *%s {`, sname, pname, g.wrapPkg(pname)))
	g.P(fmt.Sprintf(`out := new(%s)`, g.wrapPkg(pname)))
	for _, field := range schema.Fields {
		switch field.Type {
		case _point:
			var typPkg string
			obj := g.objectNamed(field.Desc.Proto.GetTypeName())
			submsg := g.extractMessage(field.Desc.Proto.GetTypeName())
			typPkg = submsg.Proto.GetName()
			if obj.File() != file {
				typPkg = string(g.gen.AddImport(obj.GoImportPath())) + "." + submsg.Proto.GetName()
			} else {
				typPkg = g.wrapPkg(submsg.Proto.GetName())
			}
			g.P(fmt.Sprintf("out.%s = (*%s)(m.%s)", field.Name, typPkg, field.Name))
		case _string:
			if field.Desc.Proto.GetType() == descriptor.FieldDescriptorProto_TYPE_BYTES {
				g.P(fmt.Sprintf(`out.%s = []byte(m.%s)`, field.Name, field.Name))
			} else {
				g.P(fmt.Sprintf(`out.%s = m.%s`, field.Name, field.Name))
			}
		default:
			if field.Desc.Proto.IsEnum() {
				enum := g.gen.ExtractEnum(field.Desc.Proto.GetTypeName())
				g.P(fmt.Sprintf(`out.%s = %s(m.%s)`, field.Name, g.wrapPkg(enum.GetName()), field.Name))
			} else {
				g.P(fmt.Sprintf(`out.%s = m.%s`, field.Name, field.Name))
			}
		}
	}
	g.P("return out")
	g.P("}")
	g.P()
}

func (g *dao) generateSchemaUtilMethods(file *generator.FileDescriptor, schema *Schema) {
	g.P("func (", schema.Name, ") TableName() string {")
	g.P(`return "`, schema.Table, `"`)
	g.P("}")
	g.P()

	g.P("func (m ", schema.Name, ") PrimaryKey() (string, interface{}, bool) {")
	if schema.PK.Type == _string {
		g.P(fmt.Sprintf(`return "%s", m.%s, m.%s == ""`, toColumnName(schema.PK.Name), schema.PK.Name, schema.PK.Name))
	} else {
		g.P(fmt.Sprintf(`return "%s", m.%s, m.%s == 0`, toColumnName(schema.PK.Name), schema.PK.Name, schema.PK.Name))
	}
	g.P("}")
	g.P()
}

func (g *dao) generateSchemaCURDMethods(file *generator.FileDescriptor, schema *Schema) {
	source, target := schema.Name, schema.Desc.Proto.GetName()
	g.P(fmt.Sprintf(`func (m *%s) FindPage(ctx %s.Context, page, size int) ([]*%s, int64, error) {`, source, g.ctxPkg.Use(), g.wrapPkg(target)))
	g.P(`pk, _, _ := m.PrimaryKey()`)
	g.P()
	g.P(`m.exprs = append(m.exprs,`)
	g.P(fmt.Sprintf(`%s.OrderBy{Columns: []%s.OrderByColumn{{Column: %s.Column{Table: m.TableName(), Name: pk}, Desc: true}}},`, g.clausePkg.Use(), g.clausePkg.Use(), g.clausePkg.Use()))
	g.P(fmt.Sprintf(`%s.Limit{Offset: (page - 1) * size, Limit: size},`, g.clausePkg.Use()))
	g.P(fmt.Sprintf(`%s.Cond().Build("deletion_timestamp", 0),`, g.clausePkg.Use()))
	g.P(`)`)
	g.P()
	g.P(`data, err := m.findAll(ctx)`)
	g.P(`if err != nil {`)
	g.P(`return nil, 0, err`)
	g.P("}")
	g.P()
	g.P(`total, _ := m.Count(ctx)`)
	g.P(`return data, total, nil`)
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) FindAll(ctx %s.Context) ([]*%s, error) {`, source, g.ctxPkg.Use(), g.wrapPkg(target)))
	g.P(fmt.Sprintf(`m.exprs = append(m.exprs, %s.Cond().Build("deletion_timestamp", 0))`, g.clausePkg.Use()))
	g.P(`return m.findAll(ctx)`)
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) FindPureAll(ctx %s.Context) ([]*%s, error) {`, source, g.ctxPkg.Use(), g.wrapPkg(target)))
	g.P(`return m.findAll(ctx)`)
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) findAll(ctx %s.Context) ([]*%s, error) {`, source, g.ctxPkg.Use(), g.wrapPkg(target)))
	g.P(fmt.Sprintf(`dest := make([]*%s, 0)`, source))
	g.P(fmt.Sprintf(`tx := m.tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx)`, g.daoPkg.Use()))
	g.P()
	g.P(`clauses := append(m.extractClauses(tx), m.exprs...)`)
	g.P(`if err := tx.Clauses(clauses...).Find(&dest).Error; err != nil {`)
	g.P("return nil, err")
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`outs := make([]*%s, len(dest))`, g.wrapPkg(target)))
	g.P(`for i := range dest {`)
	g.P(fmt.Sprintf(`outs[i] = dest[i].To%s()`, target))
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`return outs, nil`))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) Count(ctx %s.Context) (total int64, err error) {`, source, g.ctxPkg.Use()))
	g.P(fmt.Sprintf(`tx := m.tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx)`, g.daoPkg.Use()))
	g.P()
	g.P(fmt.Sprintf(`clauses := append(m.extractClauses(tx), %s.Cond().Build("deletion_timestamp", 0))`, g.clausePkg.Use()))
	g.P(`clauses = append(clauses, m.exprs...)`)
	g.P()
	g.P(`err = tx.Clauses(clauses...).Count(&total).Error`)
	g.P(`return`)
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) FindOne(ctx %s.Context) (*%s, error) {`, source, g.ctxPkg.Use(), g.wrapPkg(target)))
	g.P(fmt.Sprintf(`tx := m.tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx)`, g.daoPkg.Use()))
	g.P()
	g.P(fmt.Sprintf(`clauses := append(m.extractClauses(tx), %s.Cond().Build("deletion_timestamp", 0))`, g.clausePkg.Use()))
	g.P(`clauses = append(clauses, m.exprs...)`)
	g.P()
	g.P(`if err := tx.Clauses(clauses...).First(&m).Error; err != nil {`)
	g.P("return nil, err")
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`return m.To%s(), nil`, target))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) Cond(exprs ...%s.Expression) *%s {`, source, g.clausePkg.Use(), source))
	g.P(`m.exprs = append(m.exprs, exprs...)`)
	g.P(`return m`)
	g.P(`}`)
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) extractClauses(tx *%s.DB) []%s.Expression {`, source, g.daoPkg.Use(), g.clausePkg.Use()))
	g.P(`exprs := make([]clause.Expression, 0)`)
	for _, field := range schema.Fields {
		column := toColumnName(field.Name)
		switch field.Type {
		case _map:
			g.P(fmt.Sprintf(`if m.%s != nil {`, field.Name))
			g.P(fmt.Sprintf(`for k, v := range m.%s {`, field.Name))
			switch field.Map.Key.GetType() {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				g.P(fmt.Sprintf(`exprs = append(exprs, %s.DefaultDialect.JSONBuild("%s").Tx(tx).Op(%s.JSONEq, v, k))`, g.daoPkg.Use(), column, g.daoPkg.Use()))
			case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_INT64:
				g.P(fmt.Sprintf(`exprs = append(exprs, %s.DefaultDialect.JSONBuild("%s").Tx(tx).Op(%s.JSONEq, v, %s))`, g.daoPkg.Use(), column, g.daoPkg.Use(), `fmt.Sprintf("%d", k)`))
			}
			g.P("}")
			g.P("}")
		case _point:
			g.P(fmt.Sprintf(`if m.%s != nil {`, field.Name))
			g.P(fmt.Sprintf(`for k, v := range dao.FieldPatch(m.%s) {`, field.Name))
			g.P(`if v == nil {`)
			g.P(fmt.Sprintf(`exprs = append(exprs, %s.DefaultDialect.JSONBuild("%s").Tx(tx).Op(%s.JSONHasKey, "", %s.Split(k, ".")...))`, g.daoPkg.Use(), column, g.daoPkg.Use(), g.stringPkg.Use()))
			g.P(`} else {`)
			g.P(fmt.Sprintf(`exprs = append(exprs, %s.DefaultDialect.JSONBuild("%s").Tx(tx).Op(%s.JSONEq, v, %s.Split(k, ".")...))`, g.daoPkg.Use(), column, g.daoPkg.Use(), g.stringPkg.Use()))
			g.P("}")
			g.P("}")
			g.P("}")
		case _slice:
			g.P(fmt.Sprintf(`if len(m.%s) != 0 {`, field.Name))
			g.P(fmt.Sprintf(`for _, item := range m.%s {`, field.Name))
			switch field.Slice.GetType() {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				g.P(fmt.Sprintf(`exprs = append(exprs, %s.DefaultDialect.JSONBuild("%s").Tx(tx).Contains(%s.JSONEq, item))`, g.daoPkg.Use(), column, g.daoPkg.Use()))
			case descriptor.FieldDescriptorProto_TYPE_UINT32,
				descriptor.FieldDescriptorProto_TYPE_UINT64,
				descriptor.FieldDescriptorProto_TYPE_INT32,
				descriptor.FieldDescriptorProto_TYPE_INT64,
				descriptor.FieldDescriptorProto_TYPE_SFIXED32,
				descriptor.FieldDescriptorProto_TYPE_SFIXED64,
				descriptor.FieldDescriptorProto_TYPE_FIXED32,
				descriptor.FieldDescriptorProto_TYPE_FIXED64:
				g.P(fmt.Sprintf(`exprs = append(exprs, %s.DefaultDialect.JSONBuild("%s").Tx(tx).Contains(%s.JSONEq, item))`, g.daoPkg.Use(), column, g.daoPkg.Use()))
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				g.P(`for k, v := range dao.FieldPatch(item) {`)
				g.P(`if v != nil {`)
				g.P(fmt.Sprintf(`exprs = append(exprs, %s.DefaultDialect.JSONBuild("%s").Tx(tx).Contains(%s.JSONEq, v, %s.Split(k, ".")...))`, g.daoPkg.Use(), column, g.daoPkg.Use(), g.stringPkg.Use()))
				g.P(`}`)
				g.P("}")
			}
			g.P("}")
			g.P("}")
		case _string:
			g.P(fmt.Sprintf(`if m.%s != "" {`, field.Name))
			g.P(fmt.Sprintf(`exprs = append(exprs, %s.Cond().Build("%s", m.%s))`, g.clausePkg.Use(), column, field.Name))
			g.P("}")
		default:
			g.P(fmt.Sprintf(`if m.%s != 0 {`, field.Name))
			g.P(fmt.Sprintf(`exprs = append(exprs,  %s.Cond().Build("%s", m.%s))`, g.clausePkg.Use(), column, field.Name))
			g.P("}")
		}
	}
	g.P()
	g.P(`return exprs`)
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) Create(ctx %s.Context) (*%s, error) {`, source, g.ctxPkg.Use(), g.wrapPkg(target)))
	g.P(fmt.Sprintf(`tx := m.tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx)`, g.daoPkg.Use()))
	g.P()
	g.P(`if err := tx.Create(m).Error; err != nil {`)
	g.P("return nil, err")
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`return m.To%s(), nil`, target))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) BatchUpdates(ctx %s.Context) error {`, source, g.ctxPkg.Use()))
	g.P(`if len(m.exprs) == 0 {`)
	g.P(fmt.Sprintf(`return %s.New("missing conditions")`, g.errPkg.Use()))
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`tx := m.tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx)`, g.daoPkg.Use()))
	g.P()
	g.P(`values := make(map[string]interface{}, 0)`)
	for _, field := range schema.Fields {
		column := toColumnName(field.Name)
		if field == schema.PK {
			continue
		}
		switch field.Type {
		case _point, _map:
			g.P(fmt.Sprintf(`if m.%s != nil {`, field.Name))
			g.P(fmt.Sprintf(`values["%s"] = m.%s`, column, field.Name))
			g.P("}")
		case _slice:
			g.P(fmt.Sprintf(`if len(m.%s) != 0 {`, field.Name))
			g.P(fmt.Sprintf(`values["%s"] = m.%s`, column, field.Name))
			g.P("}")
		case _string:
			g.P(fmt.Sprintf(`if m.%s != "" {`, field.Name))
			g.P(fmt.Sprintf(`values["%s"] = m.%s`, column, field.Name))
			g.P("}")
		default:
			g.P(fmt.Sprintf(`if m.%s != 0 {`, field.Name))
			g.P(fmt.Sprintf(`values["%s"] = m.%s`, column, field.Name))
			g.P("}")
		}
	}
	g.P()
	g.P(`return tx.Clauses(m.exprs...).Updates(values).Error`)
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) Updates(ctx %s.Context) (*%s, error) {`, source, g.ctxPkg.Use(), g.wrapPkg(target)))
	g.P(`pk, pkv, isNil := m.PrimaryKey()`)
	g.P(`if isNil {`)
	g.P(fmt.Sprintf(`return nil, %s.New("missing primary key")`, g.errPkg.Use()))
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`tx := m.tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx).Where(pk+" = ?", pkv)`, g.daoPkg.Use()))
	g.P()
	g.P(`values := make(map[string]interface{}, 0)`)
	for _, field := range schema.Fields {
		column := toColumnName(field.Name)
		if field == schema.PK {
			continue
		}
		switch field.Type {
		case _point, _map:
			g.P(fmt.Sprintf(`if m.%s != nil {`, field.Name))
			g.P(fmt.Sprintf(`values["%s"] = m.%s`, column, field.Name))
			g.P("}")
		case _slice:
			g.P(fmt.Sprintf(`if len(m.%s) != 0 {`, field.Name))
			g.P(fmt.Sprintf(`values["%s"] = m.%s`, column, field.Name))
			g.P("}")
		case _string:
			g.P(fmt.Sprintf(`if m.%s != "" {`, field.Name))
			g.P(fmt.Sprintf(`values["%s"] = m.%s`, column, field.Name))
			g.P("}")
		default:
			g.P(fmt.Sprintf(`if m.%s != 0 {`, field.Name))
			g.P(fmt.Sprintf(`values["%s"] = m.%s`, column, field.Name))
			g.P("}")
		}
	}
	g.P()
	g.P(`if err := tx.Updates(values).Error; err != nil {`)
	g.P("return nil, err")
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`err := m.tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx).Where(pk+" = ?", pkv).First(m).Error`, g.daoPkg.Use()))
	g.P("if err != nil {")
	g.P(`return nil, err`)
	g.P("}")
	g.P(fmt.Sprintf(`return m.To%s(), nil`, target))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) BatchDelete(ctx %s.Context, soft bool) error {`, source, g.ctxPkg.Use()))
	g.P(`if len(m.exprs) == 0 {`)
	g.P(fmt.Sprintf(`return %s.New("missing conditions")`, g.errPkg.Use()))
	g.P("}")
	g.P()
	g.P(`tx := m.tx.Session(&dao.Session{}).Table(m.TableName()).WithContext(ctx)`)
	g.P()
	g.P(`if soft {`)
	g.P(fmt.Sprintf(`return tx.Clauses(m.exprs...).Updates(map[string]interface{}{"deletion_timestamp": %s.Now().UnixNano()}).Error`, g.timePkg.Use()))
	g.P(`}`)
	g.P(fmt.Sprintf(`return tx.Clauses(m.exprs...).Delete(&%s{}).Error`, source))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) Delete(ctx %s.Context, soft bool) error {`, source, g.ctxPkg.Use()))
	g.P(`pk, pkv, isNil := m.PrimaryKey()`)
	g.P(`if isNil {`)
	g.P(fmt.Sprintf(`return %s.New("missing primary key")`, g.errPkg.Use()))
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`tx := m.tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx)`, g.daoPkg.Use()))
	g.P()
	g.P(`if soft {`)
	g.P(fmt.Sprintf(`return tx.Where(pk+" = ?", pkv).Updates(map[string]interface{}{"deletion_timestamp": %s.Now().UnixNano()}).Error`, g.timePkg.Use()))
	g.P(`}`)
	g.P(fmt.Sprintf(`return tx.Where(pk+" = ?", pkv).Delete(&%s{}).Error`, source))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) Tx(ctx %s.Context) *dao.DB {`, source, g.ctxPkg.Use()))
	g.P(fmt.Sprintf(`return m.tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx).Clauses(m.exprs...)`, g.daoPkg.Use()))
	g.P("}")
	g.P()
}

func (g *dao) buildFieldTypeAndTags(field *generator.FieldDescriptor) (fieldType, []*FieldTag, error) {
	var (
		typ   fieldType
		incre bool
		tags  = make([]*FieldTag, 0)
	)
	tags = append(tags, g.buildJSONTags(field))
	switch field.Proto.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		typ, incre = _float64, false
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		typ, incre = _float32, false
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_ENUM:
		typ, incre = _int32, true
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_INT64:
		typ, incre = _int64, true
	case descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_UINT32:
		typ, incre = _uint32, true
	case descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_UINT64:
		typ, incre = _uint64, true
	case descriptor.FieldDescriptorProto_TYPE_STRING,
		descriptor.FieldDescriptorProto_TYPE_BYTES:
		typ, incre = _string, false
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		typ, incre = _point, false
	default:
		return "", nil, errors.New("invalid field type")
	}

	tags = append(tags, g.buildDaoTags(field, incre))
	return typ, tags, nil
}

func (g *dao) buildJSONTags(field *generator.FieldDescriptor) *FieldTag {
	fTag := &FieldTag{Key: "json", Values: []string{}, Seq: ","}
	fTag.Values = append(fTag.Values, field.Proto.GetJsonName(), "omitempty")
	return fTag
}

func (g *dao) buildDaoTags(field *generator.FieldDescriptor, autoIncr bool) *FieldTag {
	fTag := &FieldTag{Key: "dao", Values: []string{}, Seq: ";"}
	fTag.Values = append(fTag.Values, fmt.Sprintf(`column:%s`, toColumnName(field.Proto.GetName())))
	for _, tag := range g.extractTags(field.Comments) {
		// TODO: parse dao tags
		switch tag.Key {
		case _pk:
			if autoIncr {
				fTag.Values = append(fTag.Values, "autoIncrement")
			}
		}
	}

	return fTag
}

func (g *dao) buildFieldGoType(file *generator.FileDescriptor, field *descriptor.FieldDescriptorProto) (string, error) {
	switch field.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return "float64", nil
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return "float32", nil
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_ENUM:
		return "int32", nil
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_INT64:
		return "int64", nil
	case descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_UINT32:
		return "uint32", nil
	case descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_UINT64:
		return "uint64", nil
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return "string", nil
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return "[]byte", nil
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		subMsg := g.extractMessage(field.GetTypeName())
		if subMsg.Proto.File() == file {
			return "*" + g.wrapPkg(subMsg.Proto.GetName()), nil
		}
		obj := g.gen.ObjectNamed(field.GetTypeName())
		v, ok := g.gen.ImportMap[obj.GoImportPath().String()]
		if !ok {
			v = string(g.gen.AddImport(obj.GoImportPath()))
		}
		return "*" + v + "." + subMsg.Proto.GetName(), nil
	default:
		return "", errors.New("invalid field type")
	}
}

func (g *dao) wrapPkg(pkg string) string {
	if g.gen.OutPut.Load {
		return g.sourcePkg + "." + pkg
	}
	return pkg
}

// extractMessage extract MessageDescriptor by name
func (g *dao) extractMessage(name string) *generator.MessageDescriptor {
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

func (g *dao) extractTags(comments []*generator.Comment) map[string]*Tag {
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

func toTableName(text string) string {
	return schema.NamingStrategy{}.TableName(text)
}

func toColumnName(text string) string {
	return schema.NamingStrategy{}.ColumnName("", text)
}

func toQuoted(text string) string {
	return "`" + text + "`"
}

func (g *dao) isContains(dir, source, typ string) (ok bool) {
	_ = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			return nil
		}
		for _, i := range f.Scope.Objects {
			if i.Kind.String() == "type" && i.Name == typ && source != info.Name() {
				g.P("// ", typ)
				ok = true
				return nil
			}
		}
		return nil
	})
	return
}
