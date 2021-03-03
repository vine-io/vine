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

package schema

import (
	"reflect"
	"time"
)

type DataType string

type TimeType int64

var TimeReflectType = reflect.TypeOf(time.Time{})

const (
	UnixSecond TimeType = iota + 1
	UnixMillisecond
	UnixNanosecond
)

const (
	Bool   DataType = "bool"
	Int    DataType = "int"
	Uint   DataType = "uint"
	Float  DataType = "float"
	String DataType = "string"
	Time   DataType = "time"
	Bytes  DataType = "bytes"
)

type Field struct {
	Name                   string
	DBName                 string
	BindNames              []string
	DataType               DataType
	DAODataType            DataType
	PrimaryKey             bool
	AutoIncrement          bool
	AutoIncrementIncrement bool
	Creatable              bool
	Updatable              bool
	Readable               bool
	HasDefaultValue        bool
	AutoCreateTime         TimeType
	AUtoUpdateTime         TimeType
	DefaultValue           string
	DefaultValueInterface  interface{}
	NotNull                bool
	Unique                 bool
	Comment                string
	Size                   int
	Precision              int
	Scale                  int
	FieldType              reflect.Type
	IndirectFieldType      reflect.Type
	StructField            reflect.StructField
	Tag                    reflect.StructTag
	TagSettings            map[string]string
	Schema                 *Schema
	EmbeddedSchema         *Schema
	OwnerSchema            *Schema
	ReflectValueOf         func(reflect.Value) reflect.Value
	ValueOf                func(reflect.Value) reflect.Value
	Set                    func(reflect.Value, interface{}) error
	IgnoreMigrate          bool
}
