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

// Package service provides the service log
package service

import (
	"context"
	"time"

	"github.com/lack-io/vine/client"
	"github.com/lack-io/vine/debug/log"
	pb "github.com/lack-io/vine/debug/service/proto"
)

// Debug provides debug service client
type debugClient struct {
	Client pb.DebugService
}

func (d *debugClient) Trace() ([]*pb.Span, error) {
	rsp, err := d.Client.Trace(context.Background(), &pb.TraceRequest{})
	if err != nil {
		return nil, err
	}
	return rsp.Spans, nil
}

// Logs queries the services logs and returns a channel to read the logs from
func (d *debugClient) Log(since time.Time, count int, stream bool) (log.Stream, error) {
	req := &pb.LogRequest{}
	if !since.IsZero() {
		req.Since = since.Unix()
	}

	if count > 0 {
		req.Count = int64(count)
	}

	// set whether to stream
	req.Stream = stream

	// get the log stream
	serverStream, err := d.Client.Log(context.Background(), req)
	if err != nil {
		return nil, err
	}

	lg := &logStream{
		stream: make(chan log.Record),
		stop:   make(chan bool),
	}

	// go stream logs
	go d.streamLogs(lg, serverStream)

	return lg, nil
}

func (d *debugClient) streamLogs(lg *logStream, stream pb.Debug_LogService) {
	defer stream.Close()
	defer lg.Stop()

	for {
		resp, err := stream.Recv()
		if err != nil {
			break
		}

		metadata := make(map[string]string)
		for k, v := range resp.Metadata {
			metadata[k] = v
		}

		record := log.Record{
			Timestamp: time.Unix(resp.Timestamp, 0),
			Message:   resp.Message,
			Metadata:  metadata,
		}

		select {
		case <-lg.stop:
			return
		case lg.stream <- record:
		}
	}
}

// NewClient provides a debug client
func NewClient(name string) *debugClient {
	// create default client
	cli := client.DefaultClient

	return &debugClient{
		Client: pb.NewDebugService(name, cli),
	}
}
