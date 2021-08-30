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
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vine-io/vine/core/registry"
	"github.com/vine-io/vine/core/registry/memory"
	regpb "github.com/vine-io/vine/proto/apis/registry"
)

func TestService(t *testing.T) {
	var (
		beforeStartCalled bool
		afterStartCalled  bool
		beforeStopCalled  bool
		afterStopCalled   bool
		str               = `<html><body><h1>Hello World</h1></body></html>`
		fn                = func(c *fiber.Ctx) error { return c.SendString(str) }
		reg               = memory.NewRegistry()
	)

	beforeStart := func() error {
		beforeStartCalled = true
		return nil
	}

	afterStart := func() error {
		afterStartCalled = true
		return nil
	}

	beforeStop := func() error {
		beforeStopCalled = true
		return nil
	}

	afterStop := func() error {
		afterStopCalled = true
		return nil
	}

	service := NewService(
		Name("go.vine.web.test"),
		Registry(reg),
		BeforeStart(beforeStart),
		AfterStart(afterStart),
		BeforeStop(beforeStop),
		AfterStop(afterStop),
	)

	service.Handle(MethodGet, "/", fn)

	errCh := make(chan error, 1)
	go func() {
		errCh <- service.Run()
		close(errCh)
	}()

	var s []*registry.Service

	eventually(func() bool {
		var err error
		s, err = reg.GetService("go.vine.web.test")
		return err == nil
	}, t.Fatal)

	if have, want := len(s), 1; have != want {
		t.Fatalf("Expected %d but got %d services", want, have)
	}

	rsp, err := http.Get(fmt.Sprintf("http://%s", s[0].Nodes[0].Address))
	if err != nil {
		t.Fatal(err)
	}
	defer rsp.Body.Close()

	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != str {
		t.Errorf("Expected %s got %s", str, string(b))
	}

	callbackTests := []struct {
		subject string
		have    interface{}
	}{
		{"beforeStartCalled", beforeStartCalled},
		{"afterStartCalled", afterStartCalled},
	}

	for _, tt := range callbackTests {
		if tt.have != true {
			t.Errorf("unexpected %s: want true, have false", tt.subject)
		}
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("service.Run():%v", err)
		}
	case <-time.After(time.Duration(time.Second)):
		if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
			t.Logf("service.Run() survived a client request without an error")
		}
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)
	p, _ := os.FindProcess(os.Getpid())
	p.Signal(syscall.SIGTERM)

	<-ch

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("service.Run():%v", err)
		} else {
			if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
				t.Log("service.Run() nil return on syscall.SIGTERM")
			}
		}
	case <-time.After(time.Duration(time.Second)):
		if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
			t.Logf("service.Run() survived a client request without an error")
		}
	}

	eventually(func() bool {
		_, err := reg.GetService("go.vine.web.test")
		return err == registry.ErrNotFound
	}, t.Error)

	callbackTests = []struct {
		subject string
		have    interface{}
	}{
		{"beforeStopCalled", beforeStopCalled},
		{"afterStopCalled", afterStopCalled},
	}

	for _, tt := range callbackTests {
		if tt.have != true {
			t.Errorf("unexpected %s: want true, have false", tt.subject)
		}
	}

}

func TestOptions(t *testing.T) {
	var (
		name             = "service-name"
		id               = "service-id"
		version          = "service-version"
		address          = "service-addr:8080"
		advertise        = "service-adv:8080"
		reg              = memory.NewRegistry()
		registerTTL      = 123 * time.Second
		registerInterval = 456 * time.Second
		app              = fiber.New()
		metadata         = map[string]string{"key": "val"}
		secure           = true
	)

	service := NewService(
		Name(name),
		Id(id),
		Version(version),
		Address(address),
		Advertise(advertise),
		Registry(reg),
		RegisterTTL(registerTTL),
		RegisterInterval(registerInterval),
		App(app),
		Metadata(metadata),
		Secure(secure),
	)

	opts := service.Options()

	tests := []struct {
		subject string
		want    interface{}
		have    interface{}
	}{
		{"name", name, opts.Name},
		{"version", version, opts.Version},
		{"id", id, opts.Id},
		{"address", address, opts.Address},
		{"advertise", advertise, opts.Advertise},
		{"registry", reg, opts.Registry},
		{"registerTTL", registerTTL, opts.RegisterTTL},
		{"registerInterval", registerInterval, opts.RegisterInterval},
		{"handler", app, opts.App},
		{"metadata", metadata["key"], opts.Metadata["key"]},
		{"secure", secure, opts.Secure},
	}

	for _, tc := range tests {
		if tc.want != tc.have {
			t.Errorf("unexpected %s: want %v, have %v", tc.subject, tc.want, tc.have)
		}
	}
}

func eventually(pass func() bool, fail func(...interface{})) {
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()

	timeout := time.After(time.Second)

	for {
		select {
		case <-timeout:
			fail("timed out")
			return
		case <-tick.C:
			if pass() {
				return
			}
		}
	}
}

func TestTLS(t *testing.T) {
	var (
		str    = `<html><body><h1>Hello World</h1></body></html>`
		fn     = func(c *fiber.Ctx) error { return c.SendString(str) }
		secure = true
		reg    = memory.NewRegistry()
	)

	service := NewService(
		Name("go.vine.web.test"),
		Secure(secure),
		Registry(reg),
	)

	service.Handle(MethodGet, "/", fn)

	errCh := make(chan error, 1)
	go func() {
		errCh <- service.Run()
		close(errCh)
	}()

	var s []*registry.Service

	eventually(func() bool {
		var err error
		s, err = reg.GetService("go.vine.web.test")
		return err == nil
	}, t.Fatal)

	if have, want := len(s), 1; have != want {
		t.Fatalf("Expected %d but got %d services", want, have)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	rsp, err := client.Get(fmt.Sprintf("https://%s", s[0].Nodes[0].Address))
	if err != nil {
		t.Fatal(err)
	}
	defer rsp.Body.Close()

	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != str {
		t.Errorf("Expected %s got %s", str, string(b))
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("service.Run():%v", err)
		}
	case <-time.After(time.Duration(time.Second)):
		if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
			t.Logf("service.Run() survived a client request without an error")
		}
	}

}
