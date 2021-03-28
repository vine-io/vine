// Protocol Buffers for Go with Gadgets
//
// Copyright (c) 2013, The GoGo Authors. All rights reserved.
// http://github.com/gogo/protobuf
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

/*
The unmarshal plugin generates a Unmarshal method for each message.
The `Unmarshal([]byte) error` method results in the fact that the message
implements the Unmarshaler interface.
The allows proto.Unmarshal to be faster by calling the generated Unmarshal method rather than using reflect.

If is enabled by the following extensions:

  - unmarshaler
  - unmarshaler_all

Or the following extensions:

  - unsafe_unmarshaler
  - unsafe_unmarshaler_all

That is if you want to use the unsafe package in your generated code.
The speed up using the unsafe package is not very significant.

The generation of unmarshalling tests are enabled using one of the following extensions:

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

  option (gogoproto.unmarshaler_all) = true;

  message B {
	option (gogoproto.description) = true;
	optional A A = 1 [(gogoproto.nullable) = false, (gogoproto.embed) = true];
	repeated bytes G = 2 [(gogoproto.customtype) = "github.com/gogo/protobuf/test/custom.Uint128", (gogoproto.nullable) = false];
  }

given to the unmarshal plugin, will generate the following code:

  func (m *B) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return proto.ErrWrongType
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.A.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return proto.ErrWrongType
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.G = append(m.G, github_com_gogo_protobuf_test_custom.Uint128{})
			if err := m.G[len(m.G)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			var sizeOfWire int
			for {
				sizeOfWire++
				wire >>= 7
				if wire == 0 {
					break
				}
			}
			iNdEx -= sizeOfWire
			skippy, err := skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}
	return nil
  }

Remember when using this code to call proto.Unmarshal.
This will call m.Reset and invoke the generated Unmarshal method for you.
If you call m.Unmarshal without m.Reset you could be merging protocol buffers.

*/
package gogo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

	"github.com/lack-io/vine/cmd/generator"
)

func (g *gogo) decodeVarint(varName string, typName string) {
	g.P(`for shift := uint(0); ; shift += 7 {`)
	g.In()
	g.P(`if shift >= 64 {`)
	g.In()
	g.P(`return ErrIntOverflow` + g.localName)
	g.Out()
	g.P(`}`)
	g.P(`if iNdEx >= l {`)
	g.In()
	g.P(`return `, g.ioPkg.Use(), `.ErrUnexpectedEOF`)
	g.Out()
	g.P(`}`)
	g.P(`b := dAtA[iNdEx]`)
	g.P(`iNdEx++`)
	g.P(varName, ` |= `, typName, `(b&0x7F) << shift`)
	g.P(`if b < 0x80 {`)
	g.In()
	g.P(`break`)
	g.Out()
	g.P(`}`)
	g.Out()
	g.P(`}`)
}

func (g *gogo) decodeFixed32(varName string, typeName string) {
	g.P(`if (iNdEx+4) > l {`)
	g.In()
	g.P(`return `, g.ioPkg.Use(), `.ErrUnexpectedEOF`)
	g.Out()
	g.P(`}`)
	g.P(varName, ` = `, typeName, `(`, g.binaryPkg.Use(), `.LittleEndian.Uint32(dAtA[iNdEx:]))`)
	g.P(`iNdEx += 4`)
}

func (g *gogo) decodeFixed64(varName string, typeName string) {
	g.P(`if (iNdEx+8) > l {`)
	g.In()
	g.P(`return `, g.ioPkg.Use(), `.ErrUnexpectedEOF`)
	g.Out()
	g.P(`}`)
	g.P(varName, ` = `, typeName, `(`, g.binaryPkg.Use(), `.LittleEndian.Uint64(dAtA[iNdEx:]))`)
	g.P(`iNdEx += 8`)
}

func (g *gogo) declareMapField(varName string, nullable bool, customType bool, field *descriptor.FieldDescriptorProto) {
	switch field.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		g.P(`var `, varName, ` float64`)
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		g.P(`var `, varName, ` float32`)
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		g.P(`var `, varName, ` int64`)
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		g.P(`var `, varName, ` uint64`)
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		g.P(`var `, varName, ` int32`)
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		g.P(`var `, varName, ` uint64`)
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		g.P(`var `, varName, ` uint32`)
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		g.P(`var `, varName, ` bool`)
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		cast, _ := g.GoType(nil, field)
		cast = strings.Replace(cast, "*", "", 1)
		g.P(`var `, varName, ` `, cast)
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		if gogoproto.IsStdTime(field) {
			g.P(varName, ` := new(time.Time)`)
		} else if gogoproto.IsStdDuration(field) {
			g.P(varName, ` := new(time.Duration)`)
		} else if gogoproto.IsStdDouble(field) {
			g.P(varName, ` := new(float64)`)
		} else if gogoproto.IsStdFloat(field) {
			g.P(varName, ` := new(float32)`)
		} else if gogoproto.IsStdInt64(field) {
			g.P(varName, ` := new(int64)`)
		} else if gogoproto.IsStdUInt64(field) {
			g.P(varName, ` := new(uint64)`)
		} else if gogoproto.IsStdInt32(field) {
			g.P(varName, ` := new(int32)`)
		} else if gogoproto.IsStdUInt32(field) {
			g.P(varName, ` := new(uint32)`)
		} else if gogoproto.IsStdBool(field) {
			g.P(varName, ` := new(bool)`)
		} else if gogoproto.IsStdString(field) {
			g.P(varName, ` := new(string)`)
		} else if gogoproto.IsStdBytes(field) {
			g.P(varName, ` := new([]byte)`)
		} else {
			desc := g.ObjectNamed(field.GetTypeName())
			msgname := g.TypeName(desc)
			if nullable {
				g.P(`var `, varName, ` *`, msgname)
			} else {
				g.P(varName, ` := &`, msgname, `{}`)
			}
		}
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		if customType {
			_, ctyp, err := generator.GetCustomType(field)
			if err != nil {
				panic(err)
			}
			g.P(`var `, varName, `1 `, ctyp)
			g.P(`var `, varName, ` = &`, varName, `1`)
		} else {
			g.P(varName, ` := []byte{}`)
		}
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		g.P(`var `, varName, ` uint32`)
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		typName := g.TypeName(g.ObjectNamed(field.GetTypeName()))
		g.P(`var `, varName, ` `, typName)
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		g.P(`var `, varName, ` int32`)
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		g.P(`var `, varName, ` int64`)
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		g.P(`var `, varName, ` int32`)
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		g.P(`var `, varName, ` int64`)
	}
}

func (g *gogo) mapUnmarshalField(varName string, customType bool, field *descriptor.FieldDescriptorProto) {
	switch field.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		g.P(`var `, varName, `temp uint64`)
		g.decodeFixed64(varName+"temp", "uint64")
		g.P(varName, ` = `, g.mathPkg.Use(), `.Float64frombits(`, varName, `temp)`)
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		g.P(`var `, varName, `temp uint32`)
		g.decodeFixed32(varName+"temp", "uint32")
		g.P(varName, ` = `, g.mathPkg.Use(), `.Float32frombits(`, varName, `temp)`)
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		g.decodeVarint(varName, "int64")
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		g.decodeVarint(varName, "uint64")
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		g.decodeVarint(varName, "int32")
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		g.decodeFixed64(varName, "uint64")
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		g.decodeFixed32(varName, "uint32")
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		g.P(`var `, varName, `temp int`)
		g.decodeVarint(varName+"temp", "int")
		g.P(varName, ` = bool(`, varName, `temp != 0)`)
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		g.P(`var stringLen`, varName, ` uint64`)
		g.decodeVarint("stringLen"+varName, "uint64")
		g.P(`intStringLen`, varName, ` := int(stringLen`, varName, `)`)
		g.P(`if intStringLen`, varName, ` < 0 {`)
		g.In()
		g.P(`return ErrInvalidLength` + g.localName)
		g.Out()
		g.P(`}`)
		g.P(`postStringIndex`, varName, ` := iNdEx + intStringLen`, varName)
		g.P(`if postStringIndex`, varName, ` < 0 {`)
		g.In()
		g.P(`return ErrInvalidLength` + g.localName)
		g.Out()
		g.P(`}`)
		g.P(`if postStringIndex`, varName, ` > l {`)
		g.In()
		g.P(`return `, g.ioPkg.Use(), `.ErrUnexpectedEOF`)
		g.Out()
		g.P(`}`)
		cast, _ := g.GoType(nil, field)
		cast = strings.Replace(cast, "*", "", 1)
		g.P(varName, ` = `, cast, `(dAtA[iNdEx:postStringIndex`, varName, `])`)
		g.P(`iNdEx = postStringIndex`, varName)
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		g.P(`var mapmsglen int`)
		g.decodeVarint("mapmsglen", "int")
		g.P(`if mapmsglen < 0 {`)
		g.In()
		g.P(`return ErrInvalidLength` + g.localName)
		g.Out()
		g.P(`}`)
		g.P(`postmsgIndex := iNdEx + mapmsglen`)
		g.P(`if postmsgIndex < 0 {`)
		g.In()
		g.P(`return ErrInvalidLength` + g.localName)
		g.Out()
		g.P(`}`)
		g.P(`if postmsgIndex > l {`)
		g.In()
		g.P(`return `, g.ioPkg.Use(), `.ErrUnexpectedEOF`)
		g.Out()
		g.P(`}`)
		buf := `dAtA[iNdEx:postmsgIndex]`
		if gogoproto.IsStdTime(field) {
			g.P(`if err := `, g.typesPkg.Use(), `.StdTimeUnmarshal(`, varName, `, `, buf, `); err != nil {`)
		} else if gogoproto.IsStdDuration(field) {
			g.P(`if err := `, g.typesPkg.Use(), `.StdDurationUnmarshal(`, varName, `, `, buf, `); err != nil {`)
		} else if gogoproto.IsStdDouble(field) {
			g.P(`if err := `, g.typesPkg.Use(), `.StdDoubleUnmarshal(`, varName, `, `, buf, `); err != nil {`)
		} else if gogoproto.IsStdFloat(field) {
			g.P(`if err := `, g.typesPkg.Use(), `.StdFloatUnmarshal(`, varName, `, `, buf, `); err != nil {`)
		} else if gogoproto.IsStdInt64(field) {
			g.P(`if err := `, g.typesPkg.Use(), `.StdInt64Unmarshal(`, varName, `, `, buf, `); err != nil {`)
		} else if gogoproto.IsStdUInt64(field) {
			g.P(`if err := `, g.typesPkg.Use(), `.StdUInt64Unmarshal(`, varName, `, `, buf, `); err != nil {`)
		} else if gogoproto.IsStdInt32(field) {
			g.P(`if err := `, g.typesPkg.Use(), `.StdInt32Unmarshal(`, varName, `, `, buf, `); err != nil {`)
		} else if gogoproto.IsStdUInt32(field) {
			g.P(`if err := `, g.typesPkg.Use(), `.StdUInt32Unmarshal(`, varName, `, `, buf, `); err != nil {`)
		} else if gogoproto.IsStdBool(field) {
			g.P(`if err := `, g.typesPkg.Use(), `.StdBoolUnmarshal(`, varName, `, `, buf, `); err != nil {`)
		} else if gogoproto.IsStdString(field) {
			g.P(`if err := `, g.typesPkg.Use(), `.StdStringUnmarshal(`, varName, `, `, buf, `); err != nil {`)
		} else if gogoproto.IsStdBytes(field) {
			g.P(`if err := `, g.typesPkg.Use(), `.StdBytesUnmarshal(`, varName, `, `, buf, `); err != nil {`)
		} else {
			desc := g.ObjectNamed(field.GetTypeName())
			msgname := g.TypeName(desc)
			g.P(varName, ` = &`, msgname, `{}`)
			g.P(`if err := `, varName, `.Unmarshal(`, buf, `); err != nil {`)
		}
		g.In()
		g.P(`return err`)
		g.Out()
		g.P(`}`)
		g.P(`iNdEx = postmsgIndex`)
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		g.P(`var mapbyteLen uint64`)
		g.decodeVarint("mapbyteLen", "uint64")
		g.P(`intMapbyteLen := int(mapbyteLen)`)
		g.P(`if intMapbyteLen < 0 {`)
		g.In()
		g.P(`return ErrInvalidLength` + g.localName)
		g.Out()
		g.P(`}`)
		g.P(`postbytesIndex := iNdEx + intMapbyteLen`)
		g.P(`if postbytesIndex < 0 {`)
		g.In()
		g.P(`return ErrInvalidLength` + g.localName)
		g.Out()
		g.P(`}`)
		g.P(`if postbytesIndex > l {`)
		g.In()
		g.P(`return `, g.ioPkg.Use(), `.ErrUnexpectedEOF`)
		g.Out()
		g.P(`}`)
		if customType {
			g.P(`if err := `, varName, `.Unmarshal(dAtA[iNdEx:postbytesIndex]); err != nil {`)
			g.In()
			g.P(`return err`)
			g.Out()
			g.P(`}`)
		} else {
			g.P(varName, ` = make([]byte, mapbyteLen)`)
			g.P(`copy(`, varName, `, dAtA[iNdEx:postbytesIndex])`)
		}
		g.P(`iNdEx = postbytesIndex`)
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		g.decodeVarint(varName, "uint32")
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		typName := g.TypeName(g.ObjectNamed(field.GetTypeName()))
		g.decodeVarint(varName, typName)
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		g.decodeFixed32(varName, "int32")
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		g.decodeFixed64(varName, "int64")
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		g.P(`var `, varName, `temp int32`)
		g.decodeVarint(varName+"temp", "int32")
		g.P(varName, `temp = int32((uint32(`, varName, `temp) >> 1) ^ uint32(((`, varName, `temp&1)<<31)>>31))`)
		g.P(varName, ` = int32(`, varName, `temp)`)
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		g.P(`var `, varName, `temp uint64`)
		g.decodeVarint(varName+"temp", "uint64")
		g.P(varName, `temp = (`, varName, `temp >> 1) ^ uint64((int64(`, varName, `temp&1)<<63)>>63)`)
		g.P(varName, ` = int64(`, varName, `temp)`)
	}
}

func (g *gogo) noStarOrSliceType(msg *generator.Descriptor, field *descriptor.FieldDescriptorProto) string {
	typ, _ := g.GoType(msg, field)
	if typ[0] == '*' {
		return typ[1:]
	}
	if typ[0] == '[' && typ[1] == ']' {
		return typ[2:]
	}
	return typ
}

func (g *gogo) field(file *generator.FileDescriptor, msg *generator.Descriptor, field *descriptor.FieldDescriptorProto, fieldname string, proto3 bool) {
	repeated := field.IsRepeated()
	nullable := gogoproto.IsNullable(field)
	typ := g.noStarOrSliceType(msg, field)
	oneof := field.OneofIndex != nil
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		g.P(`var v uint64`)
		g.decodeFixed64("v", "uint64")
		if oneof {
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{`, typ, "(", g.mathPkg.Use(), `.Float64frombits(v))}`)
		} else if repeated {
			g.P(`v2 := `, typ, "(", g.mathPkg.Use(), `.Float64frombits(v))`)
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v2)`)
		} else if proto3 || !nullable {
			g.P(`m.`, fieldname, ` = `, typ, "(", g.mathPkg.Use(), `.Float64frombits(v))`)
		} else {
			g.P(`v2 := `, typ, "(", g.mathPkg.Use(), `.Float64frombits(v))`)
			g.P(`m.`, fieldname, ` = &v2`)
		}
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		g.P(`var v uint32`)
		g.decodeFixed32("v", "uint32")
		if oneof {
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{`, typ, "(", g.mathPkg.Use(), `.Float32frombits(v))}`)
		} else if repeated {
			g.P(`v2 := `, typ, "(", g.mathPkg.Use(), `.Float32frombits(v))`)
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v2)`)
		} else if proto3 || !nullable {
			g.P(`m.`, fieldname, ` = `, typ, "(", g.mathPkg.Use(), `.Float32frombits(v))`)
		} else {
			g.P(`v2 := `, typ, "(", g.mathPkg.Use(), `.Float32frombits(v))`)
			g.P(`m.`, fieldname, ` = &v2`)
		}
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		if oneof {
			g.P(`var v `, typ)
			g.decodeVarint("v", typ)
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{v}`)
		} else if repeated {
			g.P(`var v `, typ)
			g.decodeVarint("v", typ)
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
		} else if proto3 || !nullable {
			g.P(`m.`, fieldname, ` = 0`)
			g.decodeVarint("m."+fieldname, typ)
		} else {
			g.P(`var v `, typ)
			g.decodeVarint("v", typ)
			g.P(`m.`, fieldname, ` = &v`)
		}
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		if oneof {
			g.P(`var v `, typ)
			g.decodeVarint("v", typ)
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{v}`)
		} else if repeated {
			g.P(`var v `, typ)
			g.decodeVarint("v", typ)
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
		} else if proto3 || !nullable {
			g.P(`m.`, fieldname, ` = 0`)
			g.decodeVarint("m."+fieldname, typ)
		} else {
			g.P(`var v `, typ)
			g.decodeVarint("v", typ)
			g.P(`m.`, fieldname, ` = &v`)
		}
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		if oneof {
			g.P(`var v `, typ)
			g.decodeVarint("v", typ)
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{v}`)
		} else if repeated {
			g.P(`var v `, typ)
			g.decodeVarint("v", typ)
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
		} else if proto3 || !nullable {
			g.P(`m.`, fieldname, ` = 0`)
			g.decodeVarint("m."+fieldname, typ)
		} else {
			g.P(`var v `, typ)
			g.decodeVarint("v", typ)
			g.P(`m.`, fieldname, ` = &v`)
		}
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		if oneof {
			g.P(`var v `, typ)
			g.decodeFixed64("v", typ)
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{v}`)
		} else if repeated {
			g.P(`var v `, typ)
			g.decodeFixed64("v", typ)
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
		} else if proto3 || !nullable {
			g.P(`m.`, fieldname, ` = 0`)
			g.decodeFixed64("m."+fieldname, typ)
		} else {
			g.P(`var v `, typ)
			g.decodeFixed64("v", typ)
			g.P(`m.`, fieldname, ` = &v`)
		}
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		if oneof {
			g.P(`var v `, typ)
			g.decodeFixed32("v", typ)
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{v}`)
		} else if repeated {
			g.P(`var v `, typ)
			g.decodeFixed32("v", typ)
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
		} else if proto3 || !nullable {
			g.P(`m.`, fieldname, ` = 0`)
			g.decodeFixed32("m."+fieldname, typ)
		} else {
			g.P(`var v `, typ)
			g.decodeFixed32("v", typ)
			g.P(`m.`, fieldname, ` = &v`)
		}
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		g.P(`var v int`)
		g.decodeVarint("v", "int")
		if oneof {
			g.P(`b := `, typ, `(v != 0)`)
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{b}`)
		} else if repeated {
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, typ, `(v != 0))`)
		} else if proto3 || !nullable {
			g.P(`m.`, fieldname, ` = `, typ, `(v != 0)`)
		} else {
			g.P(`b := `, typ, `(v != 0)`)
			g.P(`m.`, fieldname, ` = &b`)
		}
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		g.P(`var stringLen uint64`)
		g.decodeVarint("stringLen", "uint64")
		g.P(`intStringLen := int(stringLen)`)
		g.P(`if intStringLen < 0 {`)
		g.In()
		g.P(`return ErrInvalidLength` + g.localName)
		g.Out()
		g.P(`}`)
		g.P(`postIndex := iNdEx + intStringLen`)
		g.P(`if postIndex < 0 {`)
		g.In()
		g.P(`return ErrInvalidLength` + g.localName)
		g.Out()
		g.P(`}`)
		g.P(`if postIndex > l {`)
		g.In()
		g.P(`return `, g.ioPkg.Use(), `.ErrUnexpectedEOF`)
		g.Out()
		g.P(`}`)
		if oneof {
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{`, typ, `(dAtA[iNdEx:postIndex])}`)
		} else if repeated {
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, typ, `(dAtA[iNdEx:postIndex]))`)
		} else if proto3 || !nullable {
			g.P(`m.`, fieldname, ` = `, typ, `(dAtA[iNdEx:postIndex])`)
		} else {
			g.P(`s := `, typ, `(dAtA[iNdEx:postIndex])`)
			g.P(`m.`, fieldname, ` = &s`)
		}
		g.P(`iNdEx = postIndex`)
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		panic(fmt.Errorf("unmarshaler does not support group %v", fieldname))
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		desc := g.ObjectNamed(field.GetTypeName())
		msgname := g.TypeName(desc)
		g.P(`var msglen int`)
		g.decodeVarint("msglen", "int")
		g.P(`if msglen < 0 {`)
		g.In()
		g.P(`return ErrInvalidLength` + g.localName)
		g.Out()
		g.P(`}`)
		g.P(`postIndex := iNdEx + msglen`)
		g.P(`if postIndex < 0 {`)
		g.In()
		g.P(`return ErrInvalidLength` + g.localName)
		g.Out()
		g.P(`}`)
		g.P(`if postIndex > l {`)
		g.In()
		g.P(`return `, g.ioPkg.Use(), `.ErrUnexpectedEOF`)
		g.Out()
		g.P(`}`)
		if oneof {
			buf := `dAtA[iNdEx:postIndex]`
			if gogoproto.IsStdTime(field) {
				if nullable {
					g.P(`v := new(time.Time)`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdTimeUnmarshal(v, `, buf, `); err != nil {`)
				} else {
					g.P(`v := time.Time{}`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdTimeUnmarshal(&v, `, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdDuration(field) {
				if nullable {
					g.P(`v := new(time.Duration)`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdDurationUnmarshal(v, `, buf, `); err != nil {`)
				} else {
					g.P(`v := time.Duration(0)`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdDurationUnmarshal(&v, `, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdDouble(field) {
				if nullable {
					g.P(`v := new(float64)`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdDoubleUnmarshal(v, `, buf, `); err != nil {`)
				} else {
					g.P(`v := 0`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdDoubleUnmarshal(&v, `, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdFloat(field) {
				if nullable {
					g.P(`v := new(float32)`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdFloatUnmarshal(v, `, buf, `); err != nil {`)
				} else {
					g.P(`v := 0`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdFloatUnmarshal(&v, `, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdInt64(field) {
				if nullable {
					g.P(`v := new(int64)`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdInt64Unmarshal(v, `, buf, `); err != nil {`)
				} else {
					g.P(`v := 0`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdInt64Unmarshal(&v, `, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdUInt64(field) {
				if nullable {
					g.P(`v := new(uint64)`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdUInt64Unmarshal(v, `, buf, `); err != nil {`)
				} else {
					g.P(`v := 0`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdUInt64Unmarshal(&v, `, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdInt32(field) {
				if nullable {
					g.P(`v := new(int32)`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdInt32Unmarshal(v, `, buf, `); err != nil {`)
				} else {
					g.P(`v := 0`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdInt32Unmarshal(&v, `, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdUInt32(field) {
				if nullable {
					g.P(`v := new(uint32)`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdUInt32Unmarshal(v, `, buf, `); err != nil {`)
				} else {
					g.P(`v := 0`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdUInt32Unmarshal(&v, `, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdBool(field) {
				if nullable {
					g.P(`v := new(bool)`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdBoolUnmarshal(v, `, buf, `); err != nil {`)
				} else {
					g.P(`v := false`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdBoolUnmarshal(&v, `, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdString(field) {
				if nullable {
					g.P(`v := new(string)`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdStringUnmarshal(v, `, buf, `); err != nil {`)
				} else {
					g.P(`v := ""`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdStringUnmarshal(&v, `, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdBytes(field) {
				if nullable {
					g.P(`v := new([]byte)`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdBytesUnmarshal(v, `, buf, `); err != nil {`)
				} else {
					g.P(`var v []byte`)
					g.P(`if err := `, g.typesPkg.Use(), `.StdBytesUnmarshal(&v, `, buf, `); err != nil {`)
				}
			} else {
				g.P(`v := &`, msgname, `{}`)
				g.P(`if err := v.Unmarshal(`, buf, `); err != nil {`)
			}
			g.In()
			g.P(`return err`)
			g.Out()
			g.P(`}`)
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{v}`)
		} else if g.IsMap(field) {
			m := g.GoMapType(nil, field)

			keygoTyp, _ := g.GoType(nil, m.KeyField)
			keygoAliasTyp, _ := g.GoType(nil, m.KeyAliasField)
			// keys may not be pointers
			keygoTyp = strings.Replace(keygoTyp, "*", "", 1)
			keygoAliasTyp = strings.Replace(keygoAliasTyp, "*", "", 1)

			valuegoTyp, _ := g.GoType(nil, m.ValueField)
			valuegoAliasTyp, _ := g.GoType(nil, m.ValueAliasField)

			// if the map type is an alias and key or values are aliases (type Foo map[Bar]Baz),
			// we need to explicitly record their use here.
			if gogoproto.IsCastKey(field) {
				g.RecordTypeUse(m.KeyAliasField.GetTypeName())
			}
			if gogoproto.IsCastValue(field) {
				g.RecordTypeUse(m.ValueAliasField.GetTypeName())
			}

			nullable, valuegoTyp, valuegoAliasTyp = generator.GoMapValueTypes(field, m.ValueField, valuegoTyp, valuegoAliasTyp)
			if gogoproto.IsStdType(field) {
				valuegoTyp = valuegoAliasTyp
			}

			g.P(`if m.`, fieldname, ` == nil {`)
			g.In()
			g.P(`m.`, fieldname, ` = make(`, m.GoType, `)`)
			g.Out()
			g.P(`}`)

			g.declareMapField("mapkey", false, false, m.KeyAliasField)
			g.declareMapField("mapvalue", nullable, gogoproto.IsCustomType(field), m.ValueAliasField)
			g.P(`for iNdEx < postIndex {`)
			g.In()

			g.P(`entryPreIndex := iNdEx`)
			g.P(`var wire uint64`)
			g.decodeVarint("wire", "uint64")
			g.P(`fieldNum := int32(wire >> 3)`)

			g.P(`if fieldNum == 1 {`)
			g.In()
			g.mapUnmarshalField("mapkey", false, m.KeyAliasField)
			g.Out()
			g.P(`} else if fieldNum == 2 {`)
			g.In()
			g.mapUnmarshalField("mapvalue", gogoproto.IsCustomType(field), m.ValueAliasField)
			g.Out()
			g.P(`} else {`)
			g.In()
			g.P(`iNdEx = entryPreIndex`)
			g.P(`skippy, err := skip`, g.localName, `(dAtA[iNdEx:])`)
			g.P(`if err != nil {`)
			g.In()
			g.P(`return err`)
			g.Out()
			g.P(`}`)
			g.P(`if (skippy < 0) || (iNdEx + skippy) < 0 {`)
			g.In()
			g.P(`return ErrInvalidLength`, g.localName)
			g.Out()
			g.P(`}`)
			g.P(`if (iNdEx + skippy) > postIndex {`)
			g.In()
			g.P(`return `, g.ioPkg.Use(), `.ErrUnexpectedEOF`)
			g.Out()
			g.P(`}`)
			g.P(`iNdEx += skippy`)
			g.Out()
			g.P(`}`)

			g.Out()
			g.P(`}`)

			s := `m.` + fieldname
			if keygoTyp == keygoAliasTyp {
				s += `[mapkey]`
			} else {
				s += `[` + keygoAliasTyp + `(mapkey)]`
			}

			v := `mapvalue`
			if (m.ValueField.IsMessage() || gogoproto.IsCustomType(field)) && !nullable {
				v = `*` + v
			}
			if valuegoTyp != valuegoAliasTyp {
				v = `((` + valuegoAliasTyp + `)(` + v + `))`
			}

			g.P(s, ` = `, v)
		} else if repeated {
			if gogoproto.IsStdTime(field) {
				if nullable {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, new(time.Time))`)
				} else {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, time.Time{})`)
				}
			} else if gogoproto.IsStdDuration(field) {
				if nullable {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, new(time.Duration))`)
				} else {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, time.Duration(0))`)
				}
			} else if gogoproto.IsStdDouble(field) {
				if nullable {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, new(float64))`)
				} else {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, 0)`)
				}
			} else if gogoproto.IsStdFloat(field) {
				if nullable {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, new(float32))`)
				} else {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, 0)`)
				}
			} else if gogoproto.IsStdInt64(field) {
				if nullable {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, new(int64))`)
				} else {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, 0)`)
				}
			} else if gogoproto.IsStdUInt64(field) {
				if nullable {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, new(uint64))`)
				} else {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, 0)`)
				}
			} else if gogoproto.IsStdInt32(field) {
				if nullable {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, new(int32))`)
				} else {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, 0)`)
				}
			} else if gogoproto.IsStdUInt32(field) {
				if nullable {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, new(uint32))`)
				} else {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, 0)`)
				}
			} else if gogoproto.IsStdBool(field) {
				if nullable {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, new(bool))`)
				} else {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, false)`)
				}
			} else if gogoproto.IsStdString(field) {
				if nullable {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, new(string))`)
				} else {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, "")`)
				}
			} else if gogoproto.IsStdBytes(field) {
				if nullable {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, new([]byte))`)
				} else {
					g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, []byte{})`)
				}
			} else if nullable && !gogoproto.IsCustomType(field) {
				g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, &`, msgname, `{})`)
			} else {
				goType, _ := g.GoType(nil, field)
				// remove the slice from the type, i.e. []*T -> *T
				goType = goType[2:]
				g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, goType, `{})`)
			}
			varName := `m.` + fieldname + `[len(m.` + fieldname + `)-1]`
			buf := `dAtA[iNdEx:postIndex]`
			if gogoproto.IsStdTime(field) {
				if nullable {
					g.P(`if err := `, g.typesPkg.Use(), `.StdTimeUnmarshal(`, varName, `,`, buf, `); err != nil {`)
				} else {
					g.P(`if err := `, g.typesPkg.Use(), `.StdTimeUnmarshal(&(`, varName, `),`, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdDuration(field) {
				if nullable {
					g.P(`if err := `, g.typesPkg.Use(), `.StdDurationUnmarshal(`, varName, `,`, buf, `); err != nil {`)
				} else {
					g.P(`if err := `, g.typesPkg.Use(), `.StdDurationUnmarshal(&(`, varName, `),`, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdDouble(field) {
				if nullable {
					g.P(`if err := `, g.typesPkg.Use(), `.StdDoubleUnmarshal(`, varName, `,`, buf, `); err != nil {`)
				} else {
					g.P(`if err := `, g.typesPkg.Use(), `.StdDoubleUnmarshal(&(`, varName, `),`, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdFloat(field) {
				if nullable {
					g.P(`if err := `, g.typesPkg.Use(), `.StdFloatUnmarshal(`, varName, `,`, buf, `); err != nil {`)
				} else {
					g.P(`if err := `, g.typesPkg.Use(), `.StdFloatUnmarshal(&(`, varName, `),`, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdInt64(field) {
				if nullable {
					g.P(`if err := `, g.typesPkg.Use(), `.StdInt64Unmarshal(`, varName, `,`, buf, `); err != nil {`)
				} else {
					g.P(`if err := `, g.typesPkg.Use(), `.StdInt64Unmarshal(&(`, varName, `),`, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdUInt64(field) {
				if nullable {
					g.P(`if err := `, g.typesPkg.Use(), `.StdUInt64Unmarshal(`, varName, `,`, buf, `); err != nil {`)
				} else {
					g.P(`if err := `, g.typesPkg.Use(), `.StdUInt64Unmarshal(&(`, varName, `),`, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdInt32(field) {
				if nullable {
					g.P(`if err := `, g.typesPkg.Use(), `.StdInt32Unmarshal(`, varName, `,`, buf, `); err != nil {`)
				} else {
					g.P(`if err := `, g.typesPkg.Use(), `.StdInt32Unmarshal(&(`, varName, `),`, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdUInt32(field) {
				if nullable {
					g.P(`if err := `, g.typesPkg.Use(), `.StdUInt32Unmarshal(`, varName, `,`, buf, `); err != nil {`)
				} else {
					g.P(`if err := `, g.typesPkg.Use(), `.StdUInt32Unmarshal(&(`, varName, `),`, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdBool(field) {
				if nullable {
					g.P(`if err := `, g.typesPkg.Use(), `.StdBoolUnmarshal(`, varName, `,`, buf, `); err != nil {`)
				} else {
					g.P(`if err := `, g.typesPkg.Use(), `.StdBoolUnmarshal(&(`, varName, `),`, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdString(field) {
				if nullable {
					g.P(`if err := `, g.typesPkg.Use(), `.StdStringUnmarshal(`, varName, `,`, buf, `); err != nil {`)
				} else {
					g.P(`if err := `, g.typesPkg.Use(), `.StdStringUnmarshal(&(`, varName, `),`, buf, `); err != nil {`)
				}
			} else if gogoproto.IsStdBytes(field) {
				if nullable {
					g.P(`if err := `, g.typesPkg.Use(), `.StdBytesUnmarshal(`, varName, `,`, buf, `); err != nil {`)
				} else {
					g.P(`if err := `, g.typesPkg.Use(), `.StdBytesUnmarshal(&(`, varName, `),`, buf, `); err != nil {`)
				}
			} else {
				g.P(`if err := `, varName, `.Unmarshal(`, buf, `); err != nil {`)
			}
			g.In()
			g.P(`return err`)
			g.Out()
			g.P(`}`)
		} else if nullable {
			g.P(`if m.`, fieldname, ` == nil {`)
			g.In()
			if gogoproto.IsStdTime(field) {
				g.P(`m.`, fieldname, ` = new(time.Time)`)
			} else if gogoproto.IsStdDuration(field) {
				g.P(`m.`, fieldname, ` = new(time.Duration)`)
			} else if gogoproto.IsStdDouble(field) {
				g.P(`m.`, fieldname, ` = new(float64)`)
			} else if gogoproto.IsStdFloat(field) {
				g.P(`m.`, fieldname, ` = new(float32)`)
			} else if gogoproto.IsStdInt64(field) {
				g.P(`m.`, fieldname, ` = new(int64)`)
			} else if gogoproto.IsStdUInt64(field) {
				g.P(`m.`, fieldname, ` = new(uint64)`)
			} else if gogoproto.IsStdInt32(field) {
				g.P(`m.`, fieldname, ` = new(int32)`)
			} else if gogoproto.IsStdUInt32(field) {
				g.P(`m.`, fieldname, ` = new(uint32)`)
			} else if gogoproto.IsStdBool(field) {
				g.P(`m.`, fieldname, ` = new(bool)`)
			} else if gogoproto.IsStdString(field) {
				g.P(`m.`, fieldname, ` = new(string)`)
			} else if gogoproto.IsStdBytes(field) {
				g.P(`m.`, fieldname, ` = new([]byte)`)
			} else {
				goType, _ := g.GoType(nil, field)
				// remove the star from the type
				g.P(`m.`, fieldname, ` = &`, goType[1:], `{}`)
			}
			g.Out()
			g.P(`}`)
			if gogoproto.IsStdTime(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdTimeUnmarshal(m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdDuration(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdDurationUnmarshal(m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdDouble(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdDoubleUnmarshal(m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdFloat(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdFloatUnmarshal(m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdInt64(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdInt64Unmarshal(m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdUInt64(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdUInt64Unmarshal(m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdInt32(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdInt32Unmarshal(m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdUInt32(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdUInt32Unmarshal(m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdBool(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdBoolUnmarshal(m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdString(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdStringUnmarshal(m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdBytes(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdBytesUnmarshal(m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else {
				g.P(`if err := m.`, fieldname, `.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {`)
			}
			g.In()
			g.P(`return err`)
			g.Out()
			g.P(`}`)
		} else {
			if gogoproto.IsStdTime(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdTimeUnmarshal(&m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdDuration(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdDurationUnmarshal(&m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdDouble(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdDoubleUnmarshal(&m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdFloat(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdFloatUnmarshal(&m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdInt64(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdInt64Unmarshal(&m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdUInt64(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdUInt64Unmarshal(&m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdInt32(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdInt32Unmarshal(&m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdUInt32(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdUInt32Unmarshal(&m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdBool(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdBoolUnmarshal(&m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdString(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdStringUnmarshal(&m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else if gogoproto.IsStdBytes(field) {
				g.P(`if err := `, g.typesPkg.Use(), `.StdBytesUnmarshal(&m.`, fieldname, `, dAtA[iNdEx:postIndex]); err != nil {`)
			} else {
				g.P(`if err := m.`, fieldname, `.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {`)
			}
			g.In()
			g.P(`return err`)
			g.Out()
			g.P(`}`)
		}
		g.P(`iNdEx = postIndex`)

	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		g.P(`var byteLen int`)
		g.decodeVarint("byteLen", "int")
		g.P(`if byteLen < 0 {`)
		g.In()
		g.P(`return ErrInvalidLength` + g.localName)
		g.Out()
		g.P(`}`)
		g.P(`postIndex := iNdEx + byteLen`)
		g.P(`if postIndex < 0 {`)
		g.In()
		g.P(`return ErrInvalidLength` + g.localName)
		g.Out()
		g.P(`}`)
		g.P(`if postIndex > l {`)
		g.In()
		g.P(`return `, g.ioPkg.Use(), `.ErrUnexpectedEOF`)
		g.Out()
		g.P(`}`)
		if !gogoproto.IsCustomType(field) {
			if oneof {
				g.P(`v := make([]byte, postIndex-iNdEx)`)
				g.P(`copy(v, dAtA[iNdEx:postIndex])`)
				g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{v}`)
			} else if repeated {
				g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, make([]byte, postIndex-iNdEx))`)
				g.P(`copy(m.`, fieldname, `[len(m.`, fieldname, `)-1], dAtA[iNdEx:postIndex])`)
			} else {
				g.P(`m.`, fieldname, ` = append(m.`, fieldname, `[:0] , dAtA[iNdEx:postIndex]...)`)
				g.P(`if m.`, fieldname, ` == nil {`)
				g.In()
				g.P(`m.`, fieldname, ` = []byte{}`)
				g.Out()
				g.P(`}`)
			}
		} else {
			_, ctyp, err := generator.GetCustomType(field)
			if err != nil {
				panic(err)
			}
			if oneof {
				g.P(`var vv `, ctyp)
				g.P(`v := &vv`)
				g.P(`if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {`)
				g.In()
				g.P(`return err`)
				g.Out()
				g.P(`}`)
				g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{*v}`)
			} else if repeated {
				g.P(`var v `, ctyp)
				g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
				g.P(`if err := m.`, fieldname, `[len(m.`, fieldname, `)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {`)
				g.In()
				g.P(`return err`)
				g.Out()
				g.P(`}`)
			} else if nullable {
				g.P(`var v `, ctyp)
				g.P(`m.`, fieldname, ` = &v`)
				g.P(`if err := m.`, fieldname, `.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {`)
				g.In()
				g.P(`return err`)
				g.Out()
				g.P(`}`)
			} else {
				g.P(`if err := m.`, fieldname, `.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {`)
				g.In()
				g.P(`return err`)
				g.Out()
				g.P(`}`)
			}
		}
		g.P(`iNdEx = postIndex`)
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		if oneof {
			g.P(`var v `, typ)
			g.decodeVarint("v", typ)
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{v}`)
		} else if repeated {
			g.P(`var v `, typ)
			g.decodeVarint("v", typ)
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
		} else if proto3 || !nullable {
			g.P(`m.`, fieldname, ` = 0`)
			g.decodeVarint("m."+fieldname, typ)
		} else {
			g.P(`var v `, typ)
			g.decodeVarint("v", typ)
			g.P(`m.`, fieldname, ` = &v`)
		}
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		typName := g.TypeName(g.ObjectNamed(field.GetTypeName()))
		if oneof {
			g.P(`var v `, typName)
			g.decodeVarint("v", typName)
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{v}`)
		} else if repeated {
			g.P(`var v `, typName)
			g.decodeVarint("v", typName)
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
		} else if proto3 || !nullable {
			g.P(`m.`, fieldname, ` = 0`)
			g.decodeVarint("m."+fieldname, typName)
		} else {
			g.P(`var v `, typName)
			g.decodeVarint("v", typName)
			g.P(`m.`, fieldname, ` = &v`)
		}
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		if oneof {
			g.P(`var v `, typ)
			g.decodeFixed32("v", typ)
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{v}`)
		} else if repeated {
			g.P(`var v `, typ)
			g.decodeFixed32("v", typ)
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
		} else if proto3 || !nullable {
			g.P(`m.`, fieldname, ` = 0`)
			g.decodeFixed32("m."+fieldname, typ)
		} else {
			g.P(`var v `, typ)
			g.decodeFixed32("v", typ)
			g.P(`m.`, fieldname, ` = &v`)
		}
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		if oneof {
			g.P(`var v `, typ)
			g.decodeFixed64("v", typ)
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{v}`)
		} else if repeated {
			g.P(`var v `, typ)
			g.decodeFixed64("v", typ)
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
		} else if proto3 || !nullable {
			g.P(`m.`, fieldname, ` = 0`)
			g.decodeFixed64("m."+fieldname, typ)
		} else {
			g.P(`var v `, typ)
			g.decodeFixed64("v", typ)
			g.P(`m.`, fieldname, ` = &v`)
		}
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		g.P(`var v `, typ)
		g.decodeVarint("v", typ)
		g.P(`v = `, typ, `((uint32(v) >> 1) ^ uint32(((v&1)<<31)>>31))`)
		if oneof {
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{v}`)
		} else if repeated {
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
		} else if proto3 || !nullable {
			g.P(`m.`, fieldname, ` = v`)
		} else {
			g.P(`m.`, fieldname, ` = &v`)
		}
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		g.P(`var v uint64`)
		g.decodeVarint("v", "uint64")
		g.P(`v = (v >> 1) ^ uint64((int64(v&1)<<63)>>63)`)
		if oneof {
			g.P(`m.`, fieldname, ` = &`, g.OneOfTypeName(msg, field), `{`, typ, `(v)}`)
		} else if repeated {
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, typ, `(v))`)
		} else if proto3 || !nullable {
			g.P(`m.`, fieldname, ` = `, typ, `(v)`)
		} else {
			g.P(`v2 := `, typ, `(v)`)
			g.P(`m.`, fieldname, ` = &v2`)
		}
	default:
		panic("not implemented")
	}
}

func (g *gogo) GenerateUnmarshal(file *generator.FileDescriptor) {
	proto3 := gogoproto.IsProto3(file.FileDescriptorProto)
	g.atleastOne = false
	g.localName = generator.FileName(file)

	g.ioPkg = g.NewImport("io")
	g.mathPkg = g.NewImport("math")
	g.typesPkg = g.NewImport("github.com/gogo/protobuf/types")
	g.binaryPkg = g.NewImport("encoding/binary")
	fmtPkg := g.NewImport("fmt")
	protoPkg := g.NewImport("github.com/gogo/protobuf/proto")
	if !gogoproto.ImportsGoGoProto(file.FileDescriptorProto) {
		protoPkg = g.NewImport("github.com/golang/protobuf/proto")
	}

	for _, message := range file.Messages() {
		ccTypeName := generator.CamelCaseSlice(message.Proto.TypeName())
		if !gogoproto.IsUnmarshaler(file.FileDescriptorProto, message.Proto.DescriptorProto) &&
			!gogoproto.IsUnsafeUnmarshaler(file.FileDescriptorProto, message.Proto.DescriptorProto) {
			continue
		}
		if message.Proto.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}
		g.atleastOne = true

		// build a map required field_id -> bitmask offset
		rfMap := make(map[int32]uint)
		rfNextId := uint(0)
		for _, field := range message.Proto.Field {
			if field.IsRequired() {
				rfMap[field.GetNumber()] = rfNextId
				rfNextId++
			}
		}
		rfCount := len(rfMap)

		g.P(`func (m *`, ccTypeName, `) Unmarshal(dAtA []byte) error {`)
		g.In()
		if rfCount > 0 {
			g.P(`var hasFields [`, strconv.Itoa(1+(rfCount-1)/64), `]uint64`)
		}
		g.P(`l := len(dAtA)`)
		g.P(`iNdEx := 0`)
		g.P(`for iNdEx < l {`)
		g.In()
		g.P(`preIndex := iNdEx`)
		g.P(`var wire uint64`)
		g.decodeVarint("wire", "uint64")
		g.P(`fieldNum := int32(wire >> 3)`)
		if len(message.Proto.Field) > 0 || !message.Proto.IsGroup() {
			g.P(`wireType := int(wire & 0x7)`)
		}
		if !message.Proto.IsGroup() {
			g.P(`if wireType == `, strconv.Itoa(proto.WireEndGroup), ` {`)
			g.In()
			g.P(`return `, fmtPkg.Use(), `.Errorf("proto: `+message.Proto.GetName()+`: wiretype end group for non-group")`)
			g.Out()
			g.P(`}`)
		}
		g.P(`if fieldNum <= 0 {`)
		g.In()
		g.P(`return `, fmtPkg.Use(), `.Errorf("proto: `+message.Proto.GetName()+`: illegal tag %d (wire type %d)", fieldNum, wire)`)
		g.Out()
		g.P(`}`)
		g.P(`switch fieldNum {`)
		g.In()
		for _, field := range message.Proto.Field {
			fieldname := g.GetFieldName(message.Proto, field)
			errFieldname := fieldname
			if field.OneofIndex != nil {
				errFieldname = g.GetOneOfFieldName(message.Proto, field)
			}
			possiblyPacked := field.IsScalar() && field.IsRepeated()
			g.P(`case `, strconv.Itoa(int(field.GetNumber())), `:`)
			g.In()
			wireType := field.WireType()
			if possiblyPacked {
				g.P(`if wireType == `, strconv.Itoa(wireType), `{`)
				g.In()
				g.field(file, message.Proto, field, fieldname, false)
				g.Out()
				g.P(`} else if wireType == `, strconv.Itoa(proto.WireBytes), `{`)
				g.In()
				g.P(`var packedLen int`)
				g.decodeVarint("packedLen", "int")
				g.P(`if packedLen < 0 {`)
				g.In()
				g.P(`return ErrInvalidLength` + g.localName)
				g.Out()
				g.P(`}`)
				g.P(`postIndex := iNdEx + packedLen`)
				g.P(`if postIndex < 0 {`)
				g.In()
				g.P(`return ErrInvalidLength` + g.localName)
				g.Out()
				g.P(`}`)
				g.P(`if postIndex > l {`)
				g.In()
				g.P(`return `, g.ioPkg.Use(), `.ErrUnexpectedEOF`)
				g.Out()
				g.P(`}`)

				g.P(`var elementCount int`)
				switch *field.Type {
				case descriptor.FieldDescriptorProto_TYPE_DOUBLE, descriptor.FieldDescriptorProto_TYPE_FIXED64, descriptor.FieldDescriptorProto_TYPE_SFIXED64:
					g.P(`elementCount = packedLen/`, 8)
				case descriptor.FieldDescriptorProto_TYPE_FLOAT, descriptor.FieldDescriptorProto_TYPE_FIXED32, descriptor.FieldDescriptorProto_TYPE_SFIXED32:
					g.P(`elementCount = packedLen/`, 4)
				case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_UINT32, descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SINT64:
					g.P(`var count int`)
					g.P(`for _, integer := range dAtA[iNdEx:postIndex] {`)
					g.In()
					g.P(`if integer < 128 {`)
					g.In()
					g.P(`count++`)
					g.Out()
					g.P(`}`)
					g.Out()
					g.P(`}`)
					g.P(`elementCount = count`)
				case descriptor.FieldDescriptorProto_TYPE_BOOL:
					g.P(`elementCount = packedLen`)
				}
				g.P(`if elementCount != 0 && len(m.`, fieldname, `) == 0 {`)
				g.In()
				g.P(`m.`, fieldname, ` = make([]`, g.noStarOrSliceType(message.Proto, field), `, 0, elementCount)`)
				g.Out()
				g.P(`}`)

				g.P(`for iNdEx < postIndex {`)
				g.In()
				g.field(file, message.Proto, field, fieldname, false)
				g.Out()
				g.P(`}`)
				g.Out()
				g.P(`} else {`)
				g.In()
				g.P(`return ` + fmtPkg.Use() + `.Errorf("proto: wrong wireType = %d for field ` + errFieldname + `", wireType)`)
				g.Out()
				g.P(`}`)
			} else {
				g.P(`if wireType != `, strconv.Itoa(wireType), `{`)
				g.In()
				g.P(`return ` + fmtPkg.Use() + `.Errorf("proto: wrong wireType = %d for field ` + errFieldname + `", wireType)`)
				g.Out()
				g.P(`}`)
				g.field(file, message.Proto, field, fieldname, proto3)
			}

			if field.IsRequired() {
				fieldBit, ok := rfMap[field.GetNumber()]
				if !ok {
					panic("field is required, but no bit registered")
				}
				g.P(`hasFields[`, strconv.Itoa(int(fieldBit/64)), `] |= uint64(`, fmt.Sprintf("0x%08x", uint64(1)<<(fieldBit%64)), `)`)
			}
		}
		g.Out()
		g.P(`default:`)
		g.In()
		if message.Proto.DescriptorProto.HasExtension() {
			c := []string{}
			for _, erange := range message.Proto.GetExtensionRange() {
				c = append(c, `((fieldNum >= `+strconv.Itoa(int(erange.GetStart()))+") && (fieldNum<"+strconv.Itoa(int(erange.GetEnd()))+`))`)
			}
			g.P(`if `, strings.Join(c, "||"), `{`)
			g.In()
			g.P(`var sizeOfWire int`)
			g.P(`for {`)
			g.In()
			g.P(`sizeOfWire++`)
			g.P(`wire >>= 7`)
			g.P(`if wire == 0 {`)
			g.In()
			g.P(`break`)
			g.Out()
			g.P(`}`)
			g.Out()
			g.P(`}`)
			g.P(`iNdEx-=sizeOfWire`)
			g.P(`skippy, err := skip`, g.localName+`(dAtA[iNdEx:])`)
			g.P(`if err != nil {`)
			g.In()
			g.P(`return err`)
			g.Out()
			g.P(`}`)
			g.P(`if (skippy < 0) || (iNdEx + skippy) < 0 {`)
			g.In()
			g.P(`return ErrInvalidLength`, g.localName)
			g.Out()
			g.P(`}`)
			g.P(`if (iNdEx + skippy) > l {`)
			g.In()
			g.P(`return `, g.ioPkg.Use(), `.ErrUnexpectedEOF`)
			g.Out()
			g.P(`}`)
			g.P(protoPkg.Use(), `.AppendExtension(m, int32(fieldNum), dAtA[iNdEx:iNdEx+skippy])`)
			g.P(`iNdEx += skippy`)
			g.Out()
			g.P(`} else {`)
			g.In()
		}
		g.P(`iNdEx=preIndex`)
		g.P(`skippy, err := skip`, g.localName, `(dAtA[iNdEx:])`)
		g.P(`if err != nil {`)
		g.In()
		g.P(`return err`)
		g.Out()
		g.P(`}`)
		g.P(`if (skippy < 0) || (iNdEx + skippy) < 0 {`)
		g.In()
		g.P(`return ErrInvalidLength`, g.localName)
		g.Out()
		g.P(`}`)
		g.P(`if (iNdEx + skippy) > l {`)
		g.In()
		g.P(`return `, g.ioPkg.Use(), `.ErrUnexpectedEOF`)
		g.Out()
		g.P(`}`)
		if gogoproto.HasUnrecognized(file.FileDescriptorProto, message.Proto.DescriptorProto) {
			g.P(`m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)`)
		}
		g.P(`iNdEx += skippy`)
		g.Out()
		if message.Proto.DescriptorProto.HasExtension() {
			g.Out()
			g.P(`}`)
		}
		g.Out()
		g.P(`}`)
		g.Out()
		g.P(`}`)

		for _, field := range message.Proto.Field {
			if !field.IsRequired() {
				continue
			}

			fieldBit, ok := rfMap[field.GetNumber()]
			if !ok {
				panic("field is required, but no bit registered")
			}

			g.P(`if hasFields[`, strconv.Itoa(int(fieldBit/64)), `] & uint64(`, fmt.Sprintf("0x%08x", uint64(1)<<(fieldBit%64)), `) == 0 {`)
			g.In()
			if !gogoproto.ImportsGoGoProto(file.FileDescriptorProto) {
				g.P(`return new(`, protoPkg.Use(), `.RequiredNotSetError)`)
			} else {
				g.P(`return `, protoPkg.Use(), `.NewRequiredNotSetError("`, field.GetName(), `")`)
			}
			g.Out()
			g.P(`}`)
		}
		g.P()
		g.P(`if iNdEx > l {`)
		g.In()
		g.P(`return ` + g.ioPkg.Use() + `.ErrUnexpectedEOF`)
		g.Out()
		g.P(`}`)
		g.P(`return nil`)
		g.Out()
		g.P(`}`)
	}
	if !g.atleastOne {
		return
	}

	g.P(`func skip` + g.localName + `(dAtA []byte) (n int, err error) {
		l := len(dAtA)
		iNdEx := 0
		depth := 0
		for iNdEx < l {
			var wire uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflow` + g.localName + `
				}
				if iNdEx >= l {
					return 0, ` + g.ioPkg.Use() + `.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				wire |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			wireType := int(wire & 0x7)
			switch wireType {
			case 0:
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflow` + g.localName + `
					}
					if iNdEx >= l {
						return 0, ` + g.ioPkg.Use() + `.ErrUnexpectedEOF
					}
					iNdEx++
					if dAtA[iNdEx-1] < 0x80 {
						break
					}
				}
			case 1:
				iNdEx += 8
			case 2:
				var length int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflow` + g.localName + `
					}
					if iNdEx >= l {
						return 0, ` + g.ioPkg.Use() + `.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					length |= (int(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if length < 0 {
					return 0, ErrInvalidLength` + g.localName + `
				}
				iNdEx += length
			case 3:
				depth++
			case 4:
				if depth == 0 {
					return 0, ErrUnexpectedEndOfGroup` + g.localName + `
				}
				depth--
			case 5:
				iNdEx += 4
			default:
				return 0, ` + fmtPkg.Use() + `.Errorf("proto: illegal wireType %d", wireType)
			}
			if iNdEx < 0 {
				return 0, ErrInvalidLength` + g.localName + `
			}
			if depth == 0 {
				return iNdEx, nil
			}
		}
		return 0, ` + g.ioPkg.Use() + `.ErrUnexpectedEOF
	}

	var (
		ErrInvalidLength` + g.localName + ` = ` + fmtPkg.Use() + `.Errorf("proto: negative length found during unmarshaling")
		ErrIntOverflow` + g.localName + ` = ` + fmtPkg.Use() + `.Errorf("proto: integer overflow")
		ErrUnexpectedEndOfGroup` + g.localName + ` = ` + fmtPkg.Use() + `.Errorf("proto: unexpected end of group")
	)
	`)
}
