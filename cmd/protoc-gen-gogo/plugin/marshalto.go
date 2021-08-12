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

/*
The marshalto plugin generates a Marshal and MarshalTo method for each message.
The `Marshal() ([]byte, error)` method results in the fact that the message
implements the Marshaler interface.
This allows proto.Marshal to be faster by calling the generated Marshal method rather than using reflect to Marshal the struct.

If is enabled by the following extensions:

  - marshaler
  - marshaler_all

Or the following extensions:

  - unsafe_marshaler
  - unsafe_marshaler_all

That is if you want to use the unsafe package in your generated code.
The speed up using the unsafe package is not very significant.

The generation of marshalling tests are enabled using one of the following extensions:

  - testgen
  - testgen_all

And benchmarks given it is enabled using one of the following extensions:

  - benchgen
  - benchgen_all

Let us look at:

  github.com/gogo/protobuf/test/example/example.proto

Btw all the output can be seen at:

  github.com/gogo/protobuf/test/example/*

The following message:

option (gogoproto.marshaler_all) = true;

message B {
	option (gogoproto.description) = true;
	optional A A = 1 [(gogoproto.nullable) = false, (gogoproto.embed) = true];
	repeated bytes G = 2 [(gogoproto.customtype) = "github.com/gogo/protobuf/test/custom.Uint128", (gogoproto.nullable) = false];
}

given to the marshalto plugin, will generate the following code:

  func (m *B) Marshal() (dAtA []byte, err error) {
          size := m.Size()
          dAtA = make([]byte, size)
          n, err := m.MarshalToSizedBuffer(dAtA[:size])
          if err != nil {
                  return nil, err
          }
          return dAtA[:n], nil
  }

  func (m *B) MarshalTo(dAtA []byte) (int, error) {
          size := m.Size()
          return m.MarshalToSizedBuffer(dAtA[:size])
  }

  func (m *B) MarshalToSizedBuffer(dAtA []byte) (int, error) {
          i := len(dAtA)
          _ = i
          var l int
          _ = l
          if m.XXX_unrecognized != nil {
                  i -= len(m.XXX_unrecognized)
                  copy(dAtA[i:], m.XXX_unrecognized)
          }
          if len(m.G) > 0 {
                  for iNdEx := len(m.G) - 1; iNdEx >= 0; iNdEx-- {
                          {
                                  size := m.G[iNdEx].Size()
                                  i -= size
                                  if _, err := m.G[iNdEx].MarshalTo(dAtA[i:]); err != nil {
                                          return 0, err
                                  }
                                  i = encodeVarintExample(dAtA, i, uint64(size))
                          }
                          i--
                          dAtA[i] = 0x12
                  }
          }
          {
                  size, err := m.A.MarshalToSizedBuffer(dAtA[:i])
                  if err != nil {
                          return 0, err
                  }
                  i -= size
                  i = encodeVarintExample(dAtA, i, uint64(size))
          }
          i--
          dAtA[i] = 0xa
          return len(dAtA) - i, nil
  }

As shown above Marshal calculates the size of the not yet marshalled message
and allocates the appropriate buffer.
This is followed by calling the MarshalToSizedBuffer method which requires a preallocated buffer, and marshals backwards.
The MarshalTo method allows a user to rather preallocated a reusable buffer.

The Size method is generated using the size plugin and the gogoproto.sizer, gogoproto.sizer_all extensions.
The user can also using the generated Size method to check that his reusable buffer is still big enough.

The generated tests and benchmarks will keep you safe and show that this is really a significant speed improvement.

An additional message-level option `stable_marshaler` (and the file-level
option `stable_marshaler_all`) exists which causes the generated marshalling
code to behave deterministically. Today, this only changes the serialization of
maps; they are serialized in sort order.
*/
package gogo

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	vanity2 "github.com/vine-io/vine/cmd/generator/vanity"

	"github.com/vine-io/vine/cmd/generator"
)

type NumGen interface {
	Next() string
	Current() string
}

type numGen struct {
	index int
}

func NewNumGen() NumGen {
	return &numGen{0}
}

func (this *numGen) Next() string {
	this.index++
	return this.Current()
}

func (this *numGen) Current() string {
	return strconv.Itoa(this.index)
}

func (g *gogo) callFixed64(varName ...string) {
	g.P(`i -= 8`)
	g.P(g.binaryPkg.Use(), `.LittleEndian.PutUint64(dAtA[i:], uint64(`, strings.Join(varName, ""), `))`)
}

func (g *gogo) callFixed32(varName ...string) {
	g.P(`i -= 4`)
	g.P(g.binaryPkg.Use(), `.LittleEndian.PutUint32(dAtA[i:], uint32(`, strings.Join(varName, ""), `))`)
}

func (g *gogo) callVarint(varName ...string) {
	g.P(`i = encodeVarint`, g.localName, `(dAtA, i, uint64(`, strings.Join(varName, ""), `))`)
}

func (g *gogo) encodeKey(fieldNumber int32, wireType int) {
	x := uint32(fieldNumber)<<3 | uint32(wireType)
	i := 0
	keybuf := make([]byte, 0)
	for i = 0; x > 127; i++ {
		keybuf = append(keybuf, 0x80|uint8(x&0x7F))
		x >>= 7
	}
	keybuf = append(keybuf, uint8(x))
	for i = len(keybuf) - 1; i >= 0; i-- {
		g.P(`i--`)
		g.P(`dAtA[i] = `, fmt.Sprintf("%#v", keybuf[i]))
	}
}

func keySize(fieldNumber int32, wireType int) int {
	x := uint32(fieldNumber)<<3 | uint32(wireType)
	size := 0
	for size = 0; x > 127; size++ {
		x >>= 7
	}
	size++
	return size
}

func (g *gogo) mapField(numGen NumGen, field *descriptor.FieldDescriptorProto, kvField *descriptor.FieldDescriptorProto, varName string, protoSizer bool) {
	switch kvField.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		g.callFixed64(g.mathPkg.Use(), `.Float64bits(float64(`, varName, `))`)
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		g.callFixed32(g.mathPkg.Use(), `.Float32bits(float32(`, varName, `))`)
	case descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_ENUM:
		g.callVarint(varName)
	case descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		g.callFixed64(varName)
	case descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		g.callFixed32(varName)
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		g.P(`i--`)
		g.P(`if `, varName, ` {`)
		g.In()
		g.P(`dAtA[i] = 1`)
		g.Out()
		g.P(`} else {`)
		g.In()
		g.P(`dAtA[i] = 0`)
		g.Out()
		g.P(`}`)
	case descriptor.FieldDescriptorProto_TYPE_STRING,
		descriptor.FieldDescriptorProto_TYPE_BYTES:
		if gogoproto.IsCustomType(field) && kvField.IsBytes() {
			g.forward(varName, true, protoSizer)
		} else {
			g.P(`i -= len(`, varName, `)`)
			g.P(`copy(dAtA[i:], `, varName, `)`)
			g.callVarint(`len(`, varName, `)`)
		}
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		g.callVarint(`(uint32(`, varName, `) << 1) ^ uint32((`, varName, ` >> 31))`)
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		g.callVarint(`(uint64(`, varName, `) << 1) ^ uint64((`, varName, ` >> 63))`)
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		if !g.marshalAllSizeOf(kvField, `(*`+varName+`)`, numGen.Next()) {
			if gogoproto.IsCustomType(field) {
				g.forward(varName, true, protoSizer)
			} else {
				g.backward(varName, true)
			}
		}

	}
}

type orderFields []*generator.FieldDescriptor

func (this orderFields) Len() int {
	return len(this)
}

func (this orderFields) Less(i, j int) bool {
	return this[i].Proto.GetNumber() < this[j].Proto.GetNumber()
}

func (this orderFields) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (g *gogo) generateMarshalField(proto3 bool, numGen NumGen, file *generator.FileDescriptor, message *generator.MessageDescriptor, f *generator.FieldDescriptor) {
	field := f.Proto
	fieldname := g.GetOneOfFieldName(message, f)
	nullable := gogoproto.IsNullable(field)
	repeated := field.IsRepeated()
	required := field.IsRequired()
	_, isInline := g.extractTags(f.Comments)[_inline]
	protoSizer := gogoproto.IsProtoSizer(file.FileDescriptorProto, message.Proto.DescriptorProto)
	doNilCheck := gogoproto.NeedsNilCheck(proto3, field)
	if required && nullable {
		g.P(`if m.`, fieldname, `== nil {`)
		g.In()
		if !gogoproto.ImportsGoGoProto(file.FileDescriptorProto) {
			g.P(`return 0, new(`, g.protoPkg.Use(), `.RequiredNotSetError)`)
		} else {
			g.P(`return 0, `, g.protoPkg.Use(), `.NewRequiredNotSetError("`, field.GetName(), `")`)
		}
		g.Out()
		g.P(`} else {`)
	} else if repeated {
		g.P(`if len(m.`, fieldname, `) > 0 {`)
		g.In()
	} else if doNilCheck {
		if isInline {
			g.P(`{`)
		} else {
			g.P(`if m.`, fieldname, ` != nil {`)
		}
		g.In()
	}
	packed := field.IsPacked() || (proto3 && field.IsPacked3())
	wireType := field.WireType()
	fieldNumber := field.GetNumber()
	if packed {
		wireType = proto.WireBytes
	}
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		if packed {
			val := g.reverseListRange(`m.`, fieldname)
			g.P(`f`, numGen.Next(), ` := `, g.mathPkg.Use(), `.Float64bits(float64(`, val, `))`)
			g.callFixed64("f" + numGen.Current())
			g.Out()
			g.P(`}`)
			g.callVarint(`len(m.`, fieldname, `) * 8`)
			g.encodeKey(fieldNumber, wireType)
		} else if repeated {
			val := g.reverseListRange(`m.`, fieldname)
			g.P(`f`, numGen.Next(), ` := `, g.mathPkg.Use(), `.Float64bits(float64(`, val, `))`)
			g.callFixed64("f" + numGen.Current())
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if proto3 {
			g.P(`if m.`, fieldname, ` != 0 {`)
			g.In()
			g.callFixed64(g.mathPkg.Use(), `.Float64bits(float64(m.`+fieldname, `))`)
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if !nullable {
			g.callFixed64(g.mathPkg.Use(), `.Float64bits(float64(m.`+fieldname, `))`)
			g.encodeKey(fieldNumber, wireType)
		} else {
			g.callFixed64(g.mathPkg.Use(), `.Float64bits(float64(*m.`+fieldname, `))`)
			g.encodeKey(fieldNumber, wireType)
		}
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		if packed {
			val := g.reverseListRange(`m.`, fieldname)
			g.P(`f`, numGen.Next(), ` := `, g.mathPkg.Use(), `.Float32bits(float32(`, val, `))`)
			g.callFixed32("f" + numGen.Current())
			g.Out()
			g.P(`}`)
			g.callVarint(`len(m.`, fieldname, `) * 4`)
			g.encodeKey(fieldNumber, wireType)
		} else if repeated {
			val := g.reverseListRange(`m.`, fieldname)
			g.P(`f`, numGen.Next(), ` := `, g.mathPkg.Use(), `.Float32bits(float32(`, val, `))`)
			g.callFixed32("f" + numGen.Current())
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if proto3 {
			g.P(`if m.`, fieldname, ` != 0 {`)
			g.In()
			g.callFixed32(g.mathPkg.Use(), `.Float32bits(float32(m.`+fieldname, `))`)
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if !nullable {
			g.callFixed32(g.mathPkg.Use(), `.Float32bits(float32(m.`+fieldname, `))`)
			g.encodeKey(fieldNumber, wireType)
		} else {
			g.callFixed32(g.mathPkg.Use(), `.Float32bits(float32(*m.`+fieldname, `))`)
			g.encodeKey(fieldNumber, wireType)
		}
	case descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_ENUM:
		if packed {
			jvar := "j" + numGen.Next()
			g.P(`dAtA`, numGen.Next(), ` := make([]byte, len(m.`, fieldname, `)*10)`)
			g.P(`var `, jvar, ` int`)
			if *field.Type == descriptor.FieldDescriptorProto_TYPE_INT64 ||
				*field.Type == descriptor.FieldDescriptorProto_TYPE_INT32 {
				g.P(`for _, num1 := range m.`, fieldname, ` {`)
				g.In()
				g.P(`num := uint64(num1)`)
			} else {
				g.P(`for _, num := range m.`, fieldname, ` {`)
				g.In()
			}
			g.P(`for num >= 1<<7 {`)
			g.In()
			g.P(`dAtA`, numGen.Current(), `[`, jvar, `] = uint8(uint64(num)&0x7f|0x80)`)
			g.P(`num >>= 7`)
			g.P(jvar, `++`)
			g.Out()
			g.P(`}`)
			g.P(`dAtA`, numGen.Current(), `[`, jvar, `] = uint8(num)`)
			g.P(jvar, `++`)
			g.Out()
			g.P(`}`)
			g.P(`i -= `, jvar)
			g.P(`copy(dAtA[i:], dAtA`, numGen.Current(), `[:`, jvar, `])`)
			g.callVarint(jvar)
			g.encodeKey(fieldNumber, wireType)
		} else if repeated {
			val := g.reverseListRange(`m.`, fieldname)
			g.callVarint(val)
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if proto3 {
			g.P(`if m.`, fieldname, ` != 0 {`)
			g.In()
			g.callVarint(`m.`, fieldname)
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if !nullable {
			g.callVarint(`m.`, fieldname)
			g.encodeKey(fieldNumber, wireType)
		} else {
			g.callVarint(`*m.`, fieldname)
			g.encodeKey(fieldNumber, wireType)
		}
	case descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		if packed {
			val := g.reverseListRange(`m.`, fieldname)
			g.callFixed64(val)
			g.Out()
			g.P(`}`)
			g.callVarint(`len(m.`, fieldname, `) * 8`)
			g.encodeKey(fieldNumber, wireType)
		} else if repeated {
			val := g.reverseListRange(`m.`, fieldname)
			g.callFixed64(val)
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if proto3 {
			g.P(`if m.`, fieldname, ` != 0 {`)
			g.In()
			g.callFixed64("m." + fieldname)
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if !nullable {
			g.callFixed64("m." + fieldname)
			g.encodeKey(fieldNumber, wireType)
		} else {
			g.callFixed64("*m." + fieldname)
			g.encodeKey(fieldNumber, wireType)
		}
	case descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		if packed {
			val := g.reverseListRange(`m.`, fieldname)
			g.callFixed32(val)
			g.Out()
			g.P(`}`)
			g.callVarint(`len(m.`, fieldname, `) * 4`)
			g.encodeKey(fieldNumber, wireType)
		} else if repeated {
			val := g.reverseListRange(`m.`, fieldname)
			g.callFixed32(val)
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if proto3 {
			g.P(`if m.`, fieldname, ` != 0 {`)
			g.In()
			g.callFixed32("m." + fieldname)
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if !nullable {
			g.callFixed32("m." + fieldname)
			g.encodeKey(fieldNumber, wireType)
		} else {
			g.callFixed32("*m." + fieldname)
			g.encodeKey(fieldNumber, wireType)
		}
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		if packed {
			val := g.reverseListRange(`m.`, fieldname)
			g.P(`i--`)
			g.P(`if `, val, ` {`)
			g.In()
			g.P(`dAtA[i] = 1`)
			g.Out()
			g.P(`} else {`)
			g.In()
			g.P(`dAtA[i] = 0`)
			g.Out()
			g.P(`}`)
			g.Out()
			g.P(`}`)
			g.callVarint(`len(m.`, fieldname, `)`)
			g.encodeKey(fieldNumber, wireType)
		} else if repeated {
			val := g.reverseListRange(`m.`, fieldname)
			g.P(`i--`)
			g.P(`if `, val, ` {`)
			g.In()
			g.P(`dAtA[i] = 1`)
			g.Out()
			g.P(`} else {`)
			g.In()
			g.P(`dAtA[i] = 0`)
			g.Out()
			g.P(`}`)
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if proto3 {
			g.P(`if m.`, fieldname, ` {`)
			g.In()
			g.P(`i--`)
			g.P(`if m.`, fieldname, ` {`)
			g.In()
			g.P(`dAtA[i] = 1`)
			g.Out()
			g.P(`} else {`)
			g.In()
			g.P(`dAtA[i] = 0`)
			g.Out()
			g.P(`}`)
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if !nullable {
			g.P(`i--`)
			g.P(`if m.`, fieldname, ` {`)
			g.In()
			g.P(`dAtA[i] = 1`)
			g.Out()
			g.P(`} else {`)
			g.In()
			g.P(`dAtA[i] = 0`)
			g.Out()
			g.P(`}`)
			g.encodeKey(fieldNumber, wireType)
		} else {
			g.P(`i--`)
			g.P(`if *m.`, fieldname, ` {`)
			g.In()
			g.P(`dAtA[i] = 1`)
			g.Out()
			g.P(`} else {`)
			g.In()
			g.P(`dAtA[i] = 0`)
			g.Out()
			g.P(`}`)
			g.encodeKey(fieldNumber, wireType)
		}
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		if repeated {
			val := g.reverseListRange(`m.`, fieldname)
			g.P(`i -= len(`, val, `)`)
			g.P(`copy(dAtA[i:], `, val, `)`)
			g.callVarint(`len(`, val, `)`)
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if proto3 {
			g.P(`if len(m.`, fieldname, `) > 0 {`)
			g.In()
			g.P(`i -= len(m.`, fieldname, `)`)
			g.P(`copy(dAtA[i:], m.`, fieldname, `)`)
			g.callVarint(`len(m.`, fieldname, `)`)
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if !nullable {
			g.P(`i -= len(m.`, fieldname, `)`)
			g.P(`copy(dAtA[i:], m.`, fieldname, `)`)
			g.callVarint(`len(m.`, fieldname, `)`)
			g.encodeKey(fieldNumber, wireType)
		} else {
			g.P(`i -= len(*m.`, fieldname, `)`)
			g.P(`copy(dAtA[i:], *m.`, fieldname, `)`)
			g.callVarint(`len(*m.`, fieldname, `)`)
			g.encodeKey(fieldNumber, wireType)
		}
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		panic(fmt.Errorf("marshaler does not support group %v", fieldname))
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		if g.IsMap(field) {
			m := g.GoMapType(nil, field)
			keygoTyp, keywire := g.GoType(nil, m.KeyField)
			keygoAliasTyp, _ := g.GoType(nil, m.KeyAliasField)
			// keys may not be pointers
			keygoTyp = strings.Replace(keygoTyp, "*", "", 1)
			keygoAliasTyp = strings.Replace(keygoAliasTyp, "*", "", 1)
			keyCapTyp := generator.CamelCase(keygoTyp)
			valuegoTyp, valuewire := g.GoType(nil, m.ValueField)
			valuegoAliasTyp, _ := g.GoType(nil, m.ValueAliasField)
			nullable, valuegoTyp, valuegoAliasTyp = generator.GoMapValueTypes(field, m.ValueField, valuegoTyp, valuegoAliasTyp)
			var val string
			if gogoproto.IsStableMarshaler(file.FileDescriptorProto, message.Proto.DescriptorProto) {
				keysName := `keysFor` + fieldname
				g.P(keysName, ` := make([]`, keygoTyp, `, 0, len(m.`, fieldname, `))`)
				g.P(`for k := range m.`, fieldname, ` {`)
				g.In()
				g.P(keysName, ` = append(`, keysName, `, `, keygoTyp, `(k))`)
				g.Out()
				g.P(`}`)
				g.P(g.sortKeysPkg.Use(), `.`, keyCapTyp, `s(`, keysName, `)`)
				val = g.reverseListRange(keysName)
			} else {
				g.P(`for k := range m.`, fieldname, ` {`)
				val = "k"
				g.In()
			}
			if gogoproto.IsStableMarshaler(file.FileDescriptorProto, message.Proto.DescriptorProto) {
				g.P(`v := m.`, fieldname, `[`, keygoAliasTyp, `(`, val, `)]`)
			} else {
				g.P(`v := m.`, fieldname, `[`, val, `]`)
			}
			g.P(`baseI := i`)
			accessor := `v`

			if m.ValueField.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
				if valuegoTyp != valuegoAliasTyp && !gogoproto.IsStdType(m.ValueAliasField) {
					if nullable {
						// cast back to the type that has the generated methods on it
						accessor = `((` + valuegoTyp + `)(` + accessor + `))`
					} else {
						accessor = `((*` + valuegoTyp + `)(&` + accessor + `))`
					}
				} else if !nullable {
					accessor = `(&v)`
				}
			}

			nullableMsg := nullable && (m.ValueField.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE ||
				gogoproto.IsCustomType(field) && m.ValueField.IsBytes())
			plainBytes := m.ValueField.IsBytes() && !gogoproto.IsCustomType(field)
			if nullableMsg {
				g.P(`if `, accessor, ` != nil { `)
				g.In()
			} else if plainBytes {
				if proto3 {
					g.P(`if len(`, accessor, `) > 0 {`)
				} else {
					g.P(`if `, accessor, ` != nil {`)
				}
				g.In()
			}
			g.mapField(numGen, field, m.ValueAliasField, accessor, protoSizer)
			g.encodeKey(2, wireToType(valuewire))
			if nullableMsg || plainBytes {
				g.Out()
				g.P(`}`)
			}

			g.mapField(numGen, field, m.KeyField, val, protoSizer)
			g.encodeKey(1, wireToType(keywire))

			g.callVarint(`baseI - i`)

			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if repeated {
			val := g.reverseListRange(`m.`, fieldname)
			sizeOfVarName := val
			if gogoproto.IsNullable(field) {
				sizeOfVarName = `*` + val
			}
			if !g.marshalAllSizeOf(field, sizeOfVarName, ``) {
				if gogoproto.IsCustomType(field) {
					g.forward(val, true, protoSizer)
				} else {
					g.backward(val, true)
				}
			}
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else {
			sizeOfVarName := `m.` + fieldname
			if gogoproto.IsNullable(field) {
				sizeOfVarName = `*` + sizeOfVarName
			}
			if !g.marshalAllSizeOf(field, sizeOfVarName, numGen.Next()) {
				if gogoproto.IsCustomType(field) {
					g.forward(`m.`+fieldname, true, protoSizer)
				} else {
					g.backward(`m.`+fieldname, true)
				}
			}
			g.encodeKey(fieldNumber, wireType)
		}
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		if !gogoproto.IsCustomType(field) {
			if repeated {
				val := g.reverseListRange(`m.`, fieldname)
				g.P(`i -= len(`, val, `)`)
				g.P(`copy(dAtA[i:], `, val, `)`)
				g.callVarint(`len(`, val, `)`)
				g.encodeKey(fieldNumber, wireType)
				g.Out()
				g.P(`}`)
			} else if proto3 {
				g.P(`if len(m.`, fieldname, `) > 0 {`)
				g.In()
				g.P(`i -= len(m.`, fieldname, `)`)
				g.P(`copy(dAtA[i:], m.`, fieldname, `)`)
				g.callVarint(`len(m.`, fieldname, `)`)
				g.encodeKey(fieldNumber, wireType)
				g.Out()
				g.P(`}`)
			} else {
				g.P(`i -= len(m.`, fieldname, `)`)
				g.P(`copy(dAtA[i:], m.`, fieldname, `)`)
				g.callVarint(`len(m.`, fieldname, `)`)
				g.encodeKey(fieldNumber, wireType)
			}
		} else {
			if repeated {
				val := g.reverseListRange(`m.`, fieldname)
				g.forward(val, true, protoSizer)
				g.encodeKey(fieldNumber, wireType)
				g.Out()
				g.P(`}`)
			} else {
				g.forward(`m.`+fieldname, true, protoSizer)
				g.encodeKey(fieldNumber, wireType)
			}
		}
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		if packed {
			datavar := "dAtA" + numGen.Next()
			jvar := "j" + numGen.Next()
			g.P(datavar, ` := make([]byte, len(m.`, fieldname, ")*5)")
			g.P(`var `, jvar, ` int`)
			g.P(`for _, num := range m.`, fieldname, ` {`)
			g.In()
			xvar := "x" + numGen.Next()
			g.P(xvar, ` := (uint32(num) << 1) ^ uint32((num >> 31))`)
			g.P(`for `, xvar, ` >= 1<<7 {`)
			g.In()
			g.P(datavar, `[`, jvar, `] = uint8(uint64(`, xvar, `)&0x7f|0x80)`)
			g.P(jvar, `++`)
			g.P(xvar, ` >>= 7`)
			g.Out()
			g.P(`}`)
			g.P(datavar, `[`, jvar, `] = uint8(`, xvar, `)`)
			g.P(jvar, `++`)
			g.Out()
			g.P(`}`)
			g.P(`i -= `, jvar)
			g.P(`copy(dAtA[i:], `, datavar, `[:`, jvar, `])`)
			g.callVarint(jvar)
			g.encodeKey(fieldNumber, wireType)
		} else if repeated {
			val := g.reverseListRange(`m.`, fieldname)
			g.P(`x`, numGen.Next(), ` := (uint32(`, val, `) << 1) ^ uint32((`, val, ` >> 31))`)
			g.callVarint(`x`, numGen.Current())
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if proto3 {
			g.P(`if m.`, fieldname, ` != 0 {`)
			g.In()
			g.callVarint(`(uint32(m.`, fieldname, `) << 1) ^ uint32((m.`, fieldname, ` >> 31))`)
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if !nullable {
			g.callVarint(`(uint32(m.`, fieldname, `) << 1) ^ uint32((m.`, fieldname, ` >> 31))`)
			g.encodeKey(fieldNumber, wireType)
		} else {
			g.callVarint(`(uint32(*m.`, fieldname, `) << 1) ^ uint32((*m.`, fieldname, ` >> 31))`)
			g.encodeKey(fieldNumber, wireType)
		}
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		if packed {
			jvar := "j" + numGen.Next()
			xvar := "x" + numGen.Next()
			datavar := "dAtA" + numGen.Next()
			g.P(`var `, jvar, ` int`)
			g.P(datavar, ` := make([]byte, len(m.`, fieldname, `)*10)`)
			g.P(`for _, num := range m.`, fieldname, ` {`)
			g.In()
			g.P(xvar, ` := (uint64(num) << 1) ^ uint64((num >> 63))`)
			g.P(`for `, xvar, ` >= 1<<7 {`)
			g.In()
			g.P(datavar, `[`, jvar, `] = uint8(uint64(`, xvar, `)&0x7f|0x80)`)
			g.P(jvar, `++`)
			g.P(xvar, ` >>= 7`)
			g.Out()
			g.P(`}`)
			g.P(datavar, `[`, jvar, `] = uint8(`, xvar, `)`)
			g.P(jvar, `++`)
			g.Out()
			g.P(`}`)
			g.P(`i -= `, jvar)
			g.P(`copy(dAtA[i:], `, datavar, `[:`, jvar, `])`)
			g.callVarint(jvar)
			g.encodeKey(fieldNumber, wireType)
		} else if repeated {
			val := g.reverseListRange(`m.`, fieldname)
			g.P(`x`, numGen.Next(), ` := (uint64(`, val, `) << 1) ^ uint64((`, val, ` >> 63))`)
			g.callVarint("x" + numGen.Current())
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if proto3 {
			g.P(`if m.`, fieldname, ` != 0 {`)
			g.In()
			g.callVarint(`(uint64(m.`, fieldname, `) << 1) ^ uint64((m.`, fieldname, ` >> 63))`)
			g.encodeKey(fieldNumber, wireType)
			g.Out()
			g.P(`}`)
		} else if !nullable {
			g.callVarint(`(uint64(m.`, fieldname, `) << 1) ^ uint64((m.`, fieldname, ` >> 63))`)
			g.encodeKey(fieldNumber, wireType)
		} else {
			g.callVarint(`(uint64(*m.`, fieldname, `) << 1) ^ uint64((*m.`, fieldname, ` >> 63))`)
			g.encodeKey(fieldNumber, wireType)
		}
	default:
		panic("not implemented")
	}
	if (required && nullable) || repeated || doNilCheck {
		g.Out()
		g.P(`}`)
	}
}

func (g *gogo) GenerateMarshal(file *generator.FileDescriptor) {
	numGen := NewNumGen()

	for _, message := range file.Messages() {
		if message.Proto.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}
		ccTypeName := generator.CamelCaseSlice(message.Proto.TypeName())
		if !gogoproto.IsMarshaler(file.FileDescriptorProto, message.Proto.DescriptorProto) &&
			!gogoproto.IsUnsafeMarshaler(file.FileDescriptorProto, message.Proto.DescriptorProto) {
			continue
		}
		g.atleastOne = true

		g.P(`func (m *`, ccTypeName, `) Marshal() (dAtA []byte, err error) {`)
		g.In()
		if gogoproto.IsProtoSizer(file.FileDescriptorProto, message.Proto.DescriptorProto) {
			g.P(`size := m.ProtoSize()`)
		} else {
			g.P(`size := m.XSize()`)
		}
		g.P(`dAtA = make([]byte, size)`)
		g.P(`n, err := m.MarshalToSizedBuffer(dAtA[:size])`)
		g.P(`if err != nil {`)
		g.In()
		g.P(`return nil, err`)
		g.Out()
		g.P(`}`)
		g.P(`return dAtA[:n], nil`)
		g.Out()
		g.P(`}`)
		g.P(``)
		g.P(`func (m *`, ccTypeName, `) MarshalTo(dAtA []byte) (int, error) {`)
		g.In()
		if gogoproto.IsProtoSizer(file.FileDescriptorProto, message.Proto.DescriptorProto) {
			g.P(`size := m.ProtoSize()`)
		} else {
			g.P(`size := m.XSize()`)
		}
		g.P(`return m.MarshalToSizedBuffer(dAtA[:size])`)
		g.Out()
		g.P(`}`)
		g.P(``)
		g.P(`func (m *`, ccTypeName, `) MarshalToSizedBuffer(dAtA []byte) (int, error) {`)
		g.In()
		g.P(`i := len(dAtA)`)
		g.P(`_ = i`)
		g.P(`var l int`)
		g.P(`_ = l`)
		if gogoproto.HasUnrecognized(file.FileDescriptorProto, message.Proto.DescriptorProto) {
			g.P(`if m.XXX_unrecognized != nil {`)
			g.In()
			g.P(`i -= len(m.XXX_unrecognized)`)
			g.P(`copy(dAtA[i:], m.XXX_unrecognized)`)
			g.Out()
			g.P(`}`)
		}
		if message.Proto.DescriptorProto.HasExtension() {
			if gogoproto.HasExtensionsMap(file.FileDescriptorProto, message.Proto.DescriptorProto) {
				g.P(`if n, err := `, g.protoPkg.Use(), `.EncodeInternalExtensionBackwards(m, dAtA[:i]); err != nil {`)
				g.In()
				g.P(`return 0, err`)
				g.Out()
				g.P(`} else {`)
				g.In()
				g.P(`i -= n`)
				g.Out()
				g.P(`}`)
			} else {
				g.P(`if m.XXX_extensions != nil {`)
				g.In()
				g.P(`i -= len(m.XXX_extensions)`)
				g.P(`copy(dAtA[i:], m.XXX_extensions)`)
				g.Out()
				g.P(`}`)
			}
		}
		fields := orderFields(message.Fields)
		sort.Sort(fields)
		oneofs := make(map[string]struct{})
		for i := len(message.Fields) - 1; i >= 0; i-- {
			field := message.Fields[i]
			oneof := field.Proto.OneofIndex != nil
			if !oneof {
				proto3 := gogoproto.IsProto3(file.FileDescriptorProto)
				g.generateMarshalField(proto3, numGen, file, message, field)
			} else {
				fieldname := g.GetFieldName(message.Proto, field)
				if _, ok := oneofs[fieldname]; !ok {
					oneofs[fieldname] = struct{}{}
					g.P(`if m.`, fieldname, ` != nil {`)
					g.In()
					g.forward(`m.`+fieldname, false, gogoproto.IsProtoSizer(file.FileDescriptorProto, message.Proto.DescriptorProto))
					g.Out()
					g.P(`}`)
				}
			}
		}
		g.P(`return len(dAtA) - i, nil`)
		g.Out()
		g.P(`}`)
		g.P()

		//Generate MarshalTo methods for oneof fields
		//m := proto.Clone(message.Proto.DescriptorProto).(*descriptor.DescriptorProto)
		for _, field := range message.Fields {
			oneof := field.Proto.OneofIndex != nil
			if !oneof {
				continue
			}
			ccTypeName := g.OneOfTypeName(message, field)
			g.P(`func (m *`, ccTypeName, `) MarshalTo(dAtA []byte) (int, error) {`)
			g.In()
			if gogoproto.IsProtoSizer(file.FileDescriptorProto, message.Proto.DescriptorProto) {
				g.P(`size := m.ProtoSize()`)
			} else {
				g.P(`size := m.Size()`)
			}
			g.P(`return m.MarshalToSizedBuffer(dAtA[:size])`)
			g.Out()
			g.P(`}`)
			g.P(``)
			g.P(`func (m *`, ccTypeName, `) MarshalToSizedBuffer(dAtA []byte) (int, error) {`)
			g.In()
			g.P(`i := len(dAtA)`)
			vanity2.TurnOffNullableForNativeTypes(field.Proto)
			g.generateMarshalField(false, numGen, file, message, field)
			g.P(`return len(dAtA) - i, nil`)
			g.Out()
			g.P(`}`)
		}
	}

	if g.atleastOne {
		g.P(`func encodeVarint`, g.localName, `(dAtA []byte, offset int, v uint64) int {`)
		g.In()
		g.P(`offset -= sov`, g.localName, `(v)`)
		g.P(`base := offset`)
		g.P(`for v >= 1<<7 {`)
		g.In()
		g.P(`dAtA[offset] = uint8(v&0x7f|0x80)`)
		g.P(`v >>= 7`)
		g.P(`offset++`)
		g.Out()
		g.P(`}`)
		g.P(`dAtA[offset] = uint8(v)`)
		g.P(`return base`)
		g.Out()
		g.P(`}`)
	}

}

func (g *gogo) reverseListRange(expression ...string) string {
	exp := strings.Join(expression, "")
	g.P(`for iNdEx := len(`, exp, `) - 1; iNdEx >= 0; iNdEx-- {`)
	g.In()
	return exp + `[iNdEx]`
}

func (g *gogo) marshalAllSizeOf(field *descriptor.FieldDescriptorProto, varName, num string) bool {
	if gogoproto.IsStdTime(field) {
		g.marshalSizeOf(`StdTimeMarshalTo`, `SizeOfStdTime`, varName, num)
	} else if gogoproto.IsStdDuration(field) {
		g.marshalSizeOf(`StdDurationMarshalTo`, `SizeOfStdDuration`, varName, num)
	} else if gogoproto.IsStdDouble(field) {
		g.marshalSizeOf(`StdDoubleMarshalTo`, `SizeOfStdDouble`, varName, num)
	} else if gogoproto.IsStdFloat(field) {
		g.marshalSizeOf(`StdFloatMarshalTo`, `SizeOfStdFloat`, varName, num)
	} else if gogoproto.IsStdInt64(field) {
		g.marshalSizeOf(`StdInt64MarshalTo`, `SizeOfStdInt64`, varName, num)
	} else if gogoproto.IsStdUInt64(field) {
		g.marshalSizeOf(`StdUInt64MarshalTo`, `SizeOfStdUInt64`, varName, num)
	} else if gogoproto.IsStdInt32(field) {
		g.marshalSizeOf(`StdInt32MarshalTo`, `SizeOfStdInt32`, varName, num)
	} else if gogoproto.IsStdUInt32(field) {
		g.marshalSizeOf(`StdUInt32MarshalTo`, `SizeOfStdUInt32`, varName, num)
	} else if gogoproto.IsStdBool(field) {
		g.marshalSizeOf(`StdBoolMarshalTo`, `SizeOfStdBool`, varName, num)
	} else if gogoproto.IsStdString(field) {
		g.marshalSizeOf(`StdStringMarshalTo`, `SizeOfStdString`, varName, num)
	} else if gogoproto.IsStdBytes(field) {
		g.marshalSizeOf(`StdBytesMarshalTo`, `SizeOfStdBytes`, varName, num)
	} else {
		return false
	}
	return true
}

func (g *gogo) marshalSizeOf(marshal, size, varName, num string) {
	g.P(`n`, num, `, err`, num, ` := `, g.typesPkg.Use(), `.`, marshal, `(`, varName, `, dAtA[i-`, g.typesPkg.Use(), `.`, size, `(`, varName, `):])`)
	g.P(`if err`, num, ` != nil {`)
	g.In()
	g.P(`return 0, err`, num)
	g.Out()
	g.P(`}`)
	g.P(`i -= n`, num)
	g.callVarint(`n`, num)
}

func (g *gogo) backward(varName string, varInt bool) {
	g.P(`{`)
	g.In()
	g.P(`size, err := `, varName, `.MarshalToSizedBuffer(dAtA[:i])`)
	g.P(`if err != nil {`)
	g.In()
	g.P(`return 0, err`)
	g.Out()
	g.P(`}`)
	g.P(`i -= size`)
	if varInt {
		g.callVarint(`size`)
	}
	g.Out()
	g.P(`}`)
}

func (g *gogo) forward(varName string, varInt, protoSizer bool) {
	g.P(`{`)
	g.In()
	if protoSizer {
		g.P(`size := `, varName, `.ProtoSize()`)
	} else {
		g.P(`size := `, varName, `.Size()`)
	}
	g.P(`i -= size`)
	g.P(`if _, err := `, varName, `.MarshalTo(dAtA[i:]); err != nil {`)
	g.In()
	g.P(`return 0, err`)
	g.Out()
	g.P(`}`)
	g.Out()
	if varInt {
		g.callVarint(`size`)
	}
	g.P(`}`)
}
