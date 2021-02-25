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

package log

import (
	"sync"

	"github.com/google/uuid"

	"github.com/lack-io/vine/util/ring"
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
