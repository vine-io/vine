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
	"io"
	"sync"
	"time"

	pb "github.com/lack-io/vine/proto/services/router"
	"github.com/lack-io/vine/service/router"
)

type watcher struct {
	sync.RWMutex
	opts    router.WatchOptions
	resChan chan *router.Event
	done    chan struct{}
	stream  pb.Router_WatchService
}

func newWatcher(rsp pb.Router_WatchService, opts router.WatchOptions) (*watcher, error) {
	w := &watcher{
		opts:    opts,
		resChan: make(chan *router.Event),
		done:    make(chan struct{}),
		stream:  rsp,
	}

	go func() {
		for {
			select {
			case <-w.done:
				return
			default:
				if err := w.watch(rsp); err != nil {
					w.Stop()
					return
				}
			}
		}
	}()

	return w, nil
}

// watchRouter watches router and send events to all registered watchers
func (w *watcher) watch(stream pb.Router_WatchService) error {
	var watchErr error

	for {
		resp, err := stream.Recv()
		if err != nil {
			if err != io.EOF {
				watchErr = err
			}
			break
		}

		route := router.Route{
			Service: resp.Route.Service,
			Address: resp.Route.Address,
			Gateway: resp.Route.Gateway,
			Network: resp.Route.Network,
			Link:    resp.Route.Link,
			Metric:  resp.Route.Metric,
		}

		event := &router.Event{
			Id:        resp.Id,
			Type:      router.EventType(resp.Type),
			Timestamp: time.Unix(0, resp.Timestamp),
			Route:     route,
		}

		select {
		case w.resChan <- event:
		case <-w.done:
		}
	}

	return watchErr
}

// Next is a blocking call that returns watch result
func (w *watcher) Next() (*router.Event, error) {
	for {
		select {
		case res := <-w.resChan:
			switch w.opts.Service {
			case res.Route.Service, "*":
				return res, nil
			default:
				continue
			}
		case <-w.done:
			return nil, router.ErrWatcherStopped
		}
	}
}

// Chan returns event channel
func (w *watcher) Chan() (<-chan *router.Event, error) {
	return w.resChan, nil
}

// Stop stops watcher
func (w *watcher) Stop() {
	w.Lock()
	defer w.Unlock()

	select {
	case <-w.done:
		return
	default:
		w.stream.Close()
		close(w.done)
	}
}
