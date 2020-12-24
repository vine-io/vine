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

package log

import (
	"context"
	"sync"

	"github.com/lack-io/vine/internal/debug/log"
	pb "github.com/lack-io/vine/proto/debug/log"
	"github.com/lack-io/vine/proto/errors"
)

type Log struct {
	// per service log
	sync.RWMutex
	Logs map[string]log.Log

	// Ability to create new logger
	New func(string) log.Log
}

func (l *Log) Read(ctx context.Context, req *pb.ReadRequest, rsp *pb.ReadResponse) error {
	if len(req.Service) == 0 {
		return errors.BadRequest("go.vine.debug.log", "Invalid service name")
	}

	l.Lock()
	defer l.Unlock()

	// get the service log
	serviceLog, ok := l.Logs[req.Service]
	if !ok {
		serviceLog = l.New(req.Service)
		l.Logs[req.Service] = serviceLog
	}

	// TODO: specify how many log records to read
	records, err := serviceLog.Read()
	if err != nil {
		return err
	}

	// append to records
	for _, rec := range records {
		rsp.Records = append(rsp.Records, &pb.Record{
			Timestamp: rec.Timestamp.Unix(),
			Metadata:  rec.Metadata,
			Message:   rec.Message.(string),
		})
	}

	return nil
}
