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

package http

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/lack-io/vine/core/transport"
)

func expectedPort(t *testing.T, expected string, lsn transport.Listener) {
	_, port, err := net.SplitHostPort(lsn.Addr())
	if err != nil {
		t.Errorf("Expected address to be `%s`, got error: %v", expected, err)
	}

	if port != expected {
		lsn.Close()
		t.Errorf("Expected address to be `%s`, got `%s`", expected, port)
	}
}

func TestHTTPTransportPortRange(t *testing.T) {
	tp := newHTTPTransport()

	lsn1, err := tp.Listen(":44444-44448")
	if err != nil {
		t.Errorf("Did not expect an error, got %v", err)
	}
	expectedPort(t, "44444", lsn1)

	lsn2, err := tp.Listen(":44444-44448")
	if err != nil {
		t.Errorf("Did not expect an error, got %v", err)
	}

	lsn, err := tp.Listen("127.0.0.1:0")
	if err != nil {
		t.Errorf("Did not expect an error, got %v", err)
	}

	lsn.Close()
	lsn1.Close()
	lsn2.Close()
}

func TestHTTPTransportCommunication(t *testing.T) {
	tr := newHTTPTransport()

	l, err := tr.Listen("127.0.0.1:0")
	if err != nil {
		t.Errorf("Unexpected listen error: %v", err)
	}
	defer l.Close()

	fn := func(sock transport.Socket) {
		defer sock.Close()

		for {
			var m transport.Message
			if err := sock.Recv(&m); err != nil {
				return
			}

			if err := sock.Send(&m); err != nil {
				return
			}
		}
	}

	done := make(chan bool)

	go func() {
		if err := l.Accept(fn); err != nil {
			select {
			case <-done:
			default:
				t.Fatalf("Unexpected accept err: %v", err)
			}
		}
	}()

	c, err := tr.Dial(l.Addr())
	if err != nil {
		t.Errorf("Unexpected dial err: %v", err)
	}
	defer c.Close()

	m := transport.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "Hello World"}`),
	}

	if err := c.Send(&m); err != nil {
		t.Errorf("Unexpected send err: %v", err)
	}

	var rm transport.Message

	if err := c.Recv(&rm); err != nil {
		t.Errorf("Unexpect recv err: %v", err)
	}

	if string(rm.Body) != string(m.Body) {
		t.Errorf("Expected %v, got %v", m.Body, rm.Body)
	}

	close(done)
}

func TestHTTPTransportError(t *testing.T) {
	tr := newHTTPTransport()

	l, err := tr.Listen("127.0.0.1:0")
	if err != nil {
		t.Errorf("Unexpected listen err: %v", err)
	}
	defer l.Close()

	fn := func(sock transport.Socket) {
		defer sock.Close()

		for {
			var m transport.Message
			if err := sock.Recv(&m); err != nil {
				if err == io.EOF {
					return
				}
				t.Fatal(err)
			}

			sock.(*httpTransportSocket).error(&transport.Message{
				Body: []byte(`an error occurred`),
			})
		}
	}

	done := make(chan bool)

	go func() {
		if err := l.Accept(fn); err != nil {
			select {
			case <-done:
			default:
				t.Errorf("Unexpected accept err: %v", err)
			}
		}
	}()

	c, err := tr.Dial(l.Addr())
	if err != nil {
		t.Errorf("Unexpected dial err: %v", err)
	}
	defer c.Close()

	m := transport.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "Hello World"}`),
	}

	if err := c.Send(&m); err != nil {
		t.Errorf("Unexpected send err: %v", err)
	}

	var rm transport.Message

	err = c.Recv(&rm)
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	if err.Error() != "500 Internal Server Error: an error occurred" {
		t.Fatalf("Did not receive expected error, got: %v", err)
	}

	close(done)
}

func TestHTTPTransportTimeout(t *testing.T) {
	tr := newHTTPTransport(transport.Timeout(time.Millisecond * 100))

	l, err := tr.Listen("127.0.0.1:0")
	if err != nil {
		t.Errorf("Unexpected listen err: %v", err)
	}
	defer l.Close()

	done := make(chan bool)

	fn := func(sock transport.Socket) {
		defer func() {
			sock.Close()
			close(done)
		}()

		go func() {
			select {
			case <-done:
				return
			case <-time.After(time.Second):
				t.Fatal("deadline not executed")

			}
		}()

		for {
			var m transport.Message

			if err := sock.Recv(&m); err != nil {
				return
			}
		}
	}

	go func() {
		if err := l.Accept(fn); err != nil {
			select {
			case <-done:
			default:
				t.Errorf("Unexpected accept err: %v", err)
			}
		}
	}()

	c, err := tr.Dial(l.Addr())
	if err != nil {
		t.Errorf("Unexpected dial err: %v", err)
	}
	defer c.Close()

	m := transport.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "Hello World"}`),
	}

	if err := c.Send(&m); err != nil {
		t.Errorf("Unexpected send err: %v", err)
	}

	<-done
}
