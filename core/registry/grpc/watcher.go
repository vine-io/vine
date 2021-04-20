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

package grpc

import (
	"github.com/lack-io/vine/core/registry"
	regpb "github.com/lack-io/vine/proto/apis/registry"
	regsvcpb "github.com/lack-io/vine/proto/services/registry"
)

type serviceWatcher struct {
	stream regsvcpb.Registry_WatchService
	closed chan bool
}

func (s *serviceWatcher) Next() (*regpb.Result, error) {
	// check if closed
	select {
	case <-s.closed:
		return nil, registry.ErrWatcherStopped
	default:
	}

	r, err := s.stream.Recv()
	if err != nil {
		return nil, err
	}

	return &regpb.Result{
		Action:  r.Action,
		Service: r.Service,
	}, nil
}

func (s *serviceWatcher) Stop() {
	select {
	case <-s.closed:
		return
	default:
		close(s.closed)
		s.stream.Close()
	}
}

func newWatcher(stream regsvcpb.Registry_WatchService) registry.Watcher {
	return &serviceWatcher{
		stream: stream,
		closed: make(chan bool),
	}
}
