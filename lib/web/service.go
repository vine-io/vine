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

package web

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/vine-io/cli"

	svc "github.com/vine-io/vine"
	"github.com/vine-io/vine/core/registry"
	"github.com/vine-io/vine/lib/logger"
	regpb "github.com/vine-io/vine/proto/apis/registry"
	maddr "github.com/vine-io/vine/util/addr"
	"github.com/vine-io/vine/util/backoff"
	mhttp "github.com/vine-io/vine/util/http"
	mnet "github.com/vine-io/vine/util/net"
	signalutil "github.com/vine-io/vine/util/signal"
	mls "github.com/vine-io/vine/util/tls"
)

type service struct {
	opts Options

	app *fiber.App
	svc *regpb.Service

	sync.RWMutex
	running bool
	static  bool
	exit    chan chan error
}

func newService(opts ...Option) Service {
	options := newOptions(opts...)
	s := &service{
		opts:   options,
		app:    fiber.New(fiber.Config{DisableStartupMessage: true}),
		static: true,
	}
	s.svc = s.genSrv()
	return s
}

func (s *service) genSrv() *regpb.Service {
	var host string
	var port string
	var err error

	// default host:port
	if len(s.opts.Address) > 0 {
		host, port, err = net.SplitHostPort(s.opts.Address)
		if err != nil {
			logger.Fatal(err)
		}
	}

	// check the advertise address first
	// if it exists then use it, otherwise
	// use the address
	if len(s.opts.Advertise) > 0 {
		host, port, err = net.SplitHostPort(s.opts.Advertise)
		if err != nil {
			logger.Fatal(err)
		}
	}

	addr, err := maddr.Extract(host)
	if err != nil {
		logger.Fatal(err)
	}

	if strings.Count(addr, ":") > 0 {
		addr = "[" + addr + "]"
	}

	return &regpb.Service{
		Name:    s.opts.Name,
		Version: s.opts.Version,
		Nodes: []*regpb.Node{{
			Id:       s.opts.Id,
			Address:  fmt.Sprintf("%s:%s", addr, port),
			Metadata: s.opts.Metadata,
		}},
	}
}

func (s *service) run(exit chan bool) {
	s.RLock()
	if s.opts.RegisterInterval <= time.Duration(0) {
		s.RUnlock()
		return
	}

	t := time.NewTicker(s.opts.RegisterInterval)
	s.RUnlock()

	for {
		select {
		case <-t.C:
			s.register()
		case <-exit:
			t.Stop()
			return
		}
	}
}

func (s *service) register() error {
	s.Lock()
	defer s.Unlock()

	if s.svc == nil {
		return nil
	}
	// default to service registry
	r := s.opts.Service.Client().Options().Registry
	// switch to option if specified
	if s.opts.Registry != nil {
		r = s.opts.Registry
	}

	// service node need modify, node address maybe changed
	svc := s.genSrv()
	svc.Endpoints = s.svc.Endpoints
	s.svc = svc

	// use RegisterCheck func before register
	if err := s.opts.RegisterCheck(s.opts.Context); err != nil {
		logger.Errorf("Server %s-%s register check error: %s", s.opts.Name, s.opts.Id, err)
		return err
	}

	var regErr error

	// try three times if necessary
	for i := 0; i < 3; i++ {
		// attempt to register
		if err := r.Register(s.svc, registry.RegisterTTL(s.opts.RegisterTTL)); err != nil {
			// set the error
			regErr = err
			// backoff then retry
			time.Sleep(backoff.Do(i + 1))
			continue
		}
		// success so nil error
		regErr = nil
		break
	}

	return regErr
}

func (s *service) deregister() error {
	s.Lock()
	defer s.Unlock()

	if s.svc == nil {
		return nil
	}
	// default to service registry
	r := s.opts.Service.Client().Options().Registry
	// switch to option if specified
	if s.opts.Registry != nil {
		r = s.opts.Registry
	}
	return r.Deregister(s.svc)
}

func (s *service) start() error {
	s.Lock()
	defer s.Unlock()

	if s.running {
		return nil
	}

	for _, fn := range s.opts.BeforeStart {
		if err := fn(); err != nil {
			return err
		}
	}

	l, err := s.listen("tcp", s.opts.Address)
	if err != nil {
		return err
	}

	s.opts.Address = l.Addr().String()
	svc := s.genSrv()
	svc.Endpoints = s.svc.Endpoints
	s.svc = svc

	var app *fiber.App

	if s.opts.App != nil {
		app = s.opts.App
	} else {
		app = s.app
		var r sync.Once

		// register the html dir
		r.Do(func() {
			// static dir
			static := s.opts.StaticDir
			if s.opts.StaticDir[0] != '/' {
				dir, _ := os.Getwd()
				static = filepath.Join(dir, static)
			}

			// set static if no / handler is registered
			if s.static {
				_, err := os.Stat(static)
				if err == nil {
					logger.Infof("Enabling static file serving from %s", static)
					s.app.Use("/", filesystem.New(filesystem.Config{Root: http.Dir(static)}))
				}
			}
		})
	}

	go app.Listener(l)

	for _, fn := range s.opts.AfterStart {
		if err := fn(); err != nil {
			return err
		}
	}

	s.exit = make(chan chan error, 1)
	s.running = true

	go func() {
		ch := <-s.exit
		ch <- l.Close()
	}()

	logger.Infof("Listening on %v", l.Addr().String())
	return nil
}

func (s *service) stop() error {
	s.Lock()
	defer s.Unlock()

	if !s.running {
		return nil
	}

	for _, fn := range s.opts.BeforeStop {
		if err := fn(); err != nil {
			return err
		}
	}

	ch := make(chan error, 1)
	s.exit <- ch
	s.running = false

	logger.Info("Stopping")

	for _, fn := range s.opts.AfterStop {
		if err := fn(); err != nil {
			if chErr := <-ch; chErr != nil {
				return chErr
			}
			return err
		}
	}

	return <-ch
}

func (s *service) Client() *http.Client {
	rt := mhttp.NewRoundTripper(
		mhttp.WithRegistry(s.opts.Registry),
	)
	return &http.Client{
		Transport: rt,
	}
}

func (s *service) Handle(method, pattern string, handler fiber.Handler) {
	var seen bool
	s.RLock()
	for _, ep := range s.svc.Endpoints {
		if ep.Name == pattern {
			seen = true
			break
		}
	}
	s.RUnlock()

	// if its unseen then add an endpoint
	if !seen {
		s.Lock()
		s.svc.Endpoints = append(s.svc.Endpoints, &regpb.Endpoint{
			Name:     pattern,
			Metadata: map[string]string{"method": method},
		})
		s.Unlock()
	}

	// disable static serving
	if pattern == "/" {
		s.Lock()
		s.static = false
		s.Unlock()
	}

	// register the handler
	switch method {
	case MethodHead:
		s.app.Head(pattern, handler)
	case MethodGet:
		s.app.Get(pattern, handler)
	case MethodPut:
		s.app.Put(pattern, handler)
	case MethodPatch:
		s.app.Patch(pattern, handler)
	case MethodPost:
		s.app.Post(pattern, handler)
	case MethodDelete:
		s.app.Delete(pattern, handler)
	case MethodConnect:
		s.app.Connect(pattern, handler)
	case MethodOptions:
		s.app.Options(pattern, handler)
	case MethodTrace:
		s.app.Trace(pattern, handler)
	case MethodAny:
		s.app.All(pattern, handler)
	default:
		s.app.All(pattern, handler)
	}
}

func (s *service) Init(opts ...Option) error {
	s.Lock()

	for _, o := range opts {
		o(&s.opts)
	}

	serviceOpts := []svc.Option{}

	if len(s.opts.Flags) > 0 {
		serviceOpts = append(serviceOpts, svc.Flags(s.opts.Flags...))
	}

	if s.opts.Registry != nil {
		serviceOpts = append(serviceOpts, svc.Registry(s.opts.Registry))
	}

	s.Unlock()

	serviceOpts = append(serviceOpts, svc.Action(func(ctx *cli.Context) error {
		s.Lock()
		defer s.Unlock()

		if ttl := ctx.Int("register-ttl"); ttl > 0 {
			s.opts.RegisterTTL = time.Duration(ttl) * time.Second
		}

		if interval := ctx.Int("register-interval"); interval > 0 {
			s.opts.RegisterInterval = time.Duration(interval) * time.Second
		}

		if name := ctx.String("server-name"); len(name) > 0 {
			s.opts.Name = name
		}

		if ver := ctx.String("server-version"); len(ver) > 0 {
			s.opts.Version = ver
		}

		if id := ctx.String("server-id"); len(id) > 0 {
			s.opts.Id = id
		}

		if addr := ctx.String("server-address"); len(addr) > 0 {
			s.opts.Address = addr
		}

		if adv := ctx.String("server-advertise"); len(adv) > 0 {
			s.opts.Advertise = adv
		}

		if s.opts.Action != nil {
			s.opts.Action(ctx)
		}

		return nil
	}))

	s.RLock()
	// pass in own name and version
	if s.opts.Service.Name() == "" {
		serviceOpts = append(serviceOpts, svc.Name(s.opts.Name))
	}
	serviceOpts = append(serviceOpts, svc.Version(s.opts.Version))
	s.RUnlock()

	s.opts.Service.Init(serviceOpts...)

	s.Lock()
	svc := s.genSrv()
	svc.Endpoints = s.svc.Endpoints
	s.svc = svc
	s.Unlock()

	return nil
}

func (s *service) Run() error {
	// generate an auth account
	if err := s.start(); err != nil {
		return err
	}

	if err := s.register(); err != nil {
		return err
	}

	// start reg loop
	ex := make(chan bool)
	go s.run(ex)

	ch := make(chan os.Signal, 1)
	if s.opts.Signal {
		signal.Notify(ch, signalutil.Shutdown()...)
	}

	select {
	// wait on kill signal
	case sig := <-ch:
		logger.Infof("Received signal %s", sig)
	// wait on context cancel
	case <-s.opts.Context.Done():
		logger.Info("Received context shutdown")
	}

	// exit reg loop
	close(ex)

	if err := s.deregister(); err != nil {
		return err
	}

	return s.stop()
}

// Options returns the options for the given service
func (s *service) Options() Options {
	return s.opts
}

func (s *service) listen(network, addr string) (net.Listener, error) {
	var l net.Listener
	var err error

	// TODO: support use of listen options
	if s.opts.Secure || s.opts.TLSConfig != nil {
		config := s.opts.TLSConfig

		fn := func(addr string) (net.Listener, error) {
			if config == nil {
				hosts := []string{addr}

				// check if its a valid host:port
				if host, _, err := net.SplitHostPort(addr); err == nil {
					if len(host) == 0 {
						hosts = maddr.IPs()
					} else {
						hosts = []string{host}
					}
				}

				// generate a certificate
				cert, err := mls.Certificate(hosts...)
				if err != nil {
					return nil, err
				}
				config = &tls.Config{Certificates: []tls.Certificate{cert}}
			}
			return tls.Listen(network, addr, config)
		}

		l, err = mnet.Listen(addr, fn)
	} else {
		fn := func(addr string) (net.Listener, error) {
			return net.Listen(network, addr)
		}

		l, err = mnet.Listen(addr, fn)
	}

	if err != nil {
		return nil, err
	}

	return l, nil
}
