// MIT License
//
// Copyright (c) 2020 The vine Authors
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

// Package grpc provides a grpc codec
package grpc

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gogo/protobuf/proto"
	json "github.com/json-iterator/go"

	"github.com/vine-io/vine/core/codec"
)

type Codec struct {
	Conn        io.ReadWriteCloser
	ContentType string
}

func (c *Codec) ReadHeader(m *codec.Message, t codec.MessageType) error {
	if ct := m.Header["Content-Type"]; len(ct) > 0 {
		c.ContentType = ct
	}

	if ct := m.Header["content-type"]; len(ct) > 0 {
		c.ContentType = ct
	}

	// service method
	path := m.Header[":path"]
	if len(path) == 0 || path[0] != '/' {
		m.Target = m.Header["Vine-Service"]
		m.Endpoint = m.Header["Vine-Endpoint"]
	} else {
		// [ , a.package.Foo, Bar]
		parts := strings.Split(path, "/")
		if len(parts) != 3 {
			return errors.New("unknown request path")
		}
		service := strings.Split(parts[1], ".")
		m.Endpoint = strings.Join([]string{service[len(service)-1], parts[2]}, ".")
		m.Target = strings.Join(service[:len(service)-1], ".")
	}

	return nil
}

func (c *Codec) ReadBody(b interface{}) error {
	// no body
	if b == nil {
		return nil
	}

	_, buf, err := decode(c.Conn)
	if err != nil {
		return err
	}

	switch c.ContentType {
	case "application/grpc+json":
		return json.Unmarshal(buf, b)
	case "application/grpc+proto", "application/grpc":
		return proto.Unmarshal(buf, b.(proto.Message))
	}

	return errors.New("unsupported Content-Type")
}

func (c *Codec) Write(m *codec.Message, b interface{}) error {
	var buf []byte
	var err error

	if ct := m.Header["Content-Type"]; len(ct) > 0 {
		c.ContentType = ct
	}

	if ct := m.Header["content-type"]; len(ct) > 0 {
		c.ContentType = ct
	}

	switch m.Type {
	case codec.Request:
		parts := strings.Split(m.Endpoint, ".")
		m.Header[":method"] = "POST"
		m.Header[":path"] = fmt.Sprintf("/%s.%s/%s", m.Target, parts[0], parts[1])
		m.Header[":proto"] = "HTTP/.0"
		m.Header["te"] = "trailers"
		m.Header["user-agent"] = "grpc-go/1.0.0"
		m.Header[":authority"] = m.Target
		m.Header["content-type"] = c.ContentType
	case codec.Response:
		m.Header["Trailer"] = "grpc-status" // , grpc-message
		m.Header["content-type"] = c.ContentType
		m.Header[":status"] = "200"
		m.Header["grpc-status"] = "0"
		// m.Header["grpc-message"] = ""
	case codec.Error:
		m.Header["Trailer"] = "grpc-status, grpc-message"
		// vine end of stream
		if m.Error == "EOS" {
			m.Header["grpc-status"] = "0"
		} else {
			m.Header["grpc-message"] = m.Error
			m.Header["grpc-status"] = "13"
		}
		return nil
	}

	// marshal content
	switch c.ContentType {
	case "application/grpc+json":
		buf, err = json.Marshal(b)
	case "application/grpc+proto", "application/grpc":
		pb, ok := b.(proto.Message)
		if ok {
			buf, err = proto.Marshal(pb)
		}
	default:
		err = errors.New("unsupported Content-Type")
	}
	// check error
	if err != nil {
		m.Header["grpc-status"] = "8"
		m.Header["grpc-message"] = err.Error()
		return err
	}

	if len(buf) == 0 {
		return nil
	}

	return encode(0, buf, c.Conn)
}

func (c *Codec) Close() error {
	return c.Conn.Close()
}

func (c *Codec) String() string {
	return "grpc"
}

func NewCodec(c io.ReadWriteCloser) codec.Codec {
	return &Codec{
		Conn:        c,
		ContentType: "application/grpc",
	}
}
