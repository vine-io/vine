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

package plugin

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

	"github.com/lack-io/vine/cmd/generator"
)

type fieldType string

const (
	_float32 fieldType = "float32"
	_float64 fieldType = "float64"
	_int32   fieldType = "int32"
	_int64   fieldType = "int64"
	_uint32  fieldType = "uint32"
	_uint64  fieldType = "uint64"
	_string  fieldType = "string"
	_map     fieldType = "map"
	_slice   fieldType = "slice"
	_point   fieldType = "point"
)

type Field struct {
	Name       string
	Type       fieldType
	Tags       []*FieldTag
	Alias      string
	Num        int
	IsRepeated bool
	Desc       *generator.FieldDescriptor
	Map        *MapFields
	Slice      *descriptor.FieldDescriptorProto
	File       *generator.FileDescriptor
}

type MapFields struct {
	Key   *descriptor.FieldDescriptorProto
	Value *descriptor.FieldDescriptorProto
}

type Schema struct {
	Name    string
	PK      *Field
	MFields map[string]*Field
	Fields  []*Field
	SDField *Field
	Desc    *generator.MessageDescriptor
}

type FieldTag struct {
	Key    string
	Values []string
	Seq    string
}

func (t *FieldTag) String() string {
	return fmt.Sprintf(`%s:"%s"`, t.Key, strings.Join(t.Values, t.Seq))
}

func MargeTags(tags ...*FieldTag) string {
	outs := make([]string, len(tags))
	for i := range tags {
		outs[i] = tags[i].String()
	}
	return toQuoted(strings.Join(outs, " "))
}
