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
