// Copyright 2020 lack
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package codec

import (
	"errors"
	"io"
)

const (
	Error MessageType = iota
	Request
	Response
	Event
)

var (
	ErrInvalidMessage = errors.New("invalid message")
)

type MessageType int

// Takes in a connection/buffer and returns a new Codec
type NewCodec func(closer io.ReadWriteCloser) Codec

// Codec encodes/decodes various types of messages used within vine
// ReadHeader and ReadBody are called in pair to read requests/responses
// from the connection. Close is called when finished with the
// connection. ReadBody may be called with a nil argument to force the
// body to be read and discarded.
type Codec interface {
	Reader
	Writer
	Close() error
	String() string
}

type Reader interface {
	ReadHeader(*Message, MessageType) error
	ReadBody(interface{}) error
}

type Writer interface {
	Write(*Message, interface{}) error
}

// Marshaler is a simple encoding interface ued for the broker/transport
// where headers are not supported by the underlying implementation.
type Marshaler interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
	String() string
}

// Message represents detailed information about
// the communication, likely followed by the body.
// In the case of an error, body may be nil.
type Message struct {
	Id       string
	Type     MessageType
	Target   string
	Method   string
	Endpoint string
	Error    string

	// The values read from the socket
	Header map[string]string
	Body   []byte
}
