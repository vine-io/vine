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
	"errors"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

	"github.com/lack-io/vine/cmd/generator"
	"github.com/lack-io/vine/service/dao/schema"
)

var TagString = "dao"

const (
	// dao generate flag
	_generate = "generate"
	// dao primary key
	_pk = "PK"
	// dao soft delete field
	_sd = "SD"
)

type Tag struct {
	Key   string
	Value string
}

// Paths for packages used by code generated in this file,
// relative to the import_prefix of the generator.Generator.
const (
	ctxPkgPath    = "context"
	timePkgPath   = "time"
	stringPkgPath = "strings"
	errPkgPath    = "errors"
	driverPkgPath = "database/sql/driver"
	jsonPkgPath   = "github.com/json-iterator/go"
	daoPkgPath    = "github.com/lack-io/vine/service/dao"
	clausePkgPath = "github.com/lack-io/vine/service/dao/clause"
)

// dao is an implementation of the Go protocol buffer compiler's
// plugin architecture. It generates bindings for dao support.
type dao struct {
	gen *generator.Generator

	schemas     []*Schema
	aliasTypes  map[string]string
	aliasFields map[string]*Field
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

// The names for packages imported in the generated code.
// They may vary from the final path component of the import path
// if the name is used by other packages.
var (
	ctxPkg     string
	timePkg    string
	stringPkg  string
	errPkg     string
	DriverPkg  string
	jsonPkg    string
	daoPkg     string
	clausePkg  string
	sourcePkg  string
	pkgImports map[generator.GoPackageName]bool
)

// Init initializes the plugin.
func (g *dao) Init(gen *generator.Generator) {
	g.gen = gen
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

// GenerateImports generates the import declaration for this file.
func (g *dao) GenerateImports(file *generator.FileDescriptor, imports map[generator.GoImportPath]generator.GoPackageName) {
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

// Generate generates code for the services in the given file.
func (g *dao) Generate(file *generator.FileDescriptor) {
	if len(file.Comments()) == 0 {
		return
	}

	ctxPkg = string(g.gen.AddImport(ctxPkgPath))
	timePkg = string(g.gen.AddImport(timePkgPath))
	stringPkg = string(g.gen.AddImport(stringPkgPath))
	DriverPkg = string(g.gen.AddImport(driverPkgPath))
	errPkg = string(g.gen.AddImport(errPkgPath))
	jsonPkg = string(g.gen.AddImport(jsonPkgPath))
	daoPkg = string(g.gen.AddImport(daoPkgPath))
	clausePkg = string(g.gen.AddImport(clausePkgPath))
	if g.gen.OutPut != nil {
		sourcePkg = string(g.gen.AddImport(generator.GoImportPath(g.gen.OutPut.SourcePkgPath)))
	}

	g.P("// Reference imports to suppress errors if they are not otherwise used.")
	g.P("var _ ", ctxPkg, ".Context")
	g.P("var _ ", timePkg, ".Timer")
	g.P("var _ =", stringPkg, ".Fields(\"\")")
	g.P("var _ ", DriverPkg, ".Valuer")
	g.P("var _ =", errPkg, ".New(\"\")")
	g.P("var _ ", jsonPkg, ".Any")
	g.P("var _ ", daoPkg, ".Dialect")
	g.P("var _ ", clausePkg, ".Clause")
	g.P()

	for i, msg := range file.Messages() {
		g.wrapSchemas(file, msg, i)
	}

	for key, value := range g.aliasFields {
		g.generateAliasField(file, key, value)
	}

	for _, item := range g.schemas {
		g.generateSchema(file, item)
	}
}

func (g *dao) wrapSchemas(file *generator.FileDescriptor, msg *generator.MessageDescriptor, index int) {
	if msg.Proto.Options != nil && msg.Proto.Options.GetMapEntry() {
		return
	}
	if !g.checkedMessage(msg) {
		return
	}

	s := &Schema{Name: msg.Proto.GetName() + "Schema", Desc: msg}
	for _, item := range msg.Fields {
		field := &Field{Name: generator.CamelCase(item.Proto.GetName()), Desc: item}
		if item.Proto.IsRepeated() {
			alias := generator.CamelCaseSlice([]string{msg.Proto.GetName(), item.Proto.GetName()})
			if strings.HasSuffix(item.Proto.GetTypeName(), "Entry") {
				field.Type = _map
				for _, nest := range msg.Proto.GetNestedType() {
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
			field.Tags = g.buildFieldTags(item, false)
			g.aliasFields[alias] = field
			s.Fields = append(s.Fields, field)
			continue
		}
		typ, tags, err := g.buildFieldTypeAndTags(item)
		if err != nil {
			continue
		}
		field.Type = typ
		field.Tags = tags
		if item.Proto.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			alias := generator.CamelCaseSlice([]string{msg.Proto.GetName(), strings.ReplaceAll(item.Proto.GetTypeName(), ".", "")})
			field.Alias = alias
			g.aliasFields[alias] = field
		}
		s.Fields = append(s.Fields, field)
		fieldTags := g.extractTags(item.Comments)
		if _, ok := fieldTags[_pk]; ok {
			s.PK = field
		}
		if s.PK == nil && (strings.ToLower(field.Name) == "id" || strings.ToLower(field.Name) == "uuid") {
			s.PK = field
		}
	}
	if s.PK == nil {
		g.gen.Fail(fmt.Sprintf(`Message:%s missing primary key`, msg.Proto.GetName()))
	}

	// append deleteTimestamp field
	g.schemas = append(g.schemas, s)
}

func (g *dao) generateAliasField(file *generator.FileDescriptor, alias string, field *Field) {
	if field.IsRepeated {
		// slice, array type
		typ, err := g.buildFieldGoType(file, field.Desc.Proto)
		if err != nil {
			return
		}
		g.P(fmt.Sprintf(`type %s []%s`, alias, typ))

	} else if field.Type == _map {
		if field.Map == nil {
			return
		}

		key, value := field.Map.Key, field.Map.Value
		keyString, _ := g.buildFieldGoType(file, key)
		valueString, _ := g.buildFieldGoType(file, value)
		g.P(fmt.Sprintf(`type %s map[%s]%s`, alias, keyString, valueString))
	} else {
		subMsg := g.extractMessage(field.Desc.Proto.GetTypeName())
		if subMsg.Proto.File() == file {
			g.P(fmt.Sprintf("type %s %s", alias, g.wrapPkg(subMsg.Proto.GetName())))
		} else {
			obj := g.gen.ObjectNamed(field.Desc.Proto.GetTypeName())
			v, ok := g.gen.ImportMap[obj.GoImportPath().String()]
			if !ok {
				v = string(g.gen.AddImport(obj.GoImportPath()))
			}
			g.P(fmt.Sprintf("type %s %s.%s", alias, v, subMsg.Proto.GetName()))
		}
	}
	g.P()

	g.P("// Value return json value, implement driver.Valuer interface")
	if field.Type == _slice || field.Type == _map {
		g.P(fmt.Sprintf(`func (m %s) Value() (driver.Value, error) {`, alias))
		g.P("if len(m) == 0 {")
		g.P("return nil, nil")
		g.P("}")
	} else {
		g.P(fmt.Sprintf(`func (m *%s) Value() (driver.Value, error) {`, alias))
		g.P("if m == nil {")
		g.P("return nil, nil")
		g.P("}")
	}
	g.P(fmt.Sprintf(`b, err := %s.Marshal(m)`, jsonPkg))
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
	g.P(fmt.Sprintf(`return %s.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))`, errPkg))
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`return %s.Unmarshal(bytes, &m)`, jsonPkg))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) DaoDataType() string {`, alias))
	g.P(`return "json"`)
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
	g.P("type ", schema.Name, " struct {")
	for _, field := range schema.Fields {
		switch field.Type {
		case _point:
			g.P(fmt.Sprintf(`%s *%s %s`, field.Name, field.Alias, field.Tags))
		case _slice, _map:
			g.P(fmt.Sprintf(`%s %s %s`, field.Name, field.Alias, field.Tags))
		default:
			g.P(fmt.Sprintf(`%s %s %s`, field.Name, field.Type, field.Tags))
		}
	}
	g.P(fmt.Sprintf(`DeletionTimestamp int64 %s`, toQuoted(`json:"deletionTimestamp" dao:"column:deletion_timestamp"`)))
	g.P("}")
	g.P()
}

func (g *dao) generateSchemaIOMethods(file *generator.FileDescriptor, schema *Schema) {
	g.P(fmt.Sprintf(`func Registry%s(in *%s) error {`, schema.Desc.Proto.GetName(), g.wrapPkg(schema.Desc.Proto.GetName())))
	g.P(fmt.Sprintf(`return dao.DefaultDialect.Migrator().AutoMigrate(From%s(in))`, schema.Desc.Proto.GetName()))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func From%s(in *%s) *%s {`, schema.Desc.Proto.GetName(), g.wrapPkg(schema.Desc.Proto.GetName()), schema.Name))
	g.P(fmt.Sprintf(`out := new(%s)`, schema.Name))
	for _, field := range schema.Fields {
		switch field.Type {
		case _float32, _float64, _int32, _int64, _uint32, _uint64:
			g.P(fmt.Sprintf(`if in.%s != 0 {`, field.Name))
			g.P(fmt.Sprintf(`out.%s = in.%s`, field.Name, field.Name))
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

	g.P(fmt.Sprintf(`func (m *%s) To%s() *%s {`, schema.Name, schema.Desc.Proto.GetName(), g.wrapPkg(schema.Desc.Proto.GetName())))
	g.P(fmt.Sprintf(`out := new(%s)`, g.wrapPkg(schema.Desc.Proto.GetName())))
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
			g.P(fmt.Sprintf(`out.%s = m.%s`, field.Name, field.Name))
		}
	}
	g.P("return out")
	g.P("}")
	g.P()
}

func (g *dao) generateSchemaUtilMethods(file *generator.FileDescriptor, schema *Schema) {
	g.P("func (", schema.Name, ") TableName() string {")
	g.P("return \"", toTableName(schema.Desc.Proto.GetName()), "\"")
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
	g.P(fmt.Sprintf(`func (m *%s) FindPage(ctx context.Context, page, size int, exprs ...clause.Expression) ([]*%s, error) {`, source, g.wrapPkg(target)))
	g.P(fmt.Sprintf(`return m.FindAll(ctx, append(exprs, clause.Limit{Offset: (page - 1) * size, Limit: size})...)`))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) extractClauses(tx *%s.DB) []clause.Expression {`, source, daoPkg))
	g.P(`exprs := make([]clause.Expression, 0)`)
	for _, field := range schema.Fields {
		column := toColumnName(field.Name)
		switch field.Type {
		case _map:
			g.P(fmt.Sprintf(`if m.%s != nil {`, field.Name))
			g.P(fmt.Sprintf(`for k, v := range m.%s {`, field.Name))
			if field.Map.Key.GetType() == descriptor.FieldDescriptorProto_TYPE_STRING {
				g.P(fmt.Sprintf(`exprs = append(exprs, dao.DefaultDialect.JSONBuild(tx, "%s").Equals(v, k))`, column))
			}
			g.P("}")
			g.P("}")
		case _point:
			g.P(fmt.Sprintf(`if m.%s != nil {`, field.Name))
			g.P(fmt.Sprintf(`for k, v := range dao.FieldPatch(m.%s) {`, field.Name))
			g.P(`if v == nil {`)
			g.P(fmt.Sprintf(`exprs = append(exprs, dao.DefaultDialect.JSONBuild(tx, "%s").HasKeys(strings.Split(k, ",")...))`, column))
			g.P(`} else {`)
			g.P(fmt.Sprintf(`exprs = append(exprs, dao.DefaultDialect.JSONBuild(tx, "%s").Equals(v, strings.Split(k, ",")...))`, column))
			g.P("}")
			g.P("}")
			g.P("}")
		case _slice:
			g.P(fmt.Sprintf(`if len(m.%s) != 0 {`, field.Name))
			g.P(fmt.Sprintf(`for _, item := range m.%s {`, field.Name))
			//sjname := field.Slice.GetJsonName()
			//sname := field.Slice.GetName()
			switch field.Slice.GetType() {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				g.P(fmt.Sprintf(`exprs = append(exprs, dao.DefaultDialect.JSONBuild(tx, "%s").Contains(item))`, column))
			case descriptor.FieldDescriptorProto_TYPE_UINT32,
				descriptor.FieldDescriptorProto_TYPE_UINT64,
				descriptor.FieldDescriptorProto_TYPE_INT32,
				descriptor.FieldDescriptorProto_TYPE_INT64,
				descriptor.FieldDescriptorProto_TYPE_SFIXED32,
				descriptor.FieldDescriptorProto_TYPE_SFIXED64,
				descriptor.FieldDescriptorProto_TYPE_FIXED32,
				descriptor.FieldDescriptorProto_TYPE_FIXED64:
				g.P(fmt.Sprintf(`exprs = append(exprs, dao.DefaultDialect.JSONBuild(tx, "%s").Contains(item))`, column))
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				g.P(`for k, v := range dao.FieldPatch(item) {`)
				g.P(`if v != nil {`)
				g.P(fmt.Sprintf(`exprs = append(exprs, dao.DefaultDialect.JSONBuild(tx, "%s").Contains(v, strings.Split(k, ",")...))`, column))
					g.P(`}`)
				g.P("}")
			}
			g.P("}")
			g.P("}")
		case _string:
			g.P(fmt.Sprintf(`if m.%s != "" {`, field.Name))
			g.P(fmt.Sprintf(`exprs = append(exprs, clause.Eq{Column: clause.Column{Name: "%s"}, Value: m.%s})`, column, field.Name))
			g.P("}")
		default:
			g.P(fmt.Sprintf(`if m.%s != 0 {`, field.Name))
			g.P(fmt.Sprintf(`exprs = append(exprs, clause.Eq{Column: clause.Column{Name: "%s"}, Value: m.%s})`, column, field.Name))
			g.P("}")
		}
	}
	g.P()
	g.P(`exprs = append(exprs, clause.Eq{Column: clause.Column{Name: "deletion_timestamp"}, Value: 0})`)
	g.P()
	g.P(`return exprs`)
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) FindAll(ctx context.Context, exprs ...clause.Expression) ([]*%s, error) {`, source, g.wrapPkg(target)))
	g.P(fmt.Sprintf(`dest := make([]*%s, 0)`, source))
	g.P(`tx := dao.DefaultDialect.NewTx().Table(m.TableName()).WithContext(ctx)`)
	g.P()
	g.P(fmt.Sprintf(`tx = tx.Clauses(m.extractClauses(tx)...).Clauses(exprs...)`))
	g.P()
	g.P(`if err := tx.Find(&dest).Error; err != nil {`)
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

	g.P(fmt.Sprintf(`func (m *%s) FindOne(ctx context.Context, exprs ...clause.Expression) (*%s, error) {`, source, g.wrapPkg(target)))
	g.P(`tx := dao.DefaultDialect.NewTx().Table(m.TableName()).WithContext(ctx)`)
	g.P()
	g.P(fmt.Sprintf(`tx = tx.Clauses(m.extractClauses(tx)...).Clauses(exprs...)`))
	g.P()
	g.P(`if err := tx.First(&m).Error; err != nil {`)
	g.P("return nil, err")
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`return m.To%s(), nil`, target))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) Create(ctx context.Context) (*%s, error) {`, source, g.wrapPkg(target)))
	g.P(`tx := dao.DefaultDialect.NewTx().Table(m.TableName()).WithContext(ctx)`)
	g.P()
	g.P(`if err := tx.Create(m).Error; err != nil {`)
	g.P("return nil, err")
	g.P("}")
	g.P()
	g.P(fmt.Sprintf(`return m.To%s(), nil`, target))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) Updates(ctx context.Context) (*%s, error) {`, source, g.wrapPkg(target)))
	g.P(`pk, pkv, isNil := m.PrimaryKey()`)
	g.P(`if isNil {`)
	g.P(`return nil, errors.New("missing primary key")`)
	g.P("}")
	g.P()
	g.P(`tx := dao.DefaultDialect.NewTx().Table(m.TableName()).WithContext(ctx).Where(pk+" = ?", pkv)`)
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
	g.P(`err := dao.DefaultDialect.NewTx().Table(m.TableName()).WithContext(ctx).Where(pk+" = ?", pkv).First(m).Error`)
	g.P("if err != nil {")
	g.P(`return nil, err`)
	g.P("}")
	g.P(fmt.Sprintf(`return m.To%s(), nil`, target))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) SoftDelete(ctx context.Context) error {`, source))
	g.P(`pk, pkv, isNil := m.PrimaryKey()`)
	g.P(`if isNil {`)
	g.P(`return errors.New("missing primary key")`)
	g.P("}")
	g.P()
	g.P(`tx := dao.DefaultDialect.NewTx().Table(m.TableName()).WithContext(ctx)`)
	g.P()
	g.P(fmt.Sprintf(`return tx.Where(pk+" = ?", pkv).Updates(map[string]interface{}{"deletion_timestamp": %s.Now().UnixNano()}).Error`, timePkg))
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`func (m *%s) Delete(ctx context.Context) error {`, source))
	g.P(`pk, pkv, isNil := m.PrimaryKey()`)
	g.P(`if isNil {`)
	g.P(`return errors.New("missing primary key")`)
	g.P("}")
	g.P()
	g.P(`tx := dao.DefaultDialect.NewTx().Table(m.TableName()).WithContext(ctx)`)
	g.P()
	g.P(fmt.Sprintf(`return tx.Where(pk+" = ?", pkv).Delete(&%s{}).Error`, source))
	g.P("}")
	g.P()
}

func (g *dao) checkedMessage(msg *generator.MessageDescriptor) bool {
	tags := g.extractTags(msg.Comments)
	for _, c := range tags {
		if c.Key == _generate {
			return true
		}
	}
	return false
}

func (g *dao) buildFieldTypeAndTags(field *generator.FieldDescriptor) (fieldType, string, error) {
	switch field.Proto.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return _float64, g.buildFieldTags(field, false), nil
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return _float32, g.buildFieldTags(field, false), nil
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_ENUM:
		return _int32, g.buildFieldTags(field, true), nil
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_INT64:
		return _int64, g.buildFieldTags(field, true), nil
	case descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_UINT32:
		return _uint32, g.buildFieldTags(field, true), nil
	case descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_UINT64:
		return _uint64, g.buildFieldTags(field, true), nil
	case descriptor.FieldDescriptorProto_TYPE_STRING,
		descriptor.FieldDescriptorProto_TYPE_BYTES:
		return _string, g.buildFieldTags(field, false), nil
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		return _point, g.buildFieldTags(field, false), nil
	default:
		return "", "", errors.New("invalid field type")
	}
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
	if g.gen.OutPut != nil {
		return sourcePkg + "." + pkg
	}
	return pkg
}

func (g *dao) buildFieldTags(field *generator.FieldDescriptor, autoIncr bool) string {
	var (
		out       = make([]string, 0)
		fieldName = generator.CamelCase(field.Proto.GetName())
		tags      = g.extractTags(field.Comments)
	)
	out = append(out, fmt.Sprintf(`json:"%s,omitempty"`, field.Proto.GetJsonName()))
	daoTag := fmt.Sprintf(`dao:"column:%s"`, toColumnName(fieldName))
	for _, tag := range tags {
		// TODO: parse dao tags
		switch tag.Key {
		case _pk:
			daoTag += ";primaryKey"
			if autoIncr {
				daoTag += ";autoIncrement"
			}
		}
	}
	out = append(out, daoTag)

	return toQuoted(strings.Join(out, " "))
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
