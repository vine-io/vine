// Copyright 2020 lack
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

package etcd

import (
	"context"
	"errors"
	"time"

	"github.com/coreos/etcd/clientv3"

	"github.com/lack-io/vine/service/registry"
)

type etcdWatcher struct {
	stop    chan bool
	w       clientv3.WatchChan
	client  *clientv3.Client
	timeout time.Duration
}

func newEtcdWatcher(r *etcdRegistry, timeout time.Duration, opts ...registry.WatchOption) (registry.Watcher, error) {
	var wo registry.WatchOptions
	for _, o := range opts {
		o(&wo)
	}

	ctx, cancel := context.WithCancel(context.Background())
	stop := make(chan bool, 1)

	go func() {
		<-stop
		cancel()
	}()

	watchPath := prefix
	if len(wo.Service) > 0 {
		watchPath = servicePath(wo.Service) + "/"
	}

	return &etcdWatcher{
		stop:    stop,
		w:       r.client.Watch(ctx, watchPath, clientv3.WithPrefix(), clientv3.WithPrevKV()),
		client:  r.client,
		timeout: timeout,
	}, nil
}

func (ew *etcdWatcher) Next() (*registry.Result, error) {
	for wresp := range ew.w {
		if wresp.Err() != nil {
			return nil, wresp.Err()
		}
		if wresp.Canceled {
			return nil, errors.New("could not get next, watch is canceled")
		}
		for _, ev := range wresp.Events {
			service := decode(ev.Kv.Value)
			var action string

			switch ev.Type {
			case clientv3.EventTypePut:
				if ev.IsCreate() {
					action = "create"
				} else if ev.IsModify() {
					action = "update"
				}
			case clientv3.EventTypeDelete:
				action = "delete"

				// get service from prevKv
				service = decode(ev.PrevKv.Value)
			}

			if service == nil {
				continue
			}
			return &registry.Result{
				Action:  action,
				Service: service,
			}, nil
		}
	}

	return nil, errors.New("could not get next")
}

func (ew *etcdWatcher) Stop() {
	select {
	case <-ew.stop:
		return
	default:
		close(ew.stop)
	}
}
