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

package grpc

import (
	"encoding/binary"
	"fmt"
	"io"
)

var (
	MaxMessageSize = 1024 * 1024 * 32 // 32Mb
	maxInt         = int(^uint(0) >> 1)
)

func decode(r io.Reader) (uint8, []byte, error) {
	header := make([]byte, 5)

	// read the header
	if _, err := r.Read(header[:]); err != nil {
		return uint8(0), nil, err
	}

	// get encoding format e.g compressed
	cf := uint8(header[0])

	// get message length
	length := binary.BigEndian.Uint32(header[1:])

	// no encoding format
	if length == 0 {
		return cf, nil, nil
	}

	if int64(length) > int64(maxInt) {
		return cf, nil, fmt.Errorf("grpc: received message larger than max length allowed on current machine (%d vs. %d)", length, maxInt)
	}
	if int(length) > MaxMessageSize {
		return cf, nil, fmt.Errorf("grpc: received message larger than max (%d vs. %d)", length, MaxMessageSize)
	}

	msg := make([]byte, int64(length))

	if _, err := r.Read(msg); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return cf, nil, err
	}

	return cf, msg, nil
}

func encode(cf uint8, buf []byte, w io.Writer) error {
	header := make([]byte, 5)

	// set compress
	header[0] = byte(cf)

	// write length as header
	binary.BigEndian.PutUint32(header[1:], uint32(len(buf)))

	// read the header
	if _, err := w.Write(header[:]); err != nil {
		return err
	}

	// write the buffer
	_, err := w.Write(buf)
	return err
}
