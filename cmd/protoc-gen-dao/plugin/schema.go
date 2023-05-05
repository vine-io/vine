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

	"github.com/vine-io/vine/cmd/generator"
)

type fieldType string

const (
	_float32 fieldType = "float32"
	_float64 fieldType = "float64"
	_int32   fieldType = "int32"
	_int64   fieldType = "int64"
	_uint32  fieldType = "uint32"
	_uint64  fieldType = "uint64"
	_bool    fieldType = "bool"
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

type Storage struct {
	Name     string
	PK       *Field
	DeleteAt *Field
	MFields  map[string]*Field
	Fields   []*Field
	SDField  *Field
	Desc     *generator.MessageDescriptor
	Table    string
	Deep     bool
	TypeMeta bool
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
