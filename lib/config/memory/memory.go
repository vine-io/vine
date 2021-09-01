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

package memory

import (
	"bytes"
	"sync"
	"time"

	"github.com/vine-io/vine/lib/config"
	"github.com/vine-io/vine/lib/config/loader"
	m "github.com/vine-io/vine/lib/config/loader/memory"
	"github.com/vine-io/vine/lib/config/reader"
	"github.com/vine-io/vine/lib/config/reader/json"
	"github.com/vine-io/vine/lib/config/source"
)

type memory struct {
	exit chan bool
	opts config.Options

	sync.RWMutex
	// the current snapshot
	snap *loader.Snapshot
	// the current values
	vals reader.Values
}

type watcher struct {
	lw    loader.Watcher
	rd    reader.Reader
	path  []string
	value reader.Value
}

func newConfig(opts ...config.Option) config.Config {
	var c memory

	c.Init(opts...)
	go c.run()

	return &c
}

func (c *memory) Init(opts ...config.Option) error {
	c.opts = config.Options{
		Reader: json.NewReader(),
	}
	c.exit = make(chan bool)
	for _, o := range opts {
		o(&c.opts)
	}

	// default loader uses the configured reader
	if c.opts.Loader == nil {
		c.opts.Loader = m.NewLoader(m.WithReader(c.opts.Reader))
	}

	err := c.opts.Loader.Load(c.opts.Source...)
	if err != nil {
		return err
	}

	c.snap, err = c.opts.Loader.Snapshot()
	if err != nil {
		return err
	}

	c.vals, err = c.opts.Reader.Values(c.snap.ChangeSet)
	if err != nil {
		return err
	}

	return nil
}

func (c *memory) Options() config.Options {
	return c.opts
}

func (c *memory) run() {
	watch := func(w loader.Watcher) error {
		for {
			// get changeset
			snap, err := w.Next()
			if err != nil {
				return err
			}

			c.Lock()

			if c.snap.Version >= snap.Version {
				c.Unlock()
				continue
			}

			// save
			c.snap = snap

			// set values
			c.vals, _ = c.opts.Reader.Values(snap.ChangeSet)

			c.Unlock()
		}
	}

	for {
		w, err := c.opts.Loader.Watch()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		done := make(chan bool)

		// the stop watch func
		go func() {
			select {
			case <-done:
			case <-c.exit:
			}
			w.Stop()
		}()

		// block watch
		if err := watch(w); err != nil {
			// to something better
			time.Sleep(time.Second)
		}

		// close done chan
		close(done)

		// if the config closed exit
		select {
		case <-c.exit:
			return
		default:
		}
	}
}

func (c *memory) Map() map[string]interface{} {
	c.RLock()
	defer c.RUnlock()
	return c.vals.Map()
}

func (c *memory) Scan(v interface{}) error {
	c.RLock()
	defer c.RUnlock()
	return c.vals.Scan(v)
}

// Sync loads all the sources, calls the parser and updates the config
func (c *memory) Sync() error {
	if err := c.opts.Loader.Sync(); err != nil {
		return err
	}

	snap, err := c.opts.Loader.Snapshot()
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	c.snap = snap
	vals, err := c.opts.Reader.Values(snap.ChangeSet)
	c.vals = vals

	return err
}

func (c *memory) Close() error {
	select {
	case <-c.exit:
		return nil
	default:
		close(c.exit)
	}
	return nil
}

func (c *memory) Get(path ...string) reader.Value {
	c.RLock()
	defer c.RUnlock()

	// did sync actually work?
	if c.vals != nil {
		return c.vals.Get(path...)
	}

	// no value
	return newValue()
}

func (c *memory) Set(val interface{}, path ...string) {
	c.Lock()
	defer c.Unlock()

	if c.vals != nil {
		c.vals.Set(val, path...)
	}

	return
}

func (c *memory) Del(path ...string) {
	c.Lock()
	defer c.Unlock()

	if c.vals != nil {
		c.vals.Del(path...)
	}

	return
}

func (c *memory) Bytes() []byte {
	c.RLock()
	defer c.RUnlock()

	if c.vals == nil {
		return []byte{}
	}

	return c.vals.Bytes()
}

func (c *memory) Load(sources ...source.Source) error {
	if err := c.opts.Loader.Load(sources...); err != nil {
		return err
	}

	snap, err := c.opts.Loader.Snapshot()
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	c.snap = snap
	vals, err := c.opts.Reader.Values(snap.ChangeSet)
	if err != nil {
		return err
	}
	c.vals = vals

	return nil
}

func (c *memory) Watch(path ...string) (config.Watcher, error) {
	value := c.Get(path...)

	w, err := c.opts.Loader.Watch(path...)
	if err != nil {
		return nil, err
	}

	return &watcher{
		lw:    w,
		rd:    c.opts.Reader,
		path:  path,
		value: value,
	}, nil
}

func (c *memory) String() string {
	return "memory"
}

func (w *watcher) Next() (reader.Value, error) {
	for {
		s, err := w.lw.Next()
		if err != nil {
			return nil, err
		}

		// only process changes
		if bytes.Equal(w.value.Bytes(), s.ChangeSet.Data) {
			continue
		}

		v, err := w.rd.Values(s.ChangeSet)
		if err != nil {
			return nil, err
		}

		w.value = v.Get()
		return w.value, nil
	}
}

func (w *watcher) Stop() error {
	return w.lw.Stop()
}

func NewConfig(opts ...config.Option) config.Config {
	return newConfig(opts...)
}
