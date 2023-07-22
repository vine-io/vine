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
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/vine-io/vine/cmd/generator"
)

var TagString = "gen"

const (
	// message tags
	// dao generate flag
	_dao    = "dao"
	_table  = "table"
	_object = "object"

	// field tags
	// inline
	_inline = "inline"
	// dao primary key
	_pk         = "primaryKey"
	_deleteAt   = "deleteAt"
	_daoExtract = "daoInject"
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

	schemas     []*Storage
	regTables   map[string]string
	aliasTypes  map[string]string
	aliasFields map[string]*Field

	sourcePkg  string
	ctxPkg     generator.Single
	fmtPkg     generator.Single
	timePkg    generator.Single
	reflectPkg generator.Single
	stringPkg  generator.Single
	errPkg     generator.Single
	DriverPkg  generator.Single
	jsonPkg    generator.Single
	gormPkg    generator.Single
	schemaPkg  generator.Single
	daoPkg     generator.Single
	clausePkg  generator.Single
	runtimePkg generator.Single
	storagePkg generator.Single
}

func New() *dao {
	return &dao{
		schemas:     []*Storage{},
		regTables:   map[string]string{},
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
	g.fmtPkg = g.NewImport("fmt", "fmt")
	g.reflectPkg = g.NewImport("reflect", "reflect")
	g.timePkg = g.NewImport("time", "time")
	g.stringPkg = g.NewImport("strings", "strings")
	g.DriverPkg = g.NewImport("database/sql/driver", "driver")
	g.jsonPkg = g.NewImport("github.com/json-iterator/go", "json")
	g.errPkg = g.NewImport("errors", "errors")
	g.gormPkg = g.NewImport("gorm.io/gorm", "gorm")
	g.schemaPkg = g.NewImport("gorm.io/gorm/schema", "schema")
	g.clausePkg = g.NewImport("gorm.io/gorm/clause", "clause")
	g.daoPkg = g.NewImport("github.com/vine-io/apimachinery/storage/dao", "dao")
	g.runtimePkg = g.NewImport("github.com/vine-io/apimachinery/runtime", "runtime")
	g.storagePkg = g.NewImport("github.com/vine-io/apimachinery/storage", "apistorage")
	if g.gen.OutPut.Load {
		g.sourcePkg = string(g.gen.AddImport(generator.GoImportPath(g.gen.OutPut.SourcePkgPath)))
	}

	for _, msg := range file.Messages() {
		g.wrapStorages(file, msg)
	}

	g.generateRegTables(file)

	aFields := make([]*Field, 0)
	for _, value := range g.aliasFields {
		if value.File.GetName() != file.GetName() {
			continue
		}
		aFields = append(aFields, value)
	}
	sort.Slice(aFields, func(i, j int) bool { return aFields[i].Num < aFields[j].Num })
	for _, value := range aFields {
		if file.GetOptions().GetGoPackage() != "" && g.gen.OutPut.Out != "" {
			f := strings.TrimSuffix(filepath.Base(file.GetName()), ".proto") + ".pb.dao.go"
			if g.isContains(filepath.Join(build.Default.GOPATH, "src", g.gen.OutPut.Out), f, "type", value.Alias) {
				continue
			}
		}
		g.generateAliasField(file, value)
	}

	for _, item := range g.schemas {
		if item.Desc.Proto.File().GetName() != file.GetName() {
			continue
		}
		g.generateStorage(file, item)
	}
}

func (g *dao) wrapStorages(file *generator.FileDescriptor, msg *generator.MessageDescriptor) {
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

	s := &Storage{
		Name:    msg.Proto.GetName() + "Storage",
		Desc:    msg,
		MFields: map[string]*Field{},
		Table:   table,
	}

	if _, ok := tags[_object]; ok {
		g.regTables[msg.Proto.GetName()] = s.Name
		s.Deep = true
	}

	n := 0
	g.buildFields(file, msg, s, &n)
	if s.PK == nil {
		g.gen.Fail(fmt.Sprintf(`Message:%s missing primary key`, msg.Proto.GetName()))
	}
	if s.DeleteAt == nil {
		g.gen.Fail(fmt.Sprintf(`Message:%s missing deleteAt`, msg.Proto.GetName()))
	}

	s.Fields = make([]*Field, 0)
	for _, item := range s.MFields {
		s.Fields = append(s.Fields, item)
	}
	sort.Slice(s.Fields, func(i, j int) bool { return s.Fields[i].Num < s.Fields[j].Num })
	g.schemas = append(g.schemas, s)
}

func (g *dao) buildFields(file *generator.FileDescriptor, m *generator.MessageDescriptor, s *Storage, n *int) {
	fileName := strings.ReplaceAll(filepath.Base(file.GetName()), ".", "_")
	for _, item := range m.Fields {
		fTags := g.extractTags(item.Comments)

		_, isInline := fTags[_inline]
		if isInline && item.Proto.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			if item.Proto.GetName() == "typeMeta" {
				s.TypeMeta = true
			}
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
			alias := generator.CamelCaseSlice([]string{fileName, m.Proto.GetName(), item.Proto.GetName()})
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
			if pkExists && s.PK == nil && (strings.HasSuffix(strings.ToLower(field.Name), "id")) {
				s.PK = field
				for _, tag := range field.Tags {
					if tag.Key == "dao" {
						tag.Values = append(tag.Values, "primaryKey")
					}
				}
			}

			_, deleteAt := fTags[_deleteAt]
			if deleteAt && s.DeleteAt == nil {
				s.DeleteAt = field
			}
		}

		*n += 1
		s.MFields[field.Name] = field
	}
}

func (g *dao) generateRegTables(file *generator.FileDescriptor) {
	if len(g.regTables) == 0 {
		return
	}

	pkg := g.gen.OutPut.Package
	if pkg == "" {
		pkg = file.GetPackage()
	}
	out := g.gen.OutPut.Out
	if out == "" {
		pwd, _ := os.Getwd()
		out = filepath.Join(pwd, filepath.Dir(file.GetName()))
	}
	register := filepath.Join(out, "register.go")
	stat, _ := os.Stat(register)
	if stat == nil {

		tpl := fmt.Sprintf(`// Code generated by proto-gen-dao.
	
	package %s

import (
	"github.com/vine-io/apimachinery/runtime"
	"github.com/vine-io/apimachinery/schema"
	"github.com/vine-io/apimachinery/storage"
)

// GroupName is the group name for this API
const GroupName = ""

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

var (
	SchemaBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemaBuilder.AddToScheme
	sets          = make([]runtime.Object, 0)
)

var (
	FactoryBuilder = storage.NewFactoryBuilder(addKnownFactory)
	AddToBuilder   = FactoryBuilder.AddToFactory
	storageSet     = make([]storage.Storage, 0)
)

func addKnownFactory(f storage.Factory) error {
	return f.AddKnownStorages(SchemeGroupVersion, storageSet...)
}

func addKnownTypes(scheme runtime.Scheme) error {
	return scheme.AddKnownTypes(SchemeGroupVersion, sets...)
}`, pkg)

		_ = ioutil.WriteFile(register, []byte(tpl), 0755)
	}

	g.P("// ", out)
	g.P("func init() {")
	g.P("storageSet = append(storageSet,")
	for _, table := range g.regTables {
		g.P(fmt.Sprintf(`&%s{},`, table))
	}
	g.P(")")
	g.P("}")
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
		if alias != typ {
			g.P(fmt.Sprintf(`// %s the alias of %s`, alias, typ))
			g.P(fmt.Sprintf("type %s %s", alias, typ))
		}
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
	g.P(fmt.Sprintf(`return %s.New(%s.Sprint("Failed to unmarshal JSONB value:", value))`, g.errPkg.Use(), g.fmtPkg.Use()))
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`return %s.Unmarshal(bytes, &m)`, g.jsonPkg.Use()))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) GormDBDataType(db *%s.DB, field *%s.Field) string {`, alias, g.gormPkg.Use(), g.schemaPkg.Use()))
	g.P(fmt.Sprintf(`return %s.GetGormDBDataType(db, field)`, g.daoPkg.Use()))
	g.P("}")
	g.P()
}

func (g *dao) generateStorage(file *generator.FileDescriptor, schema *Storage) {
	g.generateEntityIOMethods(file, schema)
	g.generateStorageEntity(file, schema)
	g.generateStorageMethods(file, schema)
}

func (g *dao) generateEntityIOMethods(file *generator.FileDescriptor, schema *Storage) {
	pname := schema.Desc.Proto.GetName()

	for _, field := range schema.Fields {
		switch field.Type {
		case _float32, _float64, _int32, _int64, _uint32, _uint64, _string, _bool:
			g.P(fmt.Sprintf(`func (m *%s) Apply%s(in %s) *%s {`, pname, field.Name, field.Type, pname))
			g.P(fmt.Sprintf(`m.%s = in`, field.Name))
			g.P(`return m`)
			g.P("}")
			g.P()
		case _map:
			key, value := field.Map.Key, field.Map.Value
			keyString, _ := g.buildFieldGoType(file, key)
			valueString, _ := g.buildFieldGoType(file, value)
			g.P(fmt.Sprintf(`func (m *%s) Apply%s(in map[%s]%s) *%s {`, pname, field.Name, keyString, valueString, pname))
			g.P(fmt.Sprintf(`m.%s = in`, field.Name))
			g.P(`return m`)
			g.P("}")
			g.P()
			g.P(fmt.Sprintf(`func (m *%s) Put%s(k %s, v %s) *%s {`, pname, field.Name, keyString, valueString, pname))
			g.P(fmt.Sprintf(`if len(m.%s) == 0 {`, field.Name))
			g.P(fmt.Sprintf(`m.%s = map[%s]%s{}`, field.Name, keyString, valueString))
			g.P(`}`)
			g.P(fmt.Sprintf(`m.%s[k] = v`, field.Name))
			g.P(`return m`)
			g.P("}")
			g.P()
			g.P(fmt.Sprintf(`func (m *%s) Remove%s(k %s) *%s {`, pname, field.Name, keyString, pname))
			g.P(fmt.Sprintf(`if len(m.%s) == 0 {`, field.Name))
			g.P(`return m`)
			g.P(`}`)
			g.P(fmt.Sprintf(`delete(m.%s, k)`, field.Name))
			g.P(`return m`)
			g.P("}")
			g.P()
		case _slice:
			typ, _ := g.buildFieldGoType(file, field.Slice)
			g.P(fmt.Sprintf(`func (m *%s) Apply%s(in []%s) *%s {`, pname, field.Name, typ, pname))
			g.P(fmt.Sprintf(`m.%s = in`, field.Name))
			g.P(`return m`)
			g.P("}")
			g.P()
			g.P(fmt.Sprintf(`func (m *%s) Add%s(in %s) *%s {`, pname, field.Name, typ, pname))
			g.P(fmt.Sprintf(`if len(m.%s) == 0 {`, field.Name))
			g.P(fmt.Sprintf(`m.%s = []%s{}`, field.Name, typ))
			g.P(`}`)
			g.P(fmt.Sprintf(`m.%s = append(m.%s, in)`, field.Name, field.Name))
			g.P(`return m`)
			g.P("}")
			g.P()
			g.P(fmt.Sprintf(`func (m *%s) Filter%s(fn func(item %s) bool) *%s {`, pname, field.Name, typ, pname))
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
			g.P(fmt.Sprintf(`func (m *%s) Remove%s(index int) *%s {`, pname, field.Name, pname))
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
			g.P(fmt.Sprintf(`func (m *%s) Apply%s(in *%s) *%s {`, pname, field.Name, typ, pname))
			g.P(fmt.Sprintf(`m.%s = in`, field.Name))
			g.P(`return m`)
			g.P("}")
			g.P()
		}
	}

	g.P("type XX", pname, " struct {")
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
	g.P("}")

	g.P(fmt.Sprintf(`func From%s(in *%s) *XX%s {`, pname, g.wrapPkg(pname), pname))
	g.P(fmt.Sprintf(`out := new(XX%s)`, pname))
	g.P("if in == nil { return out }")
	for _, field := range schema.Fields {
		switch field.Type {
		case _map:
			key, value := field.Map.Key, field.Map.Value
			keyString, _ := g.buildFieldGoType(file, key)
			valueString, _ := g.buildFieldGoType(file, value)
			g.P(fmt.Sprintf(`if out.%s == nil { out.%s = make(map[%s]%s) }`, field.Name, field.Name, keyString, valueString))
			g.P(fmt.Sprintf(`for k, v := range in.%s {`, field.Name))
			g.P(fmt.Sprintf(`out.%s[k] = v`, field.Name))
			g.P("}")
		case _slice:
			typ, _ := g.buildFieldGoType(file, field.Desc.Proto)
			g.P(fmt.Sprintf(`out.%s = make([]%s, len(in.%s))`, field.Name, typ, field.Name))
			g.P(fmt.Sprintf(`for i, item := range in.%s {`, field.Name))
			g.P(fmt.Sprintf(`out.%s[i] = item`, field.Name))
			g.P("}")
		case _float32, _float64, _int32, _int64, _uint32, _uint64:
			g.P(fmt.Sprintf(`if in.%s != 0 {`, field.Name))
			if field.Desc.Proto.GetType() == descriptor.FieldDescriptorProto_TYPE_ENUM {
				g.P(fmt.Sprintf(`out.%s = int32(in.%s)`, field.Name, field.Name))
			} else {
				g.P(fmt.Sprintf(`out.%s = in.%s`, field.Name, field.Name))
			}
			g.P("}")
		case _bool:
			continue
		case _string:
			if field.Desc.Proto.GetType() == descriptor.FieldDescriptorProto_TYPE_BYTES {
				g.P(fmt.Sprintf(`if in.%s != nil {`, field.Name))
				g.P(fmt.Sprintf(`out.%s = string(in.%s)`, field.Name, field.Name))
			} else {
				g.P(fmt.Sprintf(`if in.%s != "" {`, field.Name))
				g.P(fmt.Sprintf(`out.%s = in.%s`, field.Name, field.Name))
			}
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

	g.P(fmt.Sprintf(`func (m *XX%s) To%s() *%s {`, pname, pname, g.wrapPkg(pname)))
	g.P(fmt.Sprintf(`out := new(%s)`, g.wrapPkg(pname)))
	for _, field := range schema.Fields {
		switch field.Type {
		case _map:
			key, value := field.Map.Key, field.Map.Value
			keyString, _ := g.buildFieldGoType(file, key)
			valueString, _ := g.buildFieldGoType(file, value)
			g.P(fmt.Sprintf(`if out.%s == nil { out.%s = make(map[%s]%s) }`, field.Name, field.Name, keyString, valueString))
			g.P(fmt.Sprintf(`for k, v := range m.%s {`, field.Name))
			g.P(fmt.Sprintf(`out.%s[k] = v`, field.Name))
			g.P("}")
		case _slice:
			typ, _ := g.buildFieldGoType(file, field.Desc.Proto)
			g.P(fmt.Sprintf(`out.%s = make([]%s, len(m.%s))`, field.Name, typ, field.Name))
			g.P(fmt.Sprintf(`for i, item := range m.%s {`, field.Name))
			g.P(fmt.Sprintf(`out.%s[i] = item`, field.Name))
			g.P("}")
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

	g.P(fmt.Sprintf("func (m XX%s) PrimaryKey() (string, any, bool) {", pname))
	fpk := schema.PK
	if fpk.Desc.Proto.GetType() == descriptor.FieldDescriptorProto_TYPE_STRING {
		g.P(fmt.Sprintf(`return "%s", m.%s, m.%s == ""`, toColumnName(fpk.Desc.Proto.GetName()), fpk.Name, fpk.Name))
	} else {
		g.P(fmt.Sprintf(`return "%s", m.%s, m.%s == 0`, toColumnName(fpk.Desc.Proto.GetName()), fpk.Name, fpk.Name))
	}
	g.P("}")
	g.P()

	g.P(fmt.Sprintf("func (m XX%s) DeleteAt() string {", pname))
	fDelete := schema.DeleteAt
	g.P(fmt.Sprintf(`return "%s"`, toColumnName(fDelete.Desc.Proto.GetName())))
	g.P("}")
	g.P()

	g.P("func (XX", pname, ") TableName() string {")
	g.P(`return "`, schema.Table, `"`)
	g.P("}")
	g.P()
}

func (g *dao) generateStorageEntity(file *generator.FileDescriptor, schema *Storage) {

	sname, pname := schema.Name, schema.Desc.Proto.GetName()

	g.P(fmt.Sprintf(`// %s the Storage for %s`, sname, pname))
	g.P("type ", sname, " struct {")
	g.P(fmt.Sprintf(`%s.EmptyHook`, g.storagePkg.Use()))
	g.P(fmt.Sprintf(`tx *gorm.DB %s`, toQuoted(`json:"-" gorm:"-"`)))
	g.P("joins []string ", toQuoted(`json:"-" gorm:"-"`))
	g.P("ptr *", pname)
	g.P(fmt.Sprintf(`exprs []%s.Expression %s`, g.clausePkg.Use(), toQuoted(`json:"-" gorm:"-"`)))
	g.P()
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func New%s(tx *%s.DB, in *%s) *%s {
	return &%s{tx: tx, joins: []string{}, exprs: make([]%s.Expression, 0), ptr: in}
}`, sname, g.gormPkg.Use(), pname, sname, sname, g.clausePkg.Use()))
}

func (g *dao) generateStorageMethods(file *generator.FileDescriptor, schema *Storage) {
	sname, pname := schema.Name, schema.Desc.Proto.GetName()

	g.P(fmt.Sprintf(`func (s *%s) AutoMigrate(tx *%s.DB) error {`, sname, g.gormPkg.Use()))
	g.P(fmt.Sprintf("return tx.Migrator().AutoMigrate(From%s(s.ptr))", pname))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf("func (s %s) Target() %s.Type {", sname, g.reflectPkg.Use()))
	g.P(fmt.Sprintf(`return reflect.TypeOf(new(%s))`, g.wrapPkg(pname)))
	g.P("}")
	g.P()

	if schema.Deep {
		g.P(fmt.Sprintf(`func (s *%s) Load(tx *%s.DB, in %s.Object) error {
	return s.XXLoad(tx, in.(*%s))
}`, sname, g.gormPkg.Use(), g.runtimePkg.Use(), g.wrapPkg(pname)))

		g.P(fmt.Sprintf(`func (s *%s) FindPage(ctx %s.Context, page, size int32) (%s.Object, error) {
	items, total, err := s.XXFindPage(ctx, page, size)
	if err != nil {
		return nil, err
	}

	out := &%sList{}
	out.Page = page
	out.Size = size
	out.Total = total
	out.Items = items
	if len(items) > 0 {
		out.TypeMeta = items[0].TypeMeta
	}
	return out, nil
}`, sname, g.ctxPkg.Use(), g.runtimePkg.Use(), pname))
		g.P()

		g.P(fmt.Sprintf(`func (s *%s) FindAll(ctx %s.Context) (%s.Object, error) {
			items, err := s.XXFindAll(ctx)
			if err != nil {
				return nil, err
			}

			out := &%sList{}
			out.Total = int64(len(items))
			out.Items = items
			if len(items) > 0 {
				out.TypeMeta = items[0].TypeMeta
			}
			return out, nil
		}`, sname, g.ctxPkg.Use(), g.runtimePkg.Use(), pname))
		g.P()

		g.P(fmt.Sprintf(`func (s *%s) FindPk(ctx %s.Context, id any) (%s.Object, error) {
			return s.XXFindPk(ctx, id)
		}`, sname, g.ctxPkg.Use(), g.runtimePkg.Use()))
		g.P()

		g.P(fmt.Sprintf(`func (s *%s) FindOne(ctx %s.Context) (%s.Object, error) {
			return s.XXFindOne(ctx)
		}`, sname, g.ctxPkg.Use(), g.runtimePkg.Use()))
		g.P()

		g.P(fmt.Sprintf(`func (s *%s) Cond(exprs ...%s.Expression) %s.Storage {
			return s.XXCond(exprs...)
		}`, sname, g.clausePkg.Use(), g.storagePkg.Use()))
		g.P()

		g.P(fmt.Sprintf(`func (s *%s) Create(ctx %s.Context) (%s.Object, error) {
			return s.XXCreate(ctx)
		}`, sname, g.ctxPkg.Use(), g.runtimePkg.Use()))
		g.P()

		g.P(fmt.Sprintf(`func (s *%s) Updates(ctx %s.Context) (%s.Object, error) {
			return s.XXUpdates(ctx)
		}`, sname, g.ctxPkg.Use(), g.runtimePkg.Use()))
		g.P()

		g.P(fmt.Sprintf(`func (s *%s) Delete(ctx %s.Context, soft bool) error {
			return s.XXDelete(ctx, soft)
		}`, sname, g.ctxPkg.Use()))
		g.P()
	}

	g.P(fmt.Sprintf(`func (s *%s) XXLoad(tx *%s.DB, in *%s) error {`, sname, g.gormPkg.Use(), g.wrapPkg(pname)))
	g.P(fmt.Sprintf(`s.tx = tx`))
	g.P("s.ptr = in")
	g.P("return nil")
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (s *%s) XXFindPage(ctx %s.Context, page, size int32) ([]*%s, int64, error) {`, sname, g.ctxPkg.Use(), g.wrapPkg(pname)))
	g.P(fmt.Sprintf("m := From%s(s.ptr)", pname))
	g.P(`pk, _, _ := m.PrimaryKey()`)
	g.P()
	g.P(`s.exprs = append(s.exprs,`)
	g.P(fmt.Sprintf(`%s.OrderBy{Columns: []%s.OrderByColumn{{Column: %s.Column{Table: m.TableName(), Name: pk}, Desc: true}}},`, g.clausePkg.Use(), g.clausePkg.Use(), g.clausePkg.Use()))
	g.P(fmt.Sprintf(`%s.Cond().Build(m.DeleteAt(), 0),`, g.daoPkg.Use()))
	g.P(`)`)
	g.P()
	g.P(`total, err := s.Count(ctx)`)
	g.P(`if err != nil {`)
	g.P(`return nil, 0, err`)
	g.P("}")
	g.P()
	g.P("limit := int(size)")
	g.P(fmt.Sprintf(`s.exprs = append(s.exprs, %s.Limit{Offset: int((page - 1) * size), Limit: &limit})`, g.clausePkg.Use()))
	g.P(`data, err := s.findAll(ctx)`)
	g.P(`if err != nil {`)
	g.P(`return nil, 0, err`)
	g.P("}")
	g.P()
	g.P(`return data, total, nil`)
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (s *%s) XXFindAll(ctx %s.Context) ([]*%s, error) {`, sname, g.ctxPkg.Use(), g.wrapPkg(pname)))
	g.P(fmt.Sprintf("m := From%s(s.ptr)", pname))
	g.P(fmt.Sprintf(`s.exprs = append(s.exprs, %s.Cond().Build(m.DeleteAt(), 0))`, g.daoPkg.Use()))
	g.P(`return s.findAll(ctx)`)
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (s *%s) XXFindPureAll(ctx %s.Context) ([]*%s, error) {`, sname, g.ctxPkg.Use(), g.wrapPkg(pname)))
	g.P(`return s.findAll(ctx)`)
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (s *%s) findAll(ctx %s.Context) ([]*%s, error) {`, sname, g.ctxPkg.Use(), g.wrapPkg(pname)))
	g.P(fmt.Sprintf(`dest := make([]*XX%s, 0)`, pname))
	g.P(fmt.Sprintf("m := From%s(s.ptr)", pname))
	g.P("tx := s.tx")
	g.P(fmt.Sprintf(`tx1 := tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx)`, g.gormPkg.Use()))
	g.P()
	g.P(`clauses := append(s.extractClauses(tx1), s.exprs...)`)
	g.P("for _, item := range s.joins { tx1 = tx1.Joins(item) }")
	g.P(`if err := tx1.Clauses(clauses...).Find(&dest).Error; err != nil {`)
	g.P("return nil, err")
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`outs := make([]*%s, len(dest))`, g.wrapPkg(pname)))
	g.P(`for i := range dest {`)
	g.P(fmt.Sprintf(`outs[i] = dest[i].To%s()`, pname))
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`tx2 := tx.Session(&gorm.Session{}).WithContext(ctx)
	if err := s.PostList(ctx, tx2, outs); err != nil {
		return nil, err
	}`))
	g.P()
	g.P(fmt.Sprintf(`return outs, nil`))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (s *%s) Count(ctx %s.Context) (total int64, err error) {`, sname, g.ctxPkg.Use()))
	g.P(fmt.Sprintf("m := From%s(s.ptr)", pname))
	g.P(fmt.Sprintf(`tx := s.tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx)`, g.gormPkg.Use()))
	g.P()
	g.P(fmt.Sprintf(`clauses := append(s.extractClauses(tx), %s.Cond().Build(m.DeleteAt(), 0))`, g.daoPkg.Use()))
	g.P(`clauses = append(clauses, s.exprs...)`)
	g.P()
	g.P(`err = tx.Clauses(clauses...).Count(&total).Error`)
	g.P(`return`)
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (s *%s) XXFindPk(ctx %s.Context, id any) (*%s, error) {`, sname, g.ctxPkg.Use(), g.wrapPkg(pname)))
	g.P(fmt.Sprintf("m := From%s(s.ptr)", pname))
	g.P("tx := s.tx")
	g.P(`pk, _, _ := m.PrimaryKey()`)
	g.P(fmt.Sprintf(`tx1 := tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx)`, g.gormPkg.Use()))
	g.P()
	g.P(fmt.Sprintf(`if err := tx1.Where(pk+" = ?", id).First(&m).Error; err != nil { return nil, err }`))
	g.P(fmt.Sprintf(`out := m.To%s()`, pname))
	g.P()
	g.P(fmt.Sprintf(`tx2 := tx.Session(&gorm.Session{}).WithContext(ctx)
	if err := s.PostGet(ctx, tx2, out); err != nil {
		return nil, err
	}`))
	g.P()
	g.P("return out, nil")
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (s *%s) XXFindOne(ctx %s.Context) (*%s, error) {`, sname, g.ctxPkg.Use(), g.wrapPkg(pname)))
	g.P(fmt.Sprintf("m := From%s(s.ptr)", pname))
	g.P("tx := s.tx")
	g.P(fmt.Sprintf(`tx1 := tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx)`, g.gormPkg.Use()))
	g.P()
	g.P(fmt.Sprintf(`clauses := append(s.extractClauses(tx1), %s.Cond().Build(m.DeleteAt(), 0))`, g.daoPkg.Use()))
	g.P(`clauses = append(clauses, s.exprs...)`)
	g.P()
	g.P("for _, item := range s.joins { tx1 = tx1.Joins(item) }")
	g.P(`if err := tx1.Clauses(clauses...).First(&m).Error; err != nil {`)
	g.P("return nil, err")
	g.P("}")
	g.P(fmt.Sprintf(`out := m.To%s()`, pname))
	g.P()
	g.P(fmt.Sprintf(`tx2 := tx.Session(&gorm.Session{}).WithContext(ctx)
	if err := s.PostGet(ctx, tx2, out); err != nil {
		return nil, err
	}`))
	g.P()
	g.P("return out, nil")
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (s *%s) XXFindPureOne(ctx %s.Context) (*%s, error) {`, sname, g.ctxPkg.Use(), g.wrapPkg(pname)))
	g.P(fmt.Sprintf("m := From%s(s.ptr)", pname))
	g.P("tx := s.tx")
	g.P(fmt.Sprintf(`tx1 := tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx)`, g.gormPkg.Use()))
	g.P()
	g.P(`if err := tx1.Clauses(s.exprs...).First(&m).Error; err != nil {`)
	g.P("return nil, err")
	g.P("}")
	g.P(fmt.Sprintf(`out := m.To%s()`, pname))
	g.P()
	g.P(fmt.Sprintf(`tx2 := tx.Session(&gorm.Session{}).WithContext(ctx)
	if err := s.PostGet(ctx, tx2, out); err != nil {
		return nil, err
	}`))
	g.P()
	g.P("return out, nil")
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (s *%s) XXCond(exprs ...%s.Expression) *%s {`, sname, g.clausePkg.Use(), sname))
	g.P(`s.exprs = append(s.exprs, exprs...)`)
	g.P(`return s`)
	g.P(`}`)
	g.P()

	g.P(fmt.Sprintf(`func (s *%s) extractClauses(tx *%s.DB) []%s.Expression {`, sname, g.gormPkg.Use(), g.clausePkg.Use()))
	g.P(fmt.Sprintf("m := From%s(s.ptr)", pname))
	g.P(`exprs := make([]clause.Expression, 0)`)
	for _, field := range schema.Fields {
		column := toColumnName(field.Name)
		if schema.TypeMeta {
			if field.Name == "Kind" || field.Name == "ApiVersion" {
				continue
			}
		}
		switch field.Type {
		case _map:
			g.P(fmt.Sprintf(`if m.%s != nil {`, field.Name))
			g.P(fmt.Sprintf(`for k, v := range m.%s {`, field.Name))
			switch field.Map.Key.GetType() {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				g.P(fmt.Sprintf(`exprs = append(exprs, %s.JSONQuery("%s").Equals(v, k))`, g.daoPkg.Use(), column))
			case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_INT64:
				g.P(fmt.Sprintf(`exprs = append(exprs, %s.JSONQuery("%s").Equals(v, %s))`, g.daoPkg.Use(), column, `fmt.Sprintf("%d", k)`))
			}
			g.P("}")
			g.P("}")
		case _point:
			g.P(fmt.Sprintf(`if m.%s != nil {`, field.Name))
			g.P(fmt.Sprintf(`for k, v := range dao.FieldPatch(m.%s) {`, field.Name))
			g.P(`if v == nil {`)
			g.P(fmt.Sprintf(`exprs = append(exprs, %s.JSONQuery("%s").HasKey(%s.Split(k, ".")...))`, g.daoPkg.Use(), column, g.stringPkg.Use()))
			g.P(`} else {`)
			g.P(fmt.Sprintf(`exprs = append(exprs, %s.JSONQuery("%s").Equals(v, %s.Split(k, ".")...))`, g.daoPkg.Use(), column, g.stringPkg.Use()))
			g.P("}")
			g.P("}")
			g.P("}")
		case _slice:
			g.P(fmt.Sprintf(`if len(m.%s) != 0 {`, field.Name))
			g.P(fmt.Sprintf(`for _, item := range m.%s {`, field.Name))
			switch field.Slice.GetType() {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				g.P(fmt.Sprintf(`expr, query := %s.JSONQuery("%s").Contains(tx, item)`, g.daoPkg.Use(), column))
				g.P("s.joins = append(s.joins, query)")
				g.P(fmt.Sprintf(`exprs = append(exprs, expr)`))
			case descriptor.FieldDescriptorProto_TYPE_UINT32,
				descriptor.FieldDescriptorProto_TYPE_UINT64,
				descriptor.FieldDescriptorProto_TYPE_INT32,
				descriptor.FieldDescriptorProto_TYPE_INT64,
				descriptor.FieldDescriptorProto_TYPE_SFIXED32,
				descriptor.FieldDescriptorProto_TYPE_SFIXED64,
				descriptor.FieldDescriptorProto_TYPE_FIXED32,
				descriptor.FieldDescriptorProto_TYPE_FIXED64:
				g.P(fmt.Sprintf(`expr, query := %s.JSONQuery("%s").Contains(tx, item)`, g.daoPkg.Use(), column))
				g.P("s.joins = append(s.joins, query)")
				g.P(fmt.Sprintf(`exprs = append(exprs, expr)`))
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				g.P(`for k, v := range dao.FieldPatch(item) {`)
				g.P(`if v != nil {`)
				g.P(fmt.Sprintf(`expr, query := %s.JSONQuery("%s").Contains(tx, v, %s.Split(k, ".")...)`, g.daoPkg.Use(), column, g.stringPkg.Use()))
				g.P("s.joins = append(s.joins, query)")
				g.P(fmt.Sprintf(`exprs = append(exprs, expr)`))
				g.P(`}`)
				g.P("}")
			}
			g.P("}")
			g.P("}")
		case _string:
			g.P(fmt.Sprintf(`if m.%s != "" {`, field.Name))
			g.P(fmt.Sprintf(`exprs = append(exprs, %s.Cond().Op(%s.ParseOp(m.%s)).Build("%s", m.%s))`, g.daoPkg.Use(), g.daoPkg.Use(), field.Name, column, field.Name))
			g.P("}")
		case _bool:
			continue
		default:
			g.P(fmt.Sprintf(`if m.%s != 0 {`, field.Name))
			g.P(fmt.Sprintf(`exprs = append(exprs,  %s.Cond().Build("%s", m.%s))`, g.daoPkg.Use(), column, field.Name))
			g.P("}")
		}
	}
	g.P()
	g.P(`return exprs`)
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (s *%s) XXCreate(ctx %s.Context) (*%s, error) {`, sname, g.ctxPkg.Use(), g.wrapPkg(pname)))
	g.P(fmt.Sprintf(`if err := s.PreCreate(ctx, s.tx, s.ptr); err != nil {
		return nil, err
	}`))
	g.P()
	g.P(fmt.Sprintf("m := From%s(s.ptr)", pname))
	g.P("tx := s.tx")
	g.P(fmt.Sprintf(`tx1 := tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx)`, g.gormPkg.Use()))
	g.P()
	g.P(`if err := tx1.Create(&m).Error; err != nil {`)
	g.P("return nil, err")
	g.P("}")
	g.P(fmt.Sprintf(`out := m.To%s()`, pname))
	g.P()
	g.P(fmt.Sprintf(`tx2 := tx.Session(&gorm.Session{}).WithContext(ctx)
	if err := s.PostCreate(ctx, tx2, out); err != nil {
		return nil, err
	}

	return out, nil`))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (s *%s) XXUpdates(ctx %s.Context) (*%s, error) {`, sname, g.ctxPkg.Use(), g.wrapPkg(pname)))
	g.P(fmt.Sprintf(`if err := s.PreUpdate(ctx, s.tx, s.ptr); err != nil {
		return nil, err
	}`))
	g.P()
	g.P(fmt.Sprintf("m := From%s(s.ptr)", pname))
	g.P(`pk, pkv, isNil := m.PrimaryKey()`)
	g.P(`if isNil {`)
	g.P(fmt.Sprintf(`return nil, %s.New("missing primary key")`, g.errPkg.Use()))
	g.P("}")
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
		case _bool:
			continue
		default:
			g.P(fmt.Sprintf(`if m.%s != 0 {`, field.Name))
			g.P(fmt.Sprintf(`values["%s"] = m.%s`, column, field.Name))
			g.P("}")
		}
	}
	g.P()
	g.P("tx := s.tx")
	g.P(fmt.Sprintf(`tx1 := tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx).Where(pk+" = ?", pkv)`, g.gormPkg.Use()))
	g.P()
	g.P(`if err := tx1.Updates(values).Error; err != nil {`)
	g.P("return nil, err")
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`if err := tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx).Where(pk+" = ?", pkv).First(&m).Error; err != nil {`, g.gormPkg.Use()))
	g.P(`return nil, err`)
	g.P("}")
	g.P(fmt.Sprintf(`out := m.To%s()`, pname))
	g.P(fmt.Sprintf(`tx2 := tx.Session(&gorm.Session{}).WithContext(ctx)
	if err := s.PostUpdate(ctx, tx2, out); err != nil {
		return nil, err
	}

	return out, nil`))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (s *%s) XXDelete(ctx %s.Context, soft bool) error {`, sname, g.ctxPkg.Use()))
	g.P(fmt.Sprintf(`if err := s.PreDelete(ctx, s.tx, s.ptr); err != nil {
		return err
	}`))
	g.P()
	g.P(fmt.Sprintf("m := From%s(s.ptr)", pname))
	g.P(`pk, pkv, isNil := m.PrimaryKey()`)
	g.P(`if isNil {`)
	g.P(fmt.Sprintf(`return %s.New("missing primary key")`, g.errPkg.Use()))
	g.P("}")
	g.P()
	g.P("tx := s.tx")
	g.P(fmt.Sprintf(`tx1 := tx.Session(&%s.Session{}).Table(m.TableName()).WithContext(ctx)`, g.gormPkg.Use()))
	g.P()
	g.P(`if soft {`)
	g.P(fmt.Sprintf(`deleteAt := m.DeleteAt()`))
	g.P(fmt.Sprintf(`if err := tx1.Where(pk+" = ?", pkv).Updates(map[string]interface{}{deleteAt: %s.Now().Unix()}).Error; err != nil {`, g.timePkg.Use()))
	g.P("return err")
	g.P("}")
	g.P(fmt.Sprintf(`} else if err := tx1.Where(pk+" = ?", pkv).Delete(&%s{}).Error; err != nil {`, sname))
	g.P("return err")
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`out := m.To%s()
	tx2 := tx.Session(&gorm.Session{}).WithContext(ctx)
	if err := s.PostDelete(ctx, tx2, out); err != nil {
		return err
	}`, pname))
	g.P()
	g.P("return nil")
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
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		typ, incre = _bool, false
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
	fTag := &FieldTag{Key: "gorm", Values: []string{}, Seq: ";"}
	fTag.Values = append(fTag.Values, fmt.Sprintf(`column:%s`, toColumnName(field.Proto.GetName())))
	if field.Proto.IsRepeated() || field.Proto.IsMessage() {
		fTag.Values = append(fTag.Values, "serializer:json")
	}
	for _, tag := range g.extractTags(field.Comments) {
		// TODO: parse dao tags
		switch tag.Key {
		case _pk:
			if autoIncr {
				fTag.Values = append(fTag.Values, "autoIncrement")
			}
		case _daoExtract:
			fTag.Values = append(fTag.Values, strings.ReplaceAll(tag.Value, `"`, ""))
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
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return "bool", nil
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
	return NamingStrategy{}.TableName(text)
}

func toColumnName(text string) string {
	return NamingStrategy{}.ColumnName("", text)
}

func toQuoted(text string) string {
	return "`" + text + "`"
}

func (g *dao) isContains(dir, source, kind, name string) (ok bool) {
	_ = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if ok || err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			return nil
		}
		obj := f.Scope.Lookup(name)
		if obj == nil {
			return nil
		}
		if obj.Kind.String() == kind && source != info.Name() {
			ok = true
			return nil
		}
		return nil
	})
	return
}
