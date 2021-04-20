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

package pool

import (
	"testing"
	"time"

	"github.com/lack-io/vine/core/transport"
	"github.com/lack-io/vine/core/transport/memory"
)

func testPool(t *testing.T, size int, ttl time.Duration) {
	// mock transport
	tr := memory.NewTransport()

	options := Options{
		TTL:       ttl,
		Size:      size,
		Transport: tr,
	}
	// zero pool
	p := newPool(options)

	// listen
	l, err := tr.Listen(":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// accept loop
	go func() {
		for {
			if err := l.Accept(func(s transport.Socket) {
				for {
					var msg transport.Message
					if err := s.Recv(&msg); err != nil {
						return
					}
					if err := s.Send(&msg); err != nil {
						return
					}
				}
			}); err != nil {
				return
			}
		}
	}()

	for i := 0; i < 10; i++ {
		// get a conn
		c, err := p.Get(l.Addr())
		if err != nil {
			t.Fatal(err)
		}

		msg := &transport.Message{
			Body: []byte(`hello world`),
		}

		if err := c.Send(msg); err != nil {
			t.Fatal(err)
		}

		var rcv transport.Message

		if err := c.Recv(&rcv); err != nil {
			t.Fatal(err)
		}

		if string(rcv.Body) != string(msg.Body) {
			t.Fatalf("got %v, expected %v", rcv.Body, msg.Body)
		}

		// release the conn
		p.Release(c, nil)

		p.Lock()
		if i := len(p.conns[l.Addr()]); i > size {
			p.Unlock()
			t.Fatalf("pool size %d is greater than expected %d", i, size)
		}
		p.Unlock()
	}
}

func TestClientPool(t *testing.T) {
	testPool(t, 0, time.Minute)
	testPool(t, 2, time.Minute)
}
