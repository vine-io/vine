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

package service

import (
	"github.com/lack-io/vine/registry"
	pb "github.com/lack-io/vine/registry/service/proto"
)

type serviceWatcher struct {
	stream pb.Registry_WatchService
	closed chan bool
}

func (s *serviceWatcher) Next() (*registry.Result, error) {
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

	return &registry.Result{
		Action:  r.Action,
		Service: toService(r.Service),
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

func newWatcher(stream pb.Registry_WatchService) registry.Watcher {
	return &serviceWatcher{
		stream: stream,
		closed: make(chan bool),
	}
}
