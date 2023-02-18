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

// Package trace provides an interface for distributed tracing
package trace

import (
	"context"
	"time"

	"github.com/spf13/pflag"
	"github.com/vine-io/vine/util/context/metadata"
)

// SpanType describe the nature of the trace span
type SpanType int

const (
	// SpanTypeRequestInbound is a span created when serving a request
	SpanTypeRequestInbound SpanType = iota
	// SpanTypeRequestOutbound is a span created when making a service call
	SpanTypeRequestOutbound
)

var (
	DefaultTracer Tracer = new(noop)

	Flag = pflag.NewFlagSet("trace", pflag.ExitOnError)
)

func init() {
	Flag.String("tracer.default", "", "Trace for vine")
	Flag.String("tracer.address", "", "Comma-separated list of tracer addresses")
}

// Tracer is an interface for distributed tracing
type Tracer interface {
	// Start a trace
	Start(ctx context.Context, name string) (context.Context, *Span)
	// Finish the trace
	Finish(*Span) error
	// Read the traces
	Read(...ReadOption) ([]*Span, error)
}

// Span is used to record an entry
type Span struct {
	// Id of the trace
	Trace string
	// name of the span
	Name string
	// id of the span
	Id string
	// parent span id
	Parent string
	// Start time
	Started time.Time
	// Duration in nano seconds
	Duration time.Duration
	// associated data
	Metadata map[string]string
	// Type
	Type SpanType
}

const (
	traceIDKey = "Vine-Trace-Id"
	spanIDKey  = "Vine-Span-Id"
)

// FromContext returns a span from context
func FromContext(ctx context.Context) (string, string, bool) {
	traceID, traceOk := metadata.Get(ctx, traceIDKey)
	vineID, vineOk := metadata.Get(ctx, "Vine-Id")
	if !traceOk && !vineOk {
		return "", "", false
	}
	if !traceOk {
		traceID = vineID
	}
	parentSpanID, ok := metadata.Get(ctx, spanIDKey)
	return traceID, parentSpanID, ok
}

// ToContext saves the trace and span ids in the context
func ToContext(ctx context.Context, traceID, parentSpanID string) context.Context {
	return metadata.MergeContext(ctx, map[string]string{
		traceIDKey: traceID,
		spanIDKey:  parentSpanID,
	}, true)
}

type noop struct{}

func (n *noop) Init(...Option) error {
	return nil
}

func (n *noop) Start(ctx context.Context, name string) (context.Context, *Span) {
	return nil, nil
}

func (n *noop) Finish(*Span) error {
	return nil
}

func (n *noop) Read(...ReadOption) ([]*Span, error) {
	return nil, nil
}

func Start(ctx context.Context, name string) (context.Context, *Span) {
	return DefaultTracer.Start(ctx, name)
}

func Finish(span *Span) error {
	return DefaultTracer.Finish(span)
}

func Read(opts ...ReadOption) ([]*Span, error) {
	return DefaultTracer.Read(opts...)
}
