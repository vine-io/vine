// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package field

import (
	"reflect"

	"github.com/lack-io/vine/service/ent/schema"
)

// Int returns a new Field with type int.
func Int(name string) *intBuilder {
	return &intBuilder{&Descriptor{
		Name: name,
		Info: &TypeInfo{Type: TypeInt},
	}}
}

// Uint returns a new Field with type uint.
func Uint(name string) *uintBuilder {
	return &uintBuilder{&Descriptor{
		Name: name,
		Info: &TypeInfo{Type: TypeUint},
	}}
}

// Int8 returns a new Field with type int8.
func Int8(name string) *int8Builder {
	return &int8Builder{&Descriptor{
		Name: name,
		Info: &TypeInfo{Type: TypeInt8},
	}}
}

// Int16 returns a new Field with type int16.
func Int16(name string) *int16Builder {
	return &int16Builder{&Descriptor{
		Name: name,
		Info: &TypeInfo{Type: TypeInt16},
	}}
}

// Int32 returns a new Field with type int32.
func Int32(name string) *int32Builder {
	return &int32Builder{&Descriptor{
		Name: name,
		Info: &TypeInfo{Type: TypeInt32},
	}}
}

// Int64 returns a new Field with type int64.
func Int64(name string) *int64Builder {
	return &int64Builder{&Descriptor{
		Name: name,
		Info: &TypeInfo{Type: TypeInt64},
	}}
}

// Uint8 returns a new Field with type uint8.
func Uint8(name string) *uint8Builder {
	return &uint8Builder{&Descriptor{
		Name: name,
		Info: &TypeInfo{Type: TypeUint8},
	}}
}

// Uint16 returns a new Field with type uint16.
func Uint16(name string) *uint16Builder {
	return &uint16Builder{&Descriptor{
		Name: name,
		Info: &TypeInfo{Type: TypeUint16},
	}}
}

// Uint32 returns a new Field with type uint32.
func Uint32(name string) *uint32Builder {
	return &uint32Builder{&Descriptor{
		Name: name,
		Info: &TypeInfo{Type: TypeUint32},
	}}
}

// Uint64 returns a new Field with type uint64.
func Uint64(name string) *uint64Builder {
	return &uint64Builder{&Descriptor{
		Name: name,
		Info: &TypeInfo{Type: TypeUint64},
	}}
}

// Float returns a new Field with type float64.
func Float(name string) *float64Builder {
	return &float64Builder{&Descriptor{
		Name: name,
		Info: &TypeInfo{Type: TypeFloat64},
	}}
}

// Float32 returns a new Field with type float32.
func Float32(name string) *float32Builder {
	return &float32Builder{&Descriptor{
		Name: name,
		Info: &TypeInfo{Type: TypeFloat32},
	}}
}

// intBuilder is the builder for int field.
type intBuilder struct {
	desc *Descriptor
}

// Unique makes the field unique within all vertices of this type.
func (b *intBuilder) Unique() *intBuilder {
	b.desc.Unique = true
	return b
}

// Default sets the default value of the field.
func (b *intBuilder) Default(i int) *intBuilder {
	b.desc.Default = i
	return b
}

// DefaultFunc sets the function that is applied to set the default value
// of the field on creation.
func (b *intBuilder) DefaultFunc(fn interface{}) *intBuilder {
	b.desc.Default = fn
	return b
}

// Nillable indicates that this field is a nillable.
// Unlike "Optional" only fields, "Nillable" fields are pointers in the generated struct.
func (b *intBuilder) Nillable() *intBuilder {
	b.desc.Nillable = true
	return b
}

// Comment sets the comment of the field.
func (b *intBuilder) Comment(c string) *intBuilder {
	b.desc.Comment = c
	return b
}

// Optional indicates that this field is optional on create.
// Unlike edges, fields are required by default.
func (b *intBuilder) Optional() *intBuilder {
	b.desc.Optional = true
	return b
}

// Immutable indicates that this field cannot be updated.
func (b *intBuilder) Immutable() *intBuilder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the field.
func (b *intBuilder) StructTag(s string) *intBuilder {
	b.desc.Tag = s
	return b
}

// StorageKey sets the storage key of the field.
// In SQL dialects is the column name and Gremlin is the property.
func (b *intBuilder) StorageKey(key string) *intBuilder {
	b.desc.StorageKey = key
	return b
}

// SchemaType overrides the default database type with a custom
// schema type (per dialect) for int.
//
//	field.Int("oid").
//		SchemaType(map[string]string{
//			dialect.Postgres: "CustomType",
//		})
//
func (b *intBuilder) SchemaType(types map[string]string) *intBuilder {
	b.desc.SchemaType = types
	return b
}

// GoType overrides the default Go type with a custom one.
//
//	field.Int("int").
//		GoType(pkg.Int(0))
//
func (b *intBuilder) GoType(typ interface{}) *intBuilder {
	b.desc.goType(typ, intType)
	return b
}

// Annotations adds a list of annotations to the field object to be used by
// codegen extensions.
//
//	field.Int("int").
//		Annotations(entgql.Config{
//			Ordered: true,
//		})
//
func (b *intBuilder) Annotations(annotations ...schema.Annotation) *intBuilder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Field interface by returning its descriptor.
func (b *intBuilder) Descriptor() *Descriptor {
	if b.desc.Default != nil {
		b.desc.checkDefaultFunc(intType)
	}
	return b.desc
}

// uintBuilder is the builder for uint field.
type uintBuilder struct {
	desc *Descriptor
}

// Unique makes the field unique within all vertices of this type.
func (b *uintBuilder) Unique() *uintBuilder {
	b.desc.Unique = true
	return b
}

// Default sets the default value of the field.
func (b *uintBuilder) Default(i uint) *uintBuilder {
	b.desc.Default = i
	return b
}

// DefaultFunc sets the function that is applied to set the default value
// of the field on creation.
func (b *uintBuilder) DefaultFunc(fn interface{}) *uintBuilder {
	b.desc.Default = fn
	return b
}

// Nillable indicates that this field is a nillable.
// Unlike "Optional" only fields, "Nillable" fields are pointers in the generated struct.
func (b *uintBuilder) Nillable() *uintBuilder {
	b.desc.Nillable = true
	return b
}

// Comment sets the comment of the field.
func (b *uintBuilder) Comment(c string) *uintBuilder {
	b.desc.Comment = c
	return b
}

// Optional indicates that this field is optional on create.
// Unlike edges, fields are required by default.
func (b *uintBuilder) Optional() *uintBuilder {
	b.desc.Optional = true
	return b
}

// Immutable indicates that this field cannot be updated.
func (b *uintBuilder) Immutable() *uintBuilder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the field.
func (b *uintBuilder) StructTag(s string) *uintBuilder {
	b.desc.Tag = s
	return b
}

// StorageKey sets the storage key of the field.
// In SQL dialects is the column name and Gremlin is the property.
func (b *uintBuilder) StorageKey(key string) *uintBuilder {
	b.desc.StorageKey = key
	return b
}

// SchemaType overrides the default database type with a custom
// schema type (per dialect) for uint.
//
//	field.Uint("oid").
//		SchemaType(map[string]string{
//			dialect.Postgres: "CustomType",
//		})
//
func (b *uintBuilder) SchemaType(types map[string]string) *uintBuilder {
	b.desc.SchemaType = types
	return b
}

// GoType overrides the default Go type with a custom one.
//
//	field.Uint("uint").
//		GoType(pkg.Uint(0))
//
func (b *uintBuilder) GoType(typ interface{}) *uintBuilder {
	b.desc.goType(typ, uintType)
	return b
}

// Annotations adds a list of annotations to the field object to be used by
// codegen extensions.
//
//	field.Uint("uint").
//		Annotations(entgql.Config{
//			Ordered: true,
//		})
//
func (b *uintBuilder) Annotations(annotations ...schema.Annotation) *uintBuilder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Field interface by returning its descriptor.
func (b *uintBuilder) Descriptor() *Descriptor {
	if b.desc.Default != nil {
		b.desc.checkDefaultFunc(uintType)
	}
	return b.desc
}

// int8Builder is the builder for int8 field.
type int8Builder struct {
	desc *Descriptor
}

// Unique makes the field unique within all vertices of this type.
func (b *int8Builder) Unique() *int8Builder {
	b.desc.Unique = true
	return b
}

// Default sets the default value of the field.
func (b *int8Builder) Default(i int8) *int8Builder {
	b.desc.Default = i
	return b
}

// DefaultFunc sets the function that is applied to set the default value
// of the field on creation.
func (b *int8Builder) DefaultFunc(fn interface{}) *int8Builder {
	b.desc.Default = fn
	return b
}

// Nillable indicates that this field is a nillable.
// Unlike "Optional" only fields, "Nillable" fields are pointers in the generated struct.
func (b *int8Builder) Nillable() *int8Builder {
	b.desc.Nillable = true
	return b
}

// Comment sets the comment of the field.
func (b *int8Builder) Comment(c string) *int8Builder {
	b.desc.Comment = c
	return b
}

// Optional indicates that this field is optional on create.
// Unlike edges, fields are required by default.
func (b *int8Builder) Optional() *int8Builder {
	b.desc.Optional = true
	return b
}

// Immutable indicates that this field cannot be updated.
func (b *int8Builder) Immutable() *int8Builder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the field.
func (b *int8Builder) StructTag(s string) *int8Builder {
	b.desc.Tag = s
	return b
}

// StorageKey sets the storage key of the field.
// In SQL dialects is the column name and Gremlin is the property.
func (b *int8Builder) StorageKey(key string) *int8Builder {
	b.desc.StorageKey = key
	return b
}

// SchemaType overrides the default database type with a custom
// schema type (per dialect) for int8.
//
//	field.Int8("oid").
//		SchemaType(map[string]string{
//			dialect.Postgres: "CustomType",
//		})
//
func (b *int8Builder) SchemaType(types map[string]string) *int8Builder {
	b.desc.SchemaType = types
	return b
}

// GoType overrides the default Go type with a custom one.
//
//	field.Int8("int8").
//		GoType(pkg.Int8(0))
//
func (b *int8Builder) GoType(typ interface{}) *int8Builder {
	b.desc.goType(typ, int8Type)
	return b
}

// Annotations adds a list of annotations to the field object to be used by
// codegen extensions.
//
//	field.Int8("int8").
//		Annotations(entgql.Config{
//			Ordered: true,
//		})
//
func (b *int8Builder) Annotations(annotations ...schema.Annotation) *int8Builder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Field interface by returning its descriptor.
func (b *int8Builder) Descriptor() *Descriptor {
	if b.desc.Default != nil {
		b.desc.checkDefaultFunc(int8Type)
	}
	return b.desc
}

// int16Builder is the builder for int16 field.
type int16Builder struct {
	desc *Descriptor
}

// Unique makes the field unique within all vertices of this type.
func (b *int16Builder) Unique() *int16Builder {
	b.desc.Unique = true
	return b
}

// Default sets the default value of the field.
func (b *int16Builder) Default(i int16) *int16Builder {
	b.desc.Default = i
	return b
}

// DefaultFunc sets the function that is applied to set the default value
// of the field on creation.
func (b *int16Builder) DefaultFunc(fn interface{}) *int16Builder {
	b.desc.Default = fn
	return b
}

// Nillable indicates that this field is a nillable.
// Unlike "Optional" only fields, "Nillable" fields are pointers in the generated struct.
func (b *int16Builder) Nillable() *int16Builder {
	b.desc.Nillable = true
	return b
}

// Comment sets the comment of the field.
func (b *int16Builder) Comment(c string) *int16Builder {
	b.desc.Comment = c
	return b
}

// Optional indicates that this field is optional on create.
// Unlike edges, fields are required by default.
func (b *int16Builder) Optional() *int16Builder {
	b.desc.Optional = true
	return b
}

// Immutable indicates that this field cannot be updated.
func (b *int16Builder) Immutable() *int16Builder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the field.
func (b *int16Builder) StructTag(s string) *int16Builder {
	b.desc.Tag = s
	return b
}

// StorageKey sets the storage key of the field.
// In SQL dialects is the column name and Gremlin is the property.
func (b *int16Builder) StorageKey(key string) *int16Builder {
	b.desc.StorageKey = key
	return b
}

// SchemaType overrides the default database type with a custom
// schema type (per dialect) for int16.
//
//	field.Int16("oid").
//		SchemaType(map[string]string{
//			dialect.Postgres: "CustomType",
//		})
//
func (b *int16Builder) SchemaType(types map[string]string) *int16Builder {
	b.desc.SchemaType = types
	return b
}

// GoType overrides the default Go type with a custom one.
//
//	field.Int16("int16").
//		GoType(pkg.Int16(0))
//
func (b *int16Builder) GoType(typ interface{}) *int16Builder {
	b.desc.goType(typ, int16Type)
	return b
}

// Annotations adds a list of annotations to the field object to be used by
// codegen extensions.
//
//	field.Int16("int16").
//		Annotations(entgql.Config{
//			Ordered: true,
//		})
//
func (b *int16Builder) Annotations(annotations ...schema.Annotation) *int16Builder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Field interface by returning its descriptor.
func (b *int16Builder) Descriptor() *Descriptor {
	if b.desc.Default != nil {
		b.desc.checkDefaultFunc(int16Type)
	}
	return b.desc
}

// int32Builder is the builder for int32 field.
type int32Builder struct {
	desc *Descriptor
}

// Unique makes the field unique within all vertices of this type.
func (b *int32Builder) Unique() *int32Builder {
	b.desc.Unique = true
	return b
}

// Default sets the default value of the field.
func (b *int32Builder) Default(i int32) *int32Builder {
	b.desc.Default = i
	return b
}

// DefaultFunc sets the function that is applied to set the default value
// of the field on creation.
func (b *int32Builder) DefaultFunc(fn interface{}) *int32Builder {
	b.desc.Default = fn
	return b
}

// Nillable indicates that this field is a nillable.
// Unlike "Optional" only fields, "Nillable" fields are pointers in the generated struct.
func (b *int32Builder) Nillable() *int32Builder {
	b.desc.Nillable = true
	return b
}

// Comment sets the comment of the field.
func (b *int32Builder) Comment(c string) *int32Builder {
	b.desc.Comment = c
	return b
}

// Optional indicates that this field is optional on create.
// Unlike edges, fields are required by default.
func (b *int32Builder) Optional() *int32Builder {
	b.desc.Optional = true
	return b
}

// Immutable indicates that this field cannot be updated.
func (b *int32Builder) Immutable() *int32Builder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the field.
func (b *int32Builder) StructTag(s string) *int32Builder {
	b.desc.Tag = s
	return b
}

// StorageKey sets the storage key of the field.
// In SQL dialects is the column name and Gremlin is the property.
func (b *int32Builder) StorageKey(key string) *int32Builder {
	b.desc.StorageKey = key
	return b
}

// SchemaType overrides the default database type with a custom
// schema type (per dialect) for int32.
//
//	field.Int32("oid").
//		SchemaType(map[string]string{
//			dialect.Postgres: "CustomType",
//		})
//
func (b *int32Builder) SchemaType(types map[string]string) *int32Builder {
	b.desc.SchemaType = types
	return b
}

// GoType overrides the default Go type with a custom one.
//
//	field.Int32("int32").
//		GoType(pkg.Int32(0))
//
func (b *int32Builder) GoType(typ interface{}) *int32Builder {
	b.desc.goType(typ, int32Type)
	return b
}

// Annotations adds a list of annotations to the field object to be used by
// codegen extensions.
//
//	field.Int32("int32").
//		Annotations(entgql.Config{
//			Ordered: true,
//		})
//
func (b *int32Builder) Annotations(annotations ...schema.Annotation) *int32Builder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Field interface by returning its descriptor.
func (b *int32Builder) Descriptor() *Descriptor {
	if b.desc.Default != nil {
		b.desc.checkDefaultFunc(int32Type)
	}
	return b.desc
}

// int64Builder is the builder for int64 field.
type int64Builder struct {
	desc *Descriptor
}

// Unique makes the field unique within all vertices of this type.
func (b *int64Builder) Unique() *int64Builder {
	b.desc.Unique = true
	return b
}

// Default sets the default value of the field.
func (b *int64Builder) Default(i int64) *int64Builder {
	b.desc.Default = i
	return b
}

// DefaultFunc sets the function that is applied to set the default value
// of the field on creation.
func (b *int64Builder) DefaultFunc(fn interface{}) *int64Builder {
	b.desc.Default = fn
	return b
}

// Nillable indicates that this field is a nillable.
// Unlike "Optional" only fields, "Nillable" fields are pointers in the generated struct.
func (b *int64Builder) Nillable() *int64Builder {
	b.desc.Nillable = true
	return b
}

// Comment sets the comment of the field.
func (b *int64Builder) Comment(c string) *int64Builder {
	b.desc.Comment = c
	return b
}

// Optional indicates that this field is optional on create.
// Unlike edges, fields are required by default.
func (b *int64Builder) Optional() *int64Builder {
	b.desc.Optional = true
	return b
}

// Immutable indicates that this field cannot be updated.
func (b *int64Builder) Immutable() *int64Builder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the field.
func (b *int64Builder) StructTag(s string) *int64Builder {
	b.desc.Tag = s
	return b
}

// StorageKey sets the storage key of the field.
// In SQL dialects is the column name and Gremlin is the property.
func (b *int64Builder) StorageKey(key string) *int64Builder {
	b.desc.StorageKey = key
	return b
}

// SchemaType overrides the default database type with a custom
// schema type (per dialect) for int64.
//
//	field.Int64("oid").
//		SchemaType(map[string]string{
//			dialect.Postgres: "CustomType",
//		})
//
func (b *int64Builder) SchemaType(types map[string]string) *int64Builder {
	b.desc.SchemaType = types
	return b
}

// GoType overrides the default Go type with a custom one.
//
//	field.Int64("int64").
//		GoType(pkg.Int64(0))
//
func (b *int64Builder) GoType(typ interface{}) *int64Builder {
	b.desc.goType(typ, int64Type)
	return b
}

// Annotations adds a list of annotations to the field object to be used by
// codegen extensions.
//
//	field.Int64("int64").
//		Annotations(entgql.Config{
//			Ordered: true,
//		})
//
func (b *int64Builder) Annotations(annotations ...schema.Annotation) *int64Builder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Field interface by returning its descriptor.
func (b *int64Builder) Descriptor() *Descriptor {
	if b.desc.Default != nil {
		b.desc.checkDefaultFunc(int64Type)
	}
	return b.desc
}

// uint8Builder is the builder for uint8 field.
type uint8Builder struct {
	desc *Descriptor
}

// Unique makes the field unique within all vertices of this type.
func (b *uint8Builder) Unique() *uint8Builder {
	b.desc.Unique = true
	return b
}

// Default sets the default value of the field.
func (b *uint8Builder) Default(i uint8) *uint8Builder {
	b.desc.Default = i
	return b
}

// DefaultFunc sets the function that is applied to set the default value
// of the field on creation.
func (b *uint8Builder) DefaultFunc(fn interface{}) *uint8Builder {
	b.desc.Default = fn
	return b
}

// Nillable indicates that this field is a nillable.
// Unlike "Optional" only fields, "Nillable" fields are pointers in the generated struct.
func (b *uint8Builder) Nillable() *uint8Builder {
	b.desc.Nillable = true
	return b
}

// Comment sets the comment of the field.
func (b *uint8Builder) Comment(c string) *uint8Builder {
	b.desc.Comment = c
	return b
}

// Optional indicates that this field is optional on create.
// Unlike edges, fields are required by default.
func (b *uint8Builder) Optional() *uint8Builder {
	b.desc.Optional = true
	return b
}

// Immutable indicates that this field cannot be updated.
func (b *uint8Builder) Immutable() *uint8Builder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the field.
func (b *uint8Builder) StructTag(s string) *uint8Builder {
	b.desc.Tag = s
	return b
}

// StorageKey sets the storage key of the field.
// In SQL dialects is the column name and Gremlin is the property.
func (b *uint8Builder) StorageKey(key string) *uint8Builder {
	b.desc.StorageKey = key
	return b
}

// SchemaType overrides the default database type with a custom
// schema type (per dialect) for uint8.
//
//	field.Uint8("oid").
//		SchemaType(map[string]string{
//			dialect.Postgres: "CustomType",
//		})
//
func (b *uint8Builder) SchemaType(types map[string]string) *uint8Builder {
	b.desc.SchemaType = types
	return b
}

// GoType overrides the default Go type with a custom one.
//
//	field.Uint8("uint8").
//		GoType(pkg.Uint8(0))
//
func (b *uint8Builder) GoType(typ interface{}) *uint8Builder {
	b.desc.goType(typ, uint8Type)
	return b
}

// Annotations adds a list of annotations to the field object to be used by
// codegen extensions.
//
//	field.Uint8("uint8").
//		Annotations(entgql.Config{
//			Ordered: true,
//		})
//
func (b *uint8Builder) Annotations(annotations ...schema.Annotation) *uint8Builder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Field interface by returning its descriptor.
func (b *uint8Builder) Descriptor() *Descriptor {
	if b.desc.Default != nil {
		b.desc.checkDefaultFunc(uint8Type)
	}
	return b.desc
}

// uint16Builder is the builder for uint16 field.
type uint16Builder struct {
	desc *Descriptor
}

// Unique makes the field unique within all vertices of this type.
func (b *uint16Builder) Unique() *uint16Builder {
	b.desc.Unique = true
	return b
}

// Default sets the default value of the field.
func (b *uint16Builder) Default(i uint16) *uint16Builder {
	b.desc.Default = i
	return b
}

// DefaultFunc sets the function that is applied to set the default value
// of the field on creation.
func (b *uint16Builder) DefaultFunc(fn interface{}) *uint16Builder {
	b.desc.Default = fn
	return b
}

// Nillable indicates that this field is a nillable.
// Unlike "Optional" only fields, "Nillable" fields are pointers in the generated struct.
func (b *uint16Builder) Nillable() *uint16Builder {
	b.desc.Nillable = true
	return b
}

// Comment sets the comment of the field.
func (b *uint16Builder) Comment(c string) *uint16Builder {
	b.desc.Comment = c
	return b
}

// Optional indicates that this field is optional on create.
// Unlike edges, fields are required by default.
func (b *uint16Builder) Optional() *uint16Builder {
	b.desc.Optional = true
	return b
}

// Immutable indicates that this field cannot be updated.
func (b *uint16Builder) Immutable() *uint16Builder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the field.
func (b *uint16Builder) StructTag(s string) *uint16Builder {
	b.desc.Tag = s
	return b
}

// StorageKey sets the storage key of the field.
// In SQL dialects is the column name and Gremlin is the property.
func (b *uint16Builder) StorageKey(key string) *uint16Builder {
	b.desc.StorageKey = key
	return b
}

// SchemaType overrides the default database type with a custom
// schema type (per dialect) for uint16.
//
//	field.Uint16("oid").
//		SchemaType(map[string]string{
//			dialect.Postgres: "CustomType",
//		})
//
func (b *uint16Builder) SchemaType(types map[string]string) *uint16Builder {
	b.desc.SchemaType = types
	return b
}

// GoType overrides the default Go type with a custom one.
//
//	field.Uint16("uint16").
//		GoType(pkg.Uint16(0))
//
func (b *uint16Builder) GoType(typ interface{}) *uint16Builder {
	b.desc.goType(typ, uint16Type)
	return b
}

// Annotations adds a list of annotations to the field object to be used by
// codegen extensions.
//
//	field.Uint16("uint16").
//		Annotations(entgql.Config{
//			Ordered: true,
//		})
//
func (b *uint16Builder) Annotations(annotations ...schema.Annotation) *uint16Builder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Field interface by returning its descriptor.
func (b *uint16Builder) Descriptor() *Descriptor {
	if b.desc.Default != nil {
		b.desc.checkDefaultFunc(uint16Type)
	}
	return b.desc
}

// uint32Builder is the builder for uint32 field.
type uint32Builder struct {
	desc *Descriptor
}

// Unique makes the field unique within all vertices of this type.
func (b *uint32Builder) Unique() *uint32Builder {
	b.desc.Unique = true
	return b
}

// Default sets the default value of the field.
func (b *uint32Builder) Default(i uint32) *uint32Builder {
	b.desc.Default = i
	return b
}

// DefaultFunc sets the function that is applied to set the default value
// of the field on creation.
func (b *uint32Builder) DefaultFunc(fn interface{}) *uint32Builder {
	b.desc.Default = fn
	return b
}

// Nillable indicates that this field is a nillable.
// Unlike "Optional" only fields, "Nillable" fields are pointers in the generated struct.
func (b *uint32Builder) Nillable() *uint32Builder {
	b.desc.Nillable = true
	return b
}

// Comment sets the comment of the field.
func (b *uint32Builder) Comment(c string) *uint32Builder {
	b.desc.Comment = c
	return b
}

// Optional indicates that this field is optional on create.
// Unlike edges, fields are required by default.
func (b *uint32Builder) Optional() *uint32Builder {
	b.desc.Optional = true
	return b
}

// Immutable indicates that this field cannot be updated.
func (b *uint32Builder) Immutable() *uint32Builder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the field.
func (b *uint32Builder) StructTag(s string) *uint32Builder {
	b.desc.Tag = s
	return b
}

// StorageKey sets the storage key of the field.
// In SQL dialects is the column name and Gremlin is the property.
func (b *uint32Builder) StorageKey(key string) *uint32Builder {
	b.desc.StorageKey = key
	return b
}

// SchemaType overrides the default database type with a custom
// schema type (per dialect) for uint32.
//
//	field.Uint32("oid").
//		SchemaType(map[string]string{
//			dialect.Postgres: "CustomType",
//		})
//
func (b *uint32Builder) SchemaType(types map[string]string) *uint32Builder {
	b.desc.SchemaType = types
	return b
}

// GoType overrides the default Go type with a custom one.
//
//	field.Uint32("uint32").
//		GoType(pkg.Uint32(0))
//
func (b *uint32Builder) GoType(typ interface{}) *uint32Builder {
	b.desc.goType(typ, uint32Type)
	return b
}

// Annotations adds a list of annotations to the field object to be used by
// codegen extensions.
//
//	field.Uint32("uint32").
//		Annotations(entgql.Config{
//			Ordered: true,
//		})
//
func (b *uint32Builder) Annotations(annotations ...schema.Annotation) *uint32Builder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Field interface by returning its descriptor.
func (b *uint32Builder) Descriptor() *Descriptor {
	if b.desc.Default != nil {
		b.desc.checkDefaultFunc(uint32Type)
	}
	return b.desc
}

// uint64Builder is the builder for uint64 field.
type uint64Builder struct {
	desc *Descriptor
}

// Unique makes the field unique within all vertices of this type.
func (b *uint64Builder) Unique() *uint64Builder {
	b.desc.Unique = true
	return b
}

// Default sets the default value of the field.
func (b *uint64Builder) Default(i uint64) *uint64Builder {
	b.desc.Default = i
	return b
}

// DefaultFunc sets the function that is applied to set the default value
// of the field on creation.
func (b *uint64Builder) DefaultFunc(fn interface{}) *uint64Builder {
	b.desc.Default = fn
	return b
}

// Nillable indicates that this field is a nillable.
// Unlike "Optional" only fields, "Nillable" fields are pointers in the generated struct.
func (b *uint64Builder) Nillable() *uint64Builder {
	b.desc.Nillable = true
	return b
}

// Comment sets the comment of the field.
func (b *uint64Builder) Comment(c string) *uint64Builder {
	b.desc.Comment = c
	return b
}

// Optional indicates that this field is optional on create.
// Unlike edges, fields are required by default.
func (b *uint64Builder) Optional() *uint64Builder {
	b.desc.Optional = true
	return b
}

// Immutable indicates that this field cannot be updated.
func (b *uint64Builder) Immutable() *uint64Builder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the field.
func (b *uint64Builder) StructTag(s string) *uint64Builder {
	b.desc.Tag = s
	return b
}

// StorageKey sets the storage key of the field.
// In SQL dialects is the column name and Gremlin is the property.
func (b *uint64Builder) StorageKey(key string) *uint64Builder {
	b.desc.StorageKey = key
	return b
}

// SchemaType overrides the default database type with a custom
// schema type (per dialect) for uint64.
//
//	field.Uint64("oid").
//		SchemaType(map[string]string{
//			dialect.Postgres: "CustomType",
//		})
//
func (b *uint64Builder) SchemaType(types map[string]string) *uint64Builder {
	b.desc.SchemaType = types
	return b
}

// GoType overrides the default Go type with a custom one.
//
//	field.Uint64("uint64").
//		GoType(pkg.Uint64(0))
//
func (b *uint64Builder) GoType(typ interface{}) *uint64Builder {
	b.desc.goType(typ, uint64Type)
	return b
}

// Annotations adds a list of annotations to the field object to be used by
// codegen extensions.
//
//	field.Uint64("uint64").
//		Annotations(entgql.Config{
//			Ordered: true,
//		})
//
func (b *uint64Builder) Annotations(annotations ...schema.Annotation) *uint64Builder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Field interface by returning its descriptor.
func (b *uint64Builder) Descriptor() *Descriptor {
	if b.desc.Default != nil {
		b.desc.checkDefaultFunc(uint64Type)
	}
	return b.desc
}

var (
	intType    = reflect.TypeOf(int(0))
	uintType   = reflect.TypeOf(uint(0))
	int8Type   = reflect.TypeOf(int8(0))
	int16Type  = reflect.TypeOf(int16(0))
	int32Type  = reflect.TypeOf(int32(0))
	int64Type  = reflect.TypeOf(int64(0))
	uint8Type  = reflect.TypeOf(uint8(0))
	uint16Type = reflect.TypeOf(uint16(0))
	uint32Type = reflect.TypeOf(uint32(0))
	uint64Type = reflect.TypeOf(uint64(0))
)

// float64Builder is the builder for float fields.
type float64Builder struct {
	desc *Descriptor
}

// Unique makes the field unique within all vertices of this type.
func (b *float64Builder) Unique() *float64Builder {
	b.desc.Unique = true
	return b
}

// Default sets the default value of the field.
func (b *float64Builder) Default(i float64) *float64Builder {
	b.desc.Default = i
	return b
}

// Nillable indicates that this field is a nillable.
// Unlike "Optional" only fields, "Nillable" fields are pointers in the generated struct.
func (b *float64Builder) Nillable() *float64Builder {
	b.desc.Nillable = true
	return b
}

// Comment sets the comment of the field.
func (b *float64Builder) Comment(c string) *float64Builder {
	b.desc.Comment = c
	return b
}

// Optional indicates that this field is optional on create.
// Unlike edges, fields are required by default.
func (b *float64Builder) Optional() *float64Builder {
	b.desc.Optional = true
	return b
}

// Immutable indicates that this field cannot be updated.
func (b *float64Builder) Immutable() *float64Builder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the field.
func (b *float64Builder) StructTag(s string) *float64Builder {
	b.desc.Tag = s
	return b
}

// StorageKey sets the storage key of the field.
// In SQL dialects is the column name and Gremlin is the property.
func (b *float64Builder) StorageKey(key string) *float64Builder {
	b.desc.StorageKey = key
	return b
}

// SchemaType overrides the default database type with a custom
// schema type (per dialect) for float64.
//
//	field.Float64("amount").
//		SchemaType(map[string]string{
//			dialect.MySQL:		"decimal(5, 2)",
//			dialect.Postgres: 	"numeric(5, 2)",
//		})
//
func (b *float64Builder) SchemaType(types map[string]string) *float64Builder {
	b.desc.SchemaType = types
	return b
}

// GoType overrides the default Go type with a custom one.
//
//	field.Float64("float64").
//		GoType(pkg.Float64(0))
//
func (b *float64Builder) GoType(typ interface{}) *float64Builder {
	b.desc.goType(typ, float64Type)
	return b
}

// Annotations adds a list of annotations to the field object to be used by
// codegen extensions.
//
//	field.Float64("float64").
//		Annotations(entgql.Config{
//			Ordered: true,
//		})
//
func (b *float64Builder) Annotations(annotations ...schema.Annotation) *float64Builder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Field interface by returning its descriptor.
func (b *float64Builder) Descriptor() *Descriptor {
	return b.desc
}

// float32Builder is the builder for float fields.
type float32Builder struct {
	desc *Descriptor
}

// Unique makes the field unique within all vertices of this type.
func (b *float32Builder) Unique() *float32Builder {
	b.desc.Unique = true
	return b
}

// Default sets the default value of the field.
func (b *float32Builder) Default(i float32) *float32Builder {
	b.desc.Default = i
	return b
}

// Nillable indicates that this field is a nillable.
// Unlike "Optional" only fields, "Nillable" fields are pointers in the generated struct.
func (b *float32Builder) Nillable() *float32Builder {
	b.desc.Nillable = true
	return b
}

// Comment sets the comment of the field.
func (b *float32Builder) Comment(c string) *float32Builder {
	b.desc.Comment = c
	return b
}

// Optional indicates that this field is optional on create.
// Unlike edges, fields are required by default.
func (b *float32Builder) Optional() *float32Builder {
	b.desc.Optional = true
	return b
}

// Immutable indicates that this field cannot be updated.
func (b *float32Builder) Immutable() *float32Builder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the field.
func (b *float32Builder) StructTag(s string) *float32Builder {
	b.desc.Tag = s
	return b
}

// StorageKey sets the storage key of the field.
// In SQL dialects is the column name and Gremlin is the property.
func (b *float32Builder) StorageKey(key string) *float32Builder {
	b.desc.StorageKey = key
	return b
}

// SchemaType overrides the default database type with a custom
// schema type (per dialect) for float32.
//
//	field.Float32("amount").
//		SchemaType(map[string]string{
//			dialect.MySQL:		"decimal(5, 2)",
//			dialect.Postgres: 	"numeric(5, 2)",
//		})
//
func (b *float32Builder) SchemaType(types map[string]string) *float32Builder {
	b.desc.SchemaType = types
	return b
}

// GoType overrides the default Go type with a custom one.
//
//	field.Float32("float32").
//		GoType(pkg.Float32(0))
//
func (b *float32Builder) GoType(typ interface{}) *float32Builder {
	b.desc.goType(typ, float32Type)
	return b
}

// Annotations adds a list of annotations to the field object to be used by
// codegen extensions.
//
//	field.Float32("float32").
//		Annotations(entgql.Config{
//			Ordered: true,
//		})
//
func (b *float32Builder) Annotations(annotations ...schema.Annotation) *float32Builder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Field interface by returning its descriptor.
func (b *float32Builder) Descriptor() *Descriptor {
	return b.desc
}

var (
	float64Type = reflect.TypeOf(float64(0))
	float32Type = reflect.TypeOf(float32(0))
)
