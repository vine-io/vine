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

package grpc

import (
	b "bytes"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	json "github.com/json-iterator/go"
	"github.com/oxtoacart/bpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	"github.com/vine-io/vine/core/codec"
	"github.com/vine-io/vine/core/codec/bytes"
	"github.com/vine-io/vine/util/jsonpb"
)

type jsonCodec struct{}
type protoCodec struct{}
type bytesCodec struct{}
type wrapCodec struct{ encoding.Codec }

var jsonpbMarshaler = &jsonpb.Marshaler{}
var useNumber bool

// create buffer pool with 16 instances each preallocated with 256 bytes
var bufferPool = bpool.NewSizedBufferPool(16, 256)

var (
	defaultGRPCCodecs = map[string]encoding.Codec{
		"application/json":         jsonCodec{},
		"application/proto":        protoCodec{},
		"application/protobuf":     protoCodec{},
		"application/octet-stream": protoCodec{},
		"application/grpc":         protoCodec{},
		"application/grpc+json":    jsonCodec{},
		"application/grpc+proto":   protoCodec{},
		"application/grpc+bytes":   bytesCodec{},
	}
)

// UserNumber fix unmarshal Number(8234567890123456789) to interface(8.234567890123457e+18)
func UserNumber() {
	useNumber = true
}

func (w wrapCodec) String() string {
	return w.Codec.Name()
}

func (w wrapCodec) Marshal(v interface{}) ([]byte, error) {
	b, ok := v.(*bytes.Frame)
	if ok {
		return b.Data, nil
	}
	return w.Codec.Marshal(v)
}

func (w wrapCodec) Unmarshal(data []byte, v interface{}) error {
	b, ok := v.(*bytes.Frame)
	if ok {
		b.Data = data
		return nil
	}
	return w.Codec.Unmarshal(data, v)
}

func (protoCodec) Marshal(v interface{}) ([]byte, error) {
	switch m := v.(type) {
	case *bytes.Frame:
		return m.Data, nil
	case proto.Message:
		return proto.Marshal(m)
	}
	return nil, fmt.Errorf("failed to marshal: %v is not type of *bytes.Frame or proto.Message", v)
}

func (protoCodec) Unmarshal(data []byte, v interface{}) error {
	m, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("failed to unmarshal: %v is not type of proto.Message", v)
	}
	return proto.Unmarshal(data, m)
}

func (protoCodec) Name() string {
	return "proto"
}

func (bytesCodec) Marshal(v interface{}) ([]byte, error) {
	b, ok := v.(*[]byte)
	if !ok {
		return nil, fmt.Errorf("failed to marshal: %v is not type of *[]byte", v)
	}
	return *b, nil
}

func (bytesCodec) Unmarshal(data []byte, v interface{}) error {
	b, ok := v.(*[]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal: %v is not type of *[]byte", v)
	}
	*b = data
	return nil
}

func (bytesCodec) Name() string {
	return "bytes"
}

func (jsonCodec) Marshal(v interface{}) ([]byte, error) {
	if b, ok := v.(*bytes.Frame); ok {
		return b.Data, nil
	}

	if pb, ok := v.(proto.Message); ok {
		buf := bufferPool.Get()
		defer bufferPool.Put(buf)
		if err := jsonpbMarshaler.Marshal(buf, pb); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, bb interface{}) error {
	if len(data) == 0 {
		return nil
	}
	if v, ok := bb.(*bytes.Frame); ok {
		v.Data = data
		return nil
	}
	if pb, ok := bb.(proto.Message); ok {
		return jsonpb.Unmarshal(b.NewReader(data), pb)
	}

	dec := json.NewDecoder(b.NewReader(data))
	if useNumber {
		dec.UseNumber()
	}
	return dec.Decode(bb)
}

func (jsonCodec) Name() string {
	return "json"
}

type grpcCodec struct {
	// headers
	id       string
	target   string
	method   string
	endpoint string

	s grpc.ClientStream
	c encoding.Codec
}

func (g *grpcCodec) ReadHeader(m *codec.Message, mt codec.MessageType) error {
	md, err := g.s.Header()
	if err != nil {
		return err
	}
	if m == nil {
		m = new(codec.Message)
	}
	if m.Header == nil {
		m.Header = make(map[string]string, len(md))
	}
	for k, v := range md {
		m.Header[k] = strings.Join(v, ",")
	}
	m.Id = g.id
	m.Target = g.target
	m.Method = g.method
	m.Endpoint = g.endpoint
	return nil
}

func (g *grpcCodec) ReadBody(v interface{}) error {
	if f, ok := v.(*bytes.Frame); ok {
		return g.s.RecvMsg(f)
	}
	return g.s.RecvMsg(v)
}

func (g *grpcCodec) Write(m *codec.Message, v interface{}) error {
	// if we don't have a body
	if v != nil {
		return g.s.SendMsg(v)
	}
	// write the body using the framing codec
	return g.s.SendMsg(&bytes.Frame{Data: m.Body})
}

func (g *grpcCodec) Close() error {
	return g.s.CloseSend()
}

func (g *grpcCodec) String() string {
	return g.c.Name()
}
