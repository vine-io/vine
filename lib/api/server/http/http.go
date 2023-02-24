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

// Package http provides a http server with features; acme, cors, etc
package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"

	"github.com/vine-io/vine/lib/api/server"
	"github.com/vine-io/vine/lib/api/server/cors"
	log "github.com/vine-io/vine/lib/logger"
)

type httpServer struct {
	mux  *http.ServeMux
	opts server.Options

	mtx     sync.RWMutex
	address string
	exit    chan chan error
}

func NewServer(address string, opts ...server.Option) server.Server {
	var options server.Options
	for _, o := range opts {
		o(&options)
	}

	return &httpServer{
		opts:    options,
		mux:     http.NewServeMux(),
		address: address,
		exit:    make(chan chan error),
	}
}

func (s *httpServer) Address() string {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.address
}

func (s *httpServer) Init(opts ...server.Option) error {
	for _, o := range opts {
		o(&s.opts)
	}
	return nil
}

func (s *httpServer) Handle(path string, handler http.Handler) {

	// apply the wrappers, e.g. auth
	for _, wrapper := range s.opts.Wrappers {
		handler = wrapper(handler)
	}

	// wrap with cors
	if s.opts.EnableCORS {
		handler = cors.CombinedCORSHandler(handler, s.opts.CORSConfig)
	}

	s.mux.Handle(path, handler)
}

func (s *httpServer) Start() error {
	var l net.Listener
	var err error

	if s.opts.EnableTLS && s.opts.TLSConfig != nil {
		l, err = tls.Listen("tcp", s.address, s.opts.TLSConfig)
	} else {
		// otherwise, plain listen
		l, err = net.Listen("tcp", s.address)
	}
	if err != nil {
		return err
	}

	log.Infof("HTTP API Listening on %s", l.Addr().String())

	s.mtx.Lock()
	s.address = l.Addr().String()
	s.mtx.Unlock()

	go func() {
		if err = http.Serve(l, s.mux); err != nil {
			// temporary fix
			log.Fatal(err)
		}
	}()

	go func() {
		ch := <-s.exit
		ch <- l.Close()
	}()

	return nil
}

func (s *httpServer) Stop() error {
	ch := make(chan error)
	s.exit <- ch
	return <-ch
}

func (s *httpServer) String() string {
	return "http"
}
