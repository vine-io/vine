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

package server

import "context"

// HandlerFunc represents a single method of a handler. It's used primarily
// for the wrappers. What's handed to the actual method is the concrete
// request and response types.
type HandlerFunc func(ctx context.Context, req Request, rsp interface{}) error

// SubscriberFunc represents a single method of a subscriber. It's used primarily
// for the wrappers. What's handed to the actual method is the concrete
// publication message.
type SubscriberFunc func(ctx context.Context, msg Message) error

// HandlerWrapper wraps the HandlerFunc and returns the equivalent
type HandlerWrapper func(HandlerFunc) HandlerFunc

// SubscriberWrapper wraps the SubscriberFunc and returns the equivalent
type SubscriberWrapper func(SubscriberFunc) SubscriberFunc

// StreamWrapper wraps a Stream interface and returns the equivalent.
// Because streams exist for the lifetime of a method invocation this
// is a convenient way to wrap a Stream as its in use for trace, monitoring.
// metrics, etc.
type StreamWrapper func(Stream) Stream
