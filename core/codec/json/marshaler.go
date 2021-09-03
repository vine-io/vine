// Copyright 2020 The vine Authors
//
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

package json

import (
	"bytes"

	"github.com/gogo/protobuf/proto"
	json "github.com/json-iterator/go"
	"github.com/oxtoacart/bpool"
	mjsonpb "github.com/vine-io/vine/util/jsonpb"
)

var jsonbMarshaler = &mjsonpb.Marshaler{}

// create buffer pool with 16 instances each preallocated with 256 bytes
var bufferPool = bpool.NewSizedBufferPool(16, 256)

type Marshaler struct{}

func (j Marshaler) Marshal(v interface{}) ([]byte, error) {
	if pb, ok := v.(proto.Message); ok {
		buf := bufferPool.Get()
		defer bufferPool.Put(buf)
		if err := jsonbMarshaler.Marshal(buf, pb); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	return json.Marshal(v)
}

func (j Marshaler) Unmarshal(d []byte, v interface{}) error {
	if pb, ok := v.(proto.Message); ok {
		return mjsonpb.Unmarshal(bytes.NewReader(d), pb)
	}
	return json.Unmarshal(d, v)
}

func (j Marshaler) String() string {
	return "json"
}
