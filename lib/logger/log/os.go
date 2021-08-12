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

package log

import (
	"sync"

	"github.com/google/uuid"

	"github.com/vine-io/vine/util/ring"
)

// Should stream from OS
type osLog struct {
	format FormatFunc
	once   sync.Once

	sync.RWMutex
	buffer *ring.Buffer
	subs   map[string]*osStream
}

type osStream struct {
	stream chan Record
}

// Read reads log entries from the logger
func (o *osLog) Read(...ReadOption) ([]Record, error) {
	var records []Record

	// read the last 100 records
	for _, v := range o.buffer.Get(100) {
		records = append(records, v.Value.(Record))
	}

	return records, nil
}

// Write writes records to log
func (o *osLog) Write(r Record) error {
	o.buffer.Put(r)
	return nil
}

// Stream log records
func (o *osLog) Stream() (Stream, error) {
	o.Lock()
	defer o.Unlock()

	// create stream
	st := &osStream{
		stream: make(chan Record, 128),
	}

	// save stream
	o.subs[uuid.New().String()] = st

	return st, nil
}

func (o *osStream) Chan() <-chan Record {
	return o.stream
}

func (o *osStream) Stop() error {
	return nil
}

func NewLog(opts ...Option) Log {
	options := Options{
		Format: DefaultFormat,
	}
	for _, o := range opts {
		o(&options)
	}

	l := &osLog{
		format: options.Format,
		buffer: ring.New(1024),
		subs:   make(map[string]*osStream),
	}

	return l
}
