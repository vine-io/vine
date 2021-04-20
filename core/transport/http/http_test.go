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
	"sync"
	"testing"

	"github.com/lack-io/vine/core/transport"
)

func call(b *testing.B, c int) {
	b.StopTimer()

	tr := NewTransport()

	// server listen
	l, err := tr.Listen("localhost:0")
	if err != nil {
		b.Fatal(err)
	}
	defer l.Close()

	// socket func
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

	// accept connections
	go func() {
		if err := l.Accept(fn); err != nil {
			select {
			case <-done:
			default:
				b.Fatalf("Unexpected accept err: %v", err)
			}
		}
	}()

	m := transport.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "Hello World"}`),
	}

	// client connection
	client, err := tr.Dial(l.Addr())
	if err != nil {
		b.Fatalf("Unexpected dial err: %v", err)
	}

	send := func(c transport.Client) {
		// send message
		if err := c.Send(&m); err != nil {
			b.Fatalf("Unexpected send err: %v", err)
		}

		var rm transport.Message
		// receive message
		if err := c.Recv(&rm); err != nil {
			b.Fatalf("Unexpected recv err: %v", err)
		}
	}

	// warm
	for i := 0; i < 10; i++ {
		send(client)
	}

	client.Close()

	ch := make(chan int, c*4)

	var wg sync.WaitGroup
	wg.Add(c)

	for i := 0; i < c; i++ {
		go func() {
			cl, err := tr.Dial(l.Addr())
			if err != nil {
				b.Fatalf("Unexpected dial err: %v", err)
			}
			defer cl.Close()

			for range ch {
				send(cl)
			}

			wg.Done()
		}()
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		ch <- i
	}

	b.StopTimer()
	close(ch)

	wg.Wait()

	// finish
	close(done)
}

func BenchmarkTransport1(b *testing.B) {
	call(b, 1)
}

func BenchmarkTransport8(b *testing.B) {
	call(b, 8)
}

func BenchmarkTransport16(b *testing.B) {
	call(b, 16)
}

func BenchmarkTransport64(b *testing.B) {
	call(b, 64)
}

func BenchmarkTransport128(b *testing.B) {
	call(b, 128)
}
