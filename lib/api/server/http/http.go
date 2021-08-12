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
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/lack-io/vine/lib/api/server"
	log "github.com/lack-io/vine/lib/logger"
)

var DefaultBodyLimit = 1024 * 1024 * 1024 * 1024 * 1024

type httpServer struct {
	app  *fiber.App
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
		app:     fiber.New(fiber.Config{BodyLimit: DefaultBodyLimit, DisableStartupMessage: true}),
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

func (s *httpServer) Handle(path string, app *fiber.App) {

	// apply the wrappers, e.g. auth
	for _, wrapper := range s.opts.Wrappers {
		app.Use(wrapper())
	}

	// wrap with cors
	if s.opts.EnableCORS {
		//app.Use()
		app.Use(cors.New())
	}

	// wrap with logger
	//handler = loggingHandler(handler)

	s.app.Use(logger.New(logger.Config{
		Format:       "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat:   "15:04:05",
		TimeZone:     "Local",
		TimeInterval: 0,
		Output:       log.DefaultLogger.Options().Out,
	}))
	s.app.Mount(path, app)
}

func (s *httpServer) Start() error {
	var l net.Listener
	var err error

	if s.opts.EnableTLS && s.opts.TLSConfig != nil {
		l, err = tls.Listen("tcp", s.address, s.opts.TLSConfig)
	} else {
		// otherwise plain listen
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
		if err = s.app.Listener(l); err != nil {
			// temporary fix
			//logger.Fatal(err)
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
