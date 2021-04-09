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

package debug

import (
	"time"

	"github.com/lack-io/vine/service/debug/log"
)

type serviceLog struct {
	Client *debugClient
}

// Read reads log entries from the logger
func (s *serviceLog) Read(opts ...log.ReadOption) ([]log.Record, error) {
	var options log.ReadOptions
	for _, o := range opts {
		o(&options)
	}

	stream, err := s.Client.Log(options.Since, options.Count, false)
	if err != nil {
		return nil, err
	}
	defer stream.Stop()

	// stream the records until nothing is left
	var records []log.Record

	for record := range stream.Chan() {
		records = append(records, record)
	}

	return records, nil
}

// There is no write support
func (s *serviceLog) Write(r log.Record) error {
	return nil
}

// Stream log records
func (s *serviceLog) Stream() (log.Stream, error) {
	return s.Client.Log(time.Time{}, 0, true)
}

// NewLog returns a new log interface
func NewLog(opts ...log.Option) log.Log {
	var options log.Options
	for _, o := range opts {
		o(&options)
	}

	name := options.Name

	// set the default name
	if len(name) == 0 {
		name = "go.vine.debug"
	}

	return &serviceLog{
		Client: NewClient(name),
	}
}
