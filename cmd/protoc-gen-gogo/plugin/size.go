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
The size plugin generates a Size or ProtoSize method for each message.
This is useful with the MarshalTo method generated by the marshalto plugin and the
gogoproto.marshaler and gogoproto.marshaler_all extensions.

It is enabled by the following extensions:

  - sizer
  - sizer_all
  - protosizer
  - protosizer_all

The size plugin also generates a test given it is enabled using one of the following extensions:

  - testgen
  - testgen_all

And a benchmark given it is enabled using one of the following extensions:

  - benchgen
  - benchgen_all

Let us look at:

  github.com/gogo/protobuf/test/example/example.proto

Btw all the output can be seen at:

  github.com/gogo/protobuf/test/example/*

The following message:

  option (gogoproto.sizer_all) = true;

  message B {
	option (gogoproto.description) = true;
	optional A A = 1 [(gogoproto.nullable) = false, (gogoproto.embed) = true];
	repeated bytes G = 2 [(gogoproto.customtype) = "github.com/gogo/protobuf/test/custom.Uint128", (gogoproto.nullable) = false];
  }

given to the size plugin, will generate the following code:

  func (m *B) XSize() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.A.XSize()
	n += 1 + l + sovExample(uint64(l))
	if len(m.G) > 0 {
		for _, e := range m.G {
			l = e.XSize()
			n += 1 + l + sovExample(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
  }

and the following test code:

	func TestBSize(t *testing5.T) {
		popr := math_rand5.New(math_rand5.NewSource(time5.Now().UnixNano()))
		p := NewPopulatedB(popr, true)
		dAtA, err := github_com_gogo_protobuf_proto2.Marshal(p)
		if err != nil {
			panic(err)
		}
		size := g.XSize()
		if len(dAtA) != size {
			t.Fatalf("size %v != marshalled size %v", size, len(dAtA))
		}
	}

	func BenchmarkBSize(b *testing5.B) {
		popr := math_rand5.New(math_rand5.NewSource(616))
		total := 0
		pops := make([]*B, 1000)
		for i := 0; i < 1000; i++ {
			pops[i] = NewPopulatedB(popr, false)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			total += pops[i%1000].Size()
		}
		b.SetBytes(int64(total / b.N))
	}

The sovExample function is a size of varint function for the example.pb.go file.

*/
package gogo

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	vanity2 "github.com/vine-io/vine/cmd/generator/vanity"

	"github.com/vine-io/vine/cmd/generator"
)

func wireToType(wire string) int {
	switch wire {
	case "fixed64":
		return proto.WireFixed64
	case "fixed32":
		return proto.WireFixed32
	case "varint":
		return proto.WireVarint
	case "bytes":
		return proto.WireBytes
	case "group":
		return proto.WireBytes
	case "zigzag32":
		return proto.WireVarint
	case "zigzag64":
		return proto.WireVarint
	}
	panic("unreachable")
}

func (g *gogo) sizeVarint() {
	g.P(`
	func sov`, g.localName, `(x uint64) (n int) {
                return (`, g.bitsPkg.Use(), `.Len64(x | 1) + 6)/ 7
	}`)
}

func (g *gogo) sizeZigZag() {
	g.P(`func soz`, g.localName, `(x uint64) (n int) {
		return sov`, g.localName, `(uint64((x << 1) ^ uint64((int64(x) >> 63))))
	}`)
}

func (g *gogo) std(field *descriptor.FieldDescriptorProto, name string) (string, bool) {
	ptr := ""
	if gogoproto.IsNullable(field) {
		ptr = "*"
	}
	if gogoproto.IsStdTime(field) {
		return g.typesPkg.Use() + `.SizeOfStdTime(` + ptr + name + `)`, true
	} else if gogoproto.IsStdDuration(field) {
		return g.typesPkg.Use() + `.SizeOfStdDuration(` + ptr + name + `)`, true
	} else if gogoproto.IsStdDouble(field) {
		return g.typesPkg.Use() + `.SizeOfStdDouble(` + ptr + name + `)`, true
	} else if gogoproto.IsStdFloat(field) {
		return g.typesPkg.Use() + `.SizeOfStdFloat(` + ptr + name + `)`, true
	} else if gogoproto.IsStdInt64(field) {
		return g.typesPkg.Use() + `.SizeOfStdInt64(` + ptr + name + `)`, true
	} else if gogoproto.IsStdUInt64(field) {
		return g.typesPkg.Use() + `.SizeOfStdUInt64(` + ptr + name + `)`, true
	} else if gogoproto.IsStdInt32(field) {
		return g.typesPkg.Use() + `.SizeOfStdInt32(` + ptr + name + `)`, true
	} else if gogoproto.IsStdUInt32(field) {
		return g.typesPkg.Use() + `.SizeOfStdUInt32(` + ptr + name + `)`, true
	} else if gogoproto.IsStdBool(field) {
		return g.typesPkg.Use() + `.SizeOfStdBool(` + ptr + name + `)`, true
	} else if gogoproto.IsStdString(field) {
		return g.typesPkg.Use() + `.SizeOfStdString(` + ptr + name + `)`, true
	} else if gogoproto.IsStdBytes(field) {
		return g.typesPkg.Use() + `.SizeOfStdBytes(` + ptr + name + `)`, true
	}
	return "", false
}

func (g *gogo) generateField(proto3 bool, file *generator.FileDescriptor, msg *generator.MessageDescriptor, f *generator.FieldDescriptor, sizeName string) {
	message := msg.Proto
	field := f.Proto
	fieldname := g.GetOneOfFieldName(msg, f)
	nullable := gogoproto.IsNullable(field)
	repeated := field.IsRepeated()
	_, isInline := g.extractTags(f.Comments)[_inline]
	doNilCheck := gogoproto.NeedsNilCheck(proto3, field)
	if repeated {
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
	_, wire := g.GoType(message, field)
	wireType := wireToType(wire)
	fieldNumber := field.GetNumber()
	if packed {
		wireType = proto.WireBytes
	}
	key := keySize(fieldNumber, wireType)
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE,
		descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		if packed {
			g.P(`n+=`, strconv.Itoa(key), `+sov`, g.localName, `(uint64(len(m.`, fieldname, `)*8))`, `+len(m.`, fieldname, `)*8`)
		} else if repeated {
			g.P(`n+=`, strconv.Itoa(key+8), `*len(m.`, fieldname, `)`)
		} else if proto3 {
			g.P(`if m.`, fieldname, ` != 0 {`)
			g.In()
			g.P(`n+=`, strconv.Itoa(key+8))
			g.Out()
			g.P(`}`)
		} else if nullable {
			g.P(`n+=`, strconv.Itoa(key+8))
		} else {
			g.P(`n+=`, strconv.Itoa(key+8))
		}
	case descriptor.FieldDescriptorProto_TYPE_FLOAT,
		descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		if packed {
			g.P(`n+=`, strconv.Itoa(key), `+sov`, g.localName, `(uint64(len(m.`, fieldname, `)*4))`, `+len(m.`, fieldname, `)*4`)
		} else if repeated {
			g.P(`n+=`, strconv.Itoa(key+4), `*len(m.`, fieldname, `)`)
		} else if proto3 {
			g.P(`if m.`, fieldname, ` != 0 {`)
			g.In()
			g.P(`n+=`, strconv.Itoa(key+4))
			g.Out()
			g.P(`}`)
		} else if nullable {
			g.P(`n+=`, strconv.Itoa(key+4))
		} else {
			g.P(`n+=`, strconv.Itoa(key+4))
		}
	case descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_ENUM,
		descriptor.FieldDescriptorProto_TYPE_INT32:
		if packed {
			g.P(`l = 0`)
			g.P(`for _, e := range m.`, fieldname, ` {`)
			g.In()
			g.P(`l+=sov`, g.localName, `(uint64(e))`)
			g.Out()
			g.P(`}`)
			g.P(`n+=`, strconv.Itoa(key), `+sov`, g.localName, `(uint64(l))+l`)
		} else if repeated {
			g.P(`for _, e := range m.`, fieldname, ` {`)
			g.In()
			g.P(`n+=`, strconv.Itoa(key), `+sov`, g.localName, `(uint64(e))`)
			g.Out()
			g.P(`}`)
		} else if proto3 {
			g.P(`if m.`, fieldname, ` != 0 {`)
			g.In()
			g.P(`n+=`, strconv.Itoa(key), `+sov`, g.localName, `(uint64(m.`, fieldname, `))`)
			g.Out()
			g.P(`}`)
		} else if nullable {
			g.P(`n+=`, strconv.Itoa(key), `+sov`, g.localName, `(uint64(*m.`, fieldname, `))`)
		} else {
			g.P(`n+=`, strconv.Itoa(key), `+sov`, g.localName, `(uint64(m.`, fieldname, `))`)
		}
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		if packed {
			g.P(`n+=`, strconv.Itoa(key), `+sov`, g.localName, `(uint64(len(m.`, fieldname, `)))`, `+len(m.`, fieldname, `)*1`)
		} else if repeated {
			g.P(`n+=`, strconv.Itoa(key+1), `*len(m.`, fieldname, `)`)
		} else if proto3 {
			g.P(`if m.`, fieldname, ` {`)
			g.In()
			g.P(`n+=`, strconv.Itoa(key+1))
			g.Out()
			g.P(`}`)
		} else if nullable {
			g.P(`n+=`, strconv.Itoa(key+1))
		} else {
			g.P(`n+=`, strconv.Itoa(key+1))
		}
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		if repeated {
			g.P(`for _, s := range m.`, fieldname, ` { `)
			g.In()
			g.P(`l = len(s)`)
			g.P(`n+=`, strconv.Itoa(key), `+l+sov`, g.localName, `(uint64(l))`)
			g.Out()
			g.P(`}`)
		} else if proto3 {
			g.P(`l=len(m.`, fieldname, `)`)
			g.P(`if l > 0 {`)
			g.In()
			g.P(`n+=`, strconv.Itoa(key), `+l+sov`, g.localName, `(uint64(l))`)
			g.Out()
			g.P(`}`)
		} else if nullable {
			g.P(`l=len(*m.`, fieldname, `)`)
			g.P(`n+=`, strconv.Itoa(key), `+l+sov`, g.localName, `(uint64(l))`)
		} else {
			g.P(`l=len(m.`, fieldname, `)`)
			g.P(`n+=`, strconv.Itoa(key), `+l+sov`, g.localName, `(uint64(l))`)
		}
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		panic(fmt.Errorf("size does not support group %v", fieldname))
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		if g.IsMap(field) {
			m := g.GoMapType(nil, field)
			_, keywire := g.GoType(nil, m.KeyAliasField)
			valuegoTyp, _ := g.GoType(nil, m.ValueField)
			valuegoAliasTyp, valuewire := g.GoType(nil, m.ValueAliasField)
			_, fieldwire := g.GoType(nil, field)

			nullable, valuegoTyp, valuegoAliasTyp = generator.GoMapValueTypes(field, m.ValueField, valuegoTyp, valuegoAliasTyp)

			fieldKeySize := keySize(field.GetNumber(), wireToType(fieldwire))
			keyKeySize := keySize(1, wireToType(keywire))
			valueKeySize := keySize(2, wireToType(valuewire))
			g.P(`for k, v := range m.`, fieldname, ` { `)
			g.In()
			g.P(`_ = k`)
			g.P(`_ = v`)
			sum := []string{strconv.Itoa(keyKeySize)}
			switch m.KeyField.GetType() {
			case descriptor.FieldDescriptorProto_TYPE_DOUBLE,
				descriptor.FieldDescriptorProto_TYPE_FIXED64,
				descriptor.FieldDescriptorProto_TYPE_SFIXED64:
				sum = append(sum, `8`)
			case descriptor.FieldDescriptorProto_TYPE_FLOAT,
				descriptor.FieldDescriptorProto_TYPE_FIXED32,
				descriptor.FieldDescriptorProto_TYPE_SFIXED32:
				sum = append(sum, `4`)
			case descriptor.FieldDescriptorProto_TYPE_INT64,
				descriptor.FieldDescriptorProto_TYPE_UINT64,
				descriptor.FieldDescriptorProto_TYPE_UINT32,
				descriptor.FieldDescriptorProto_TYPE_ENUM,
				descriptor.FieldDescriptorProto_TYPE_INT32:
				sum = append(sum, `sov`+g.localName+`(uint64(k))`)
			case descriptor.FieldDescriptorProto_TYPE_BOOL:
				sum = append(sum, `1`)
			case descriptor.FieldDescriptorProto_TYPE_STRING,
				descriptor.FieldDescriptorProto_TYPE_BYTES:
				sum = append(sum, `len(k)+sov`+g.localName+`(uint64(len(k)))`)
			case descriptor.FieldDescriptorProto_TYPE_SINT32,
				descriptor.FieldDescriptorProto_TYPE_SINT64:
				sum = append(sum, `soz`+g.localName+`(uint64(k))`)
			}
			switch m.ValueField.GetType() {
			case descriptor.FieldDescriptorProto_TYPE_DOUBLE,
				descriptor.FieldDescriptorProto_TYPE_FIXED64,
				descriptor.FieldDescriptorProto_TYPE_SFIXED64:
				sum = append(sum, strconv.Itoa(valueKeySize))
				sum = append(sum, strconv.Itoa(8))
			case descriptor.FieldDescriptorProto_TYPE_FLOAT,
				descriptor.FieldDescriptorProto_TYPE_FIXED32,
				descriptor.FieldDescriptorProto_TYPE_SFIXED32:
				sum = append(sum, strconv.Itoa(valueKeySize))
				sum = append(sum, strconv.Itoa(4))
			case descriptor.FieldDescriptorProto_TYPE_INT64,
				descriptor.FieldDescriptorProto_TYPE_UINT64,
				descriptor.FieldDescriptorProto_TYPE_UINT32,
				descriptor.FieldDescriptorProto_TYPE_ENUM,
				descriptor.FieldDescriptorProto_TYPE_INT32:
				sum = append(sum, strconv.Itoa(valueKeySize))
				sum = append(sum, `sov`+g.localName+`(uint64(v))`)
			case descriptor.FieldDescriptorProto_TYPE_BOOL:
				sum = append(sum, strconv.Itoa(valueKeySize))
				sum = append(sum, `1`)
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				sum = append(sum, strconv.Itoa(valueKeySize))
				sum = append(sum, `len(v)+sov`+g.localName+`(uint64(len(v)))`)
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				if gogoproto.IsCustomType(field) {
					g.P(`l = 0`)
					if nullable {
						g.P(`if v != nil {`)
						g.In()
					}
					g.P(`l = v.`, sizeName, `()`)
					g.P(`l += `, strconv.Itoa(valueKeySize), ` + sov`+g.localName+`(uint64(l))`)
					if nullable {
						g.Out()
						g.P(`}`)
					}
					sum = append(sum, `l`)
				} else {
					g.P(`l = 0`)
					if proto3 {
						g.P(`if len(v) > 0 {`)
					} else {
						g.P(`if v != nil {`)
					}
					g.In()
					g.P(`l = `, strconv.Itoa(valueKeySize), ` + len(v)+sov`+g.localName+`(uint64(len(v)))`)
					g.Out()
					g.P(`}`)
					sum = append(sum, `l`)
				}
			case descriptor.FieldDescriptorProto_TYPE_SINT32,
				descriptor.FieldDescriptorProto_TYPE_SINT64:
				sum = append(sum, strconv.Itoa(valueKeySize))
				sum = append(sum, `soz`+g.localName+`(uint64(v))`)
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				stdSizeCall, stdOk := g.std(m.ValueAliasField, "v")
				if nullable {
					g.P(`l = 0`)
					g.P(`if v != nil {`)
					g.In()
					if stdOk {
						g.P(`l = `, stdSizeCall)
					} else if valuegoTyp != valuegoAliasTyp {
						g.P(`l = ((`, valuegoTyp, `)(v)).`, sizeName, `()`)
					} else {
						g.P(`l = v.`, sizeName, `()`)
					}
					g.P(`l += `, strconv.Itoa(valueKeySize), ` + sov`+g.localName+`(uint64(l))`)
					g.Out()
					g.P(`}`)
					sum = append(sum, `l`)
				} else {
					if stdOk {
						g.P(`l = `, stdSizeCall)
					} else if valuegoTyp != valuegoAliasTyp {
						g.P(`l = ((*`, valuegoTyp, `)(&v)).`, sizeName, `()`)
					} else {
						g.P(`l = v.`, sizeName, `()`)
					}
					sum = append(sum, strconv.Itoa(valueKeySize))
					sum = append(sum, `l+sov`+g.localName+`(uint64(l))`)
				}
			}
			g.P(`mapEntrySize := `, strings.Join(sum, "+"))
			g.P(`n+=mapEntrySize+`, fieldKeySize, `+sov`, g.localName, `(uint64(mapEntrySize))`)
			g.Out()
			g.P(`}`)
		} else if repeated {
			g.P(`for _, e := range m.`, fieldname, ` { `)
			g.In()
			stdSizeCall, stdOk := g.std(field, "e")
			if stdOk {
				g.P(`l=`, stdSizeCall)
			} else {
				g.P(`l=e.`, sizeName, `()`)
			}
			g.P(`n+=`, strconv.Itoa(key), `+l+sov`, g.localName, `(uint64(l))`)
			g.Out()
			g.P(`}`)
		} else {
			stdSizeCall, stdOk := g.std(field, "m."+fieldname)
			if stdOk {
				g.P(`l=`, stdSizeCall)
			} else {
				g.P(`l=m.`, fieldname, `.`, sizeName, `()`)
			}
			g.P(`n+=`, strconv.Itoa(key), `+l+sov`, g.localName, `(uint64(l))`)
		}
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		if !gogoproto.IsCustomType(field) {
			if repeated {
				g.P(`for _, b := range m.`, fieldname, ` { `)
				g.In()
				g.P(`l = len(b)`)
				g.P(`n+=`, strconv.Itoa(key), `+l+sov`, g.localName, `(uint64(l))`)
				g.Out()
				g.P(`}`)
			} else if proto3 {
				g.P(`l=len(m.`, fieldname, `)`)
				g.P(`if l > 0 {`)
				g.In()
				g.P(`n+=`, strconv.Itoa(key), `+l+sov`, g.localName, `(uint64(l))`)
				g.Out()
				g.P(`}`)
			} else {
				g.P(`l=len(m.`, fieldname, `)`)
				g.P(`n+=`, strconv.Itoa(key), `+l+sov`, g.localName, `(uint64(l))`)
			}
		} else {
			if repeated {
				g.P(`for _, e := range m.`, fieldname, ` { `)
				g.In()
				g.P(`l=e.`, sizeName, `()`)
				g.P(`n+=`, strconv.Itoa(key), `+l+sov`, g.localName, `(uint64(l))`)
				g.Out()
				g.P(`}`)
			} else {
				g.P(`l=m.`, fieldname, `.`, sizeName, `()`)
				g.P(`n+=`, strconv.Itoa(key), `+l+sov`, g.localName, `(uint64(l))`)
			}
		}
	case descriptor.FieldDescriptorProto_TYPE_SINT32,
		descriptor.FieldDescriptorProto_TYPE_SINT64:
		if packed {
			g.P(`l = 0`)
			g.P(`for _, e := range m.`, fieldname, ` {`)
			g.In()
			g.P(`l+=soz`, g.localName, `(uint64(e))`)
			g.Out()
			g.P(`}`)
			g.P(`n+=`, strconv.Itoa(key), `+sov`, g.localName, `(uint64(l))+l`)
		} else if repeated {
			g.P(`for _, e := range m.`, fieldname, ` {`)
			g.In()
			g.P(`n+=`, strconv.Itoa(key), `+soz`, g.localName, `(uint64(e))`)
			g.Out()
			g.P(`}`)
		} else if proto3 {
			g.P(`if m.`, fieldname, ` != 0 {`)
			g.In()
			g.P(`n+=`, strconv.Itoa(key), `+soz`, g.localName, `(uint64(m.`, fieldname, `))`)
			g.Out()
			g.P(`}`)
		} else if nullable {
			g.P(`n+=`, strconv.Itoa(key), `+soz`, g.localName, `(uint64(*m.`, fieldname, `))`)
		} else {
			g.P(`n+=`, strconv.Itoa(key), `+soz`, g.localName, `(uint64(m.`, fieldname, `))`)
		}
	default:
		panic("not implemented")
	}
	if repeated || doNilCheck {
		g.Out()
		g.P(`}`)
	}
}

func (g *gogo) GenerateSize(file *generator.FileDescriptor) {
	for _, message := range file.Messages() {
		sizeName := ""
		if gogoproto.IsSizer(file.FileDescriptorProto, message.Proto.DescriptorProto) && gogoproto.IsProtoSizer(file.FileDescriptorProto, message.Proto.DescriptorProto) {
			fmt.Fprintf(os.Stderr, "ERROR: message %v cannot support both sizer and protosizer plugins\n", generator.CamelCase(message.Proto.GetName()))
			os.Exit(1)
		}
		if gogoproto.IsSizer(file.FileDescriptorProto, message.Proto.DescriptorProto) {
			sizeName = "XSize"
		} else if gogoproto.IsProtoSizer(file.FileDescriptorProto, message.Proto.DescriptorProto) {
			sizeName = "ProtoSize"
		} else {
			continue
		}
		if message.Proto.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}
		g.atleastOne = true
		ccTypeName := generator.CamelCaseSlice(message.Proto.TypeName())
		g.P(`func (m *`, ccTypeName, `) `, sizeName, `() (n int) {`)
		g.In()
		g.P(`if m == nil {`)
		g.In()
		g.P(`return 0`)
		g.Out()
		g.P(`}`)
		g.P(`var l int`)
		g.P(`_ = l`)
		oneofs := make(map[string]struct{})
		for _, field := range message.Fields {
			oneof := field.Proto.OneofIndex != nil
			if !oneof {
				proto3 := gogoproto.IsProto3(file.FileDescriptorProto)
				g.generateField(proto3, file, message, field, sizeName)
			} else {
				fieldname := g.GetFieldName(message.Proto, field)
				if _, ok := oneofs[fieldname]; ok {
					continue
				} else {
					oneofs[fieldname] = struct{}{}
				}
				g.P(`if m.`, fieldname, ` != nil {`)
				g.In()
				g.P(`n+=m.`, fieldname, `.`, sizeName, `()`)
				g.Out()
				g.P(`}`)
			}
		}
		if message.Proto.DescriptorProto.HasExtension() {
			if gogoproto.HasExtensionsMap(file.FileDescriptorProto, message.Proto.DescriptorProto) {
				g.P(`n += `, g.protoPkg.Use(), `.SizeOfInternalExtension(m)`)
			} else {
				g.P(`if m.XXX_extensions != nil {`)
				g.In()
				g.P(`n+=len(m.XXX_extensions)`)
				g.Out()
				g.P(`}`)
			}
		}
		if gogoproto.HasUnrecognized(file.FileDescriptorProto, message.Proto.DescriptorProto) {
			g.P(`if m.XXX_unrecognized != nil {`)
			g.In()
			g.P(`n+=len(m.XXX_unrecognized)`)
			g.Out()
			g.P(`}`)
		}
		g.P(`return n`)
		g.Out()
		g.P(`}`)
		g.P()

		//Generate Size methods for oneof fields
		//m := proto.Clone(message).(*generator.MessageDescriptor)
		for _, f := range message.Fields {
			oneof := f.Proto.OneofIndex != nil
			if !oneof {
				continue
			}
			ccTypeName := g.OneOfTypeName(message, f)
			g.P(`func (m *`, ccTypeName, `) `, sizeName, `() (n int) {`)
			g.In()
			g.P(`if m == nil {`)
			g.In()
			g.P(`return 0`)
			g.Out()
			g.P(`}`)
			g.P(`var l int`)
			g.P(`_ = l`)
			vanity2.TurnOffNullableForNativeTypes(f.Proto)
			g.generateField(false, file, message, f, sizeName)
			g.P(`return n`)
			g.Out()
			g.P(`}`)
		}
	}

	if !g.atleastOne {
		return
	}

	g.sizeVarint()
	g.sizeZigZag()

}
