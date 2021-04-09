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

package stats

import (
	"runtime"
	"sync"
	"time"

	"github.com/lack-io/vine/util/ring"
)

type stats struct {
	// used to store past stats
	buffer *ring.Buffer

	sync.RWMutex
	started  int64
	requests uint64
	errors   uint64
}

func (s *stats) snapshot() *Stat {
	s.RLock()
	defer s.RUnlock()

	var mstat runtime.MemStats
	runtime.ReadMemStats(&mstat)

	now := time.Now().Unix()

	return &Stat{
		Timestamp: now,
		Started:   s.started,
		Uptime:    now - s.started,
		Memory:    mstat.Alloc,
		GC:        mstat.PauseTotalNs,
		Threads:   uint64(runtime.NumGoroutine()),
		Requests:  s.requests,
		Errors:    s.errors,
	}
}

func (s *stats) Read() ([]*Stat, error) {
	// TODO adjustable size and optional read values
	buf := s.buffer.Get(60)
	var stats []*Stat

	// get a value from the buffer if it exists
	for _, b := range buf {
		stat, ok := b.Value.(*Stat)
		if !ok {
			continue
		}
		stats = append(stats, stat)
	}

	// get a snapshot
	stats = append(stats, s.snapshot())

	return stats, nil
}

func (s *stats) Write(stat *Stat) error {
	s.buffer.Put(stat)
	return nil
}

func (s *stats) Record(err error) error {
	s.Lock()
	defer s.Unlock()

	// increment the total request count
	s.requests++

	// increment the error count
	if err != nil {
		s.errors++
	}

	return nil
}

// NewStats returns a new in memory stats buffer
// TODO add options
func NewStats() Stats {
	return &stats{
		started: time.Now().Unix(),
		buffer:  ring.New(60),
	}
}
