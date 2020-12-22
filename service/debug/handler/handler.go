// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package handler implements service debug handler embedded in vine services
package handler

import (
	"context"
	"time"

	"github.com/lack-io/vine/internal/debug/log"
	"github.com/lack-io/vine/internal/debug/stats"
	"github.com/lack-io/vine/internal/debug/trace"
	pb "github.com/lack-io/vine/proto/debug"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/debug"
)

// NewHandler returns an instance of the Debug Handler
func NewHandler(c client.Client) *Debug {
	return &Debug{
		log:   debug.DefaultLog,
		stats: debug.DefaultStats,
		trace: debug.DefaultTracer,
	}
}

type Debug struct {
	// must honour the debug handler
	pb.DebugHandler
	// the logger for retrieving logs
	log log.Log
	// the stats collector
	stats stats.Stats
	// the tracer
	trace trace.Tracer
}

func (d *Debug) Health(ctx context.Context, req *pb.HealthRequest, rsp *pb.HealthResponse) error {
	rsp.Status = "ok"
	return nil
}

func (d *Debug) Stats(ctx context.Context, req *pb.StatsRequest, rsp *pb.StatsResponse) error {
	stats, err := d.stats.Read()
	if err != nil {
		return err
	}

	if len(stats) == 0 {
		return nil
	}

	// write the response values
	rsp.Timestamp = uint64(stats[0].Timestamp)
	rsp.Started = uint64(stats[0].Started)
	rsp.Uptime = uint64(stats[0].Uptime)
	rsp.Memory = stats[0].Memory
	rsp.Gc = stats[0].GC
	rsp.Threads = stats[0].Threads
	rsp.Requests = stats[0].Requests
	rsp.Errors = stats[0].Errors

	return nil
}

func (d *Debug) Trace(ctx context.Context, req *pb.TraceRequest, rsp *pb.TraceResponse) error {
	traces, err := d.trace.Read(trace.ReadTrace(req.Id))
	if err != nil {
		return err
	}

	for _, t := range traces {
		var typ pb.SpanType
		switch t.Type {
		case trace.SpanTypeRequestInbound:
			typ = pb.SpanType_INBOUND
		case trace.SpanTypeRequestOutbound:
			typ = pb.SpanType_OUTBOUND
		}
		rsp.Spans = append(rsp.Spans, &pb.Span{
			Trace:    t.Trace,
			Id:       t.Id,
			Parent:   t.Parent,
			Name:     t.Name,
			Started:  uint64(t.Started.UnixNano()),
			Duration: uint64(t.Duration.Nanoseconds()),
			Type:     typ,
			Metadata: t.Metadata,
		})
	}

	return nil
}

// Log returns some log lines
func (d *Debug) Log(ctx context.Context, req pb.LogRequest, rsp *pb.LogResponse) error {
	var options []log.ReadOption

	since := time.Unix(req.Since, 0)
	if !since.IsZero() {
		options = append(options, log.Since(since))
	}

	count := int(req.Count)
	if count > 0 {
		options = append(options, log.Count(count))
	}

	// get the log records
	records, err := d.log.Read(options...)
	if err != nil {
		return err
	}

	for _, record := range records {
		// copy metadata
		metadata := make(map[string]string)
		for k, v := range record.Metadata {
			metadata[k] = v
		}
		// send record
		rsp.Records = append(rsp.Records, &pb.Record{
			Timestamp: record.Timestamp.Unix(),
			Message:   record.Message.(string),
			Metadata:  metadata,
		})
	}

	return nil
}
