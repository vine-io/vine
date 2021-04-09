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

// Package memory provides an in memory log buffer
package memory

import (
	"fmt"

	"github.com/lack-io/vine/service/debug/log"
	"github.com/lack-io/vine/util/ring"
)

var (
	// DefaultSize of the logger buffer
	DefaultSize = 1024
)

// memoryLog is default vine log
type memoryLog struct {
	*ring.Buffer
}

// NewLog returns default Logger with
func NewLog(opts ...log.Option) log.Log {
	// get default options
	options := log.DefaultOptions()

	// apply requested options
	for _, o := range opts {
		o(&options)
	}

	return &memoryLog{
		Buffer: ring.New(options.Size),
	}
}

// Write writes logs into logger
func (l *memoryLog) Write(r log.Record) error {
	l.Buffer.Put(fmt.Sprint(r.Message))
	return nil
}

// Read reads logs and returns them
func (l *memoryLog) Read(opts ...log.ReadOption) ([]log.Record, error) {
	options := log.ReadOptions{}
	// initialize the read options
	for _, o := range opts {
		o(&options)
	}

	var entries []*ring.Entry
	// if Since options ha sbeen specified we honor it
	if !options.Since.IsZero() {
		entries = l.Buffer.Since(options.Since)
	}

	// only if we specified valid count constraint
	// do we end up doing some serious if-else kung-fu
	// if since constraint has been provided
	// we return *count* number of logs since the given timestamp;
	// otherwise we return last count number of logs
	if options.Count > 0 {
		switch len(entries) > 0 {
		case true:
			// if we request fewer logs than what since constraint gives us
			if options.Count < len(entries) {
				entries = entries[0:options.Count]
			}
		default:
			entries = l.Buffer.Get(options.Count)
		}
	}

	records := make([]log.Record, 0, len(entries))
	for _, entry := range entries {
		record := log.Record{
			Timestamp: entry.Timestamp,
			Message:   entry.Value,
		}
		records = append(records, record)
	}

	return records, nil
}

// Stream returns channel for reading log records
// along with a stop channel, close it when done
func (l *memoryLog) Stream() (log.Stream, error) {
	// get stream channel from ring buffer
	stream, stop := l.Buffer.Stream()
	// make a buffered channel
	records := make(chan log.Record, 128)
	// get last 10 records
	last10 := l.Buffer.Get(10)

	// stream the log records
	go func() {
		// first send last 10 records
		for _, entry := range last10 {
			records <- log.Record{
				Timestamp: entry.Timestamp,
				Message:   entry.Value,
				Metadata:  make(map[string]string),
			}
		}
		// now stream continuously
		for entry := range stream {
			records <- log.Record{
				Timestamp: entry.Timestamp,
				Message:   entry.Value,
				Metadata:  make(map[string]string),
			}
		}
	}()

	return &logStream{
		stream: records,
		stop:   stop,
	}, nil
}
