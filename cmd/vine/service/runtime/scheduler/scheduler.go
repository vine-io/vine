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

// Package scheduler is a file server notifer
package scheduler

import (
	"errors"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/runtime"
)

type scheduler struct {
	service string
	version string
	path    string

	once sync.Once
	sync.Mutex

	fsnotify *fsnotify.Watcher
	notify   chan runtime.Event
	update   chan fsnotify.Event
	exit     chan bool
}

func (n *scheduler) run() {
	for {
		select {
		case <-n.exit:
			return
		case <-n.update:
			select {
			case n.notify <- runtime.Event{
				Type:      runtime.Update,
				Timestamp: time.Now(),
				Service:   &runtime.Service{Name: n.service},
			}:
			default:
				// bail out
			}
		case ev := <-n.fsnotify.Events:
			select {
			case n.update <- ev:
			default:
				// bail out
			}
		case <-n.fsnotify.Errors:
			// replace the watcher
			n.fsnotify, _ = fsnotify.NewWatcher()
			n.fsnotify.Add(n.path)
		}
	}
}

func (n *scheduler) Notify() (<-chan runtime.Event, error) {
	select {
	case <-n.exit:
		return nil, errors.New("closed")
	default:
	}

	n.once.Do(func() {
		go n.run()
	})

	return n.notify, nil
}

func (n *scheduler) Close() error {
	n.Lock()
	defer n.Unlock()
	select {
	case <-n.exit:
		return nil
	default:
		close(n.exit)
		n.fsnotify.Close()
		return nil
	}
}

// New returns a new scheduler which watches the source
func New(service, version, source string) runtime.Scheduler {
	n := &scheduler{
		path:    filepath.Dir(source),
		exit:    make(chan bool),
		notify:  make(chan runtime.Event, 32),
		update:  make(chan fsnotify.Event, 32),
		service: service,
		version: version,
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	w.Add(n.path)
	// set the watcher
	n.fsnotify = w

	return n
}
