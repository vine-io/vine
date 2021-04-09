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

// Package http enables the http profiler
package http

import (
	"context"
	"net/http"
	"net/http/pprof"
	"sync"

	"github.com/lack-io/vine/service/debug/profile"
)

type httpProfile struct {
	sync.Mutex
	running bool
	server  *http.Server
}

var (
	DefaultAddress = ":6060"
)

// Start the profiler
func (h *httpProfile) Start() error {
	h.Lock()
	defer h.Unlock()

	if h.running {
		return nil
	}

	go func() {
		if err := h.server.ListenAndServe(); err != nil {
			h.Lock()
			h.running = false
			h.Unlock()
		}
	}()

	h.running = true

	return nil
}

// Stop the profiler
func (h *httpProfile) Stop() error {
	h.Lock()
	defer h.Unlock()

	if !h.running {
		return nil
	}

	h.running = false

	return h.server.Shutdown(context.TODO())
}

func (h *httpProfile) String() string {
	return "http"
}

func NewProfile(opts ...profile.Option) profile.Profile {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return &httpProfile{
		server: &http.Server{
			Addr:    DefaultAddress,
			Handler: mux,
		},
	}
}
