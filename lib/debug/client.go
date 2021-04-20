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

// Package service provides the service log
package debug

import (
	"context"
	"time"

	"github.com/lack-io/vine/core/client"
	"github.com/lack-io/vine/lib/debug/log"
	pb "github.com/lack-io/vine/proto/services/debug"
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
