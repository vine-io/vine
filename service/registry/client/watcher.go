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

package client

import (
	"time"

	pb "github.com/lack-io/vine/proto/registry"
	"github.com/lack-io/vine/service/registry"
	"github.com/lack-io/vine/service/registry/util"
)

type serviceWatcher struct {
	stream pb.Registry_WatchService
	closed chan bool
}

func (s *serviceWatcher) Next() (*registry.Result, error) {
	var i int

	for {
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

		// result is nil
		if r == nil {
			i++

			// only process for 3 attempts if nil
			if i > 3 {
				return nil, registry.ErrWatcherStopped
			}

			// wait a moment
			time.Sleep(time.Second)

			// otherwise continue
			continue
		}

		return &registry.Result{
			Action:  r.Action,
			Service: util.ToService(r.Service),
		}, nil
	}
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
