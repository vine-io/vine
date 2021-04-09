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

// Package pprof provides a pprof profiler
package pprof

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/lack-io/vine/service/debug/profile"
)

type profiler struct {
	opts profile.Options

	sync.Mutex
	running bool
	exit    chan bool

	// where the cpu profile is written
	cpuFile *os.File
	// where the mem profile is written
	memFile *os.File
}

func (p *profiler) writeHeap(f *os.File) {
	defer f.Close()

	t := time.NewTicker(time.Second * 30)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			runtime.GC()
			pprof.WriteHeapProfile(f)
		case <-p.exit:
			return
		}
	}
}

func (p *profiler) Start() error {
	p.Lock()
	defer p.Unlock()

	if p.running {
		return nil
	}

	// create exit channel
	p.exit = make(chan bool)

	cpuFile := filepath.Join("/tmp", "cpu.pprof")
	memFile := filepath.Join("/tmp", "mem.pprof")

	if len(p.opts.Name) > 0 {
		cpuFile = filepath.Join("/tmp", p.opts.Name+".cpu.pprof")
		memFile = filepath.Join("/tmp", p.opts.Name+".mem.pprof")
	}

	f1, err := os.Create(cpuFile)
	if err != nil {
		return err
	}

	f2, err := os.Create(memFile)
	if err != nil {
		return err
	}

	// start cpu profiling
	if err := pprof.StartCPUProfile(f1); err != nil {
		return err
	}

	// write the heap periodically
	go p.writeHeap(f2)

	// set cpu file
	p.cpuFile = f1
	// set mem file
	p.memFile = f2

	p.running = true

	return nil
}

func (p *profiler) Stop() error {
	p.Lock()
	defer p.Unlock()

	select {
	case <-p.exit:
		return nil
	default:
		close(p.exit)
		pprof.StopCPUProfile()
		p.cpuFile.Close()
		p.running = false
		p.cpuFile = nil
		p.memFile = nil
		return nil
	}
}

func (p *profiler) String() string {
	return "pprof"
}

func NewProfile(opts ...profile.Option) profile.Profile {
	var options profile.Options
	for _, o := range opts {
		o(&options)
	}
	p := new(profiler)
	p.opts = options
	return p
}
