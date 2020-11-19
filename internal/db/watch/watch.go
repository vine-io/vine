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

package watch

import "github.com/lack-io/vine/internal/runtime"

// Watch can be implemented by anything that knows how to watch and report changes.
type Watch interface {
	// Stops watching. Will close the channel returned ResultChan(). Releases
	// any resources used by the watch.
	Stop()

	// Returns a chan which will receive all the events. If an error occurs
	// or Stop() is called, this channel will be closed, in which case the
	// watch should be completely cleaned up.
	ResultChan() <-chan Event
}

// EventType defines the possible types of events.
type EventType string

const (
	Added    EventType = "ADDED"
	Modified EventType = "MODIFIED"
	Deleted  EventType = "DELETE"
	Bookmark EventType = "BOOKMARK"
	Error    EventType = "ERROR"

	DefaultChanSize int32 = 100
)

// Event represents a single event to a watched resource.
type Event struct {
	Type EventType

	// Object is:
	//	* If Type is Added or Modified: the new state of the object.
	//	* If Type is Deleted: the state of the object immediately before deletion.
	//	* If Type is Bookmark: the object (instance of a type being watched) where
	// 	  only ResourceVersion field is set. On successful restart of watch from a
	//	  bookmark resource, client is guaranteed to not get repeat event
	//	  nor miss any events.
	//	* If Type is Error: *metav1.WatchStatus
	Object runtime.Object
}

type emptyWatch chan Event

// NewEmptyWatch returns a watch interface that returns no results and is closed.
// May be used in certain error conditions where no information is available but
// an error is not warranted.
func NewEmptyWatch() Watch {
	ch := make(chan Event)
	close(ch)
	return emptyWatch(ch)
}

// Stop implements Watch
func (w emptyWatch) Stop() {}

// ResultChan implements Interface
func (w emptyWatch) ResultChan() <-chan Event {
	return chan Event(w)
}