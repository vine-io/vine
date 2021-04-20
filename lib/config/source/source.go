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

package source

import (
	"errors"
	"time"
)

var (
	// ErrWatcherStopped is returned when source watcher has been stopped
	ErrWatcherStopped = errors.New("watcher stopped")
)

// Source is the source from which config is loaded
type Source interface {
	Read() (*ChangeSet, error)
	Write(*ChangeSet) error
	Watch() (Watcher, error)
	String() string
}

// ChangeSet represents a set of changes from a source
type ChangeSet struct {
	Data      []byte
	CheckSum  string
	Format    string
	Source    string
	Timestamp time.Time
}

// Watcher watches a source for changes
type Watcher interface {
	Next() (*ChangeSet, error)
	Stop() error
}
