// Copyright 2020 The vine Authors
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
