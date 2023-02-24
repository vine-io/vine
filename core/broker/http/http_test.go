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
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/vine-io/vine/core/broker"
	"github.com/vine-io/vine/core/registry"
	"github.com/vine-io/vine/core/registry/memory"
)

var (
	// mock data
	testData = map[string][]*registry.Service{
		"foo": {
			{
				Name:    "foo",
				Version: "1.0.0",
				Nodes: []*registry.Node{
					{
						Id:      "foo-1.0.0-123",
						Address: "localhost:9999",
					},
					{
						Id:      "foo-1.0.0-321",
						Address: "localhost:9999",
					},
				},
			},
			{
				Name:    "foo",
				Version: "1.0.1",
				Nodes: []*registry.Node{
					{
						Id:      "foo-1.0.1-321",
						Address: "localhost:6666",
					},
				},
			},
			{
				Name:    "foo",
				Version: "1.0.3",
				Nodes: []*registry.Node{
					{
						Id:      "foo-1.0.3-345",
						Address: "localhost:8888",
					},
				},
			},
		},
	}
)

func newTestRegistry() registry.Registry {
	return memory.NewRegistry(memory.Services(testData))
}

func sub(be *testing.B, c int) {
	be.StopTimer()
	m := newTestRegistry()

	b := NewBroker(broker.Registry(m))
	topic := uuid.New().String()

	if err := b.Init(); err != nil {
		be.Fatalf("Unexpected init error: %v", err)
	}

	if err := b.Connect(); err != nil {
		be.Fatalf("Unexpected connect error: %v", err)
	}

	msg := &broker.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "Hello World"}`),
	}

	ctx := context.TODO()
	var subs []broker.Subscriber
	done := make(chan bool, c)

	for i := 0; i < c; i++ {
		sub, err := b.Subscribe(topic, func(p broker.Event) error {
			done <- true
			m := p.Message()

			if string(m.Body) != string(msg.Body) {
				be.Fatalf("Unexpected msg %s, expected %s", string(m.Body), string(msg.Body))
			}

			return nil
		}, broker.Queue("shared"))
		if err != nil {
			be.Fatalf("Unexpected subscribe error: %v", err)
		}
		subs = append(subs, sub)
	}

	for i := 0; i < be.N; i++ {
		be.StartTimer()
		if err := b.Publish(ctx, topic, msg); err != nil {
			be.Fatalf("Unexpected publish error: %v", err)
		}
		<-done
		be.StopTimer()
	}

	for _, sub := range subs {
		sub.Unsubscribe()
	}

	if err := b.Disconnect(); err != nil {
		be.Fatalf("Unexpected disconnect error: %v", err)
	}
}

func pub(be *testing.B, c int) {
	be.StopTimer()
	m := newTestRegistry()
	b := NewBroker(broker.Registry(m))
	topic := uuid.New().String()

	if err := b.Init(); err != nil {
		be.Fatalf("Unexpected init error: %v", err)
	}

	if err := b.Connect(); err != nil {
		be.Fatalf("Unexpected connect error: %v", err)
	}

	msg := &broker.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "Hello World"}`),
	}

	ctx := context.TODO()
	done := make(chan bool, c*4)

	bsub, err := b.Subscribe(topic, func(p broker.Event) error {
		done <- true
		m := p.Message()
		if string(m.Body) != string(msg.Body) {
			be.Fatalf("Unexpected msg %s, expected %s", string(m.Body), string(msg.Body))
		}
		return nil
	}, broker.Queue("shared"))
	if err != nil {
		be.Fatalf("Unexpected subscribe error: %v", err)
	}

	var wg sync.WaitGroup
	ch := make(chan int, c*4)
	be.StartTimer()

	for i := 0; i < c; i++ {
		go func() {
			for range ch {
				if err := b.Publish(ctx, topic, msg); err != nil {
					be.Fatalf("Unexpected publish error: %v", err)
				}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				wg.Done()
			}
		}()
	}

	for i := 0; i < be.N; i++ {
		wg.Add(1)
		ch <- i
	}

	wg.Wait()
	be.StopTimer()
	bsub.Unsubscribe()
	close(ch)
	close(done)

	if err := b.Disconnect(); err != nil {
		be.Fatalf("Unexpected disconnect error: %v", err)
	}
}

func TestBroker(t *testing.T) {
	m := newTestRegistry()
	b := NewBroker(broker.Registry(m))

	if err := b.Init(); err != nil {
		t.Fatalf("Unexpected init error: %v", err)
	}

	if err := b.Connect(); err != nil {
		t.Fatalf("Unexpected connect error: %v", err)
	}

	msg := &broker.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "Hello World"}`),
	}

	ctx := context.TODO()
	done := make(chan bool)

	sub, err := b.Subscribe("test", func(p broker.Event) error {
		m := p.Message()

		if string(m.Body) != string(msg.Body) {
			t.Fatalf("Unexpected msg %s, expected %s", string(m.Body), string(msg.Body))
		}

		close(done)
		return nil
	})
	if err != nil {
		t.Fatalf("Unexpected subscribe error: %v", err)
	}

	if err := b.Publish(ctx, "test", msg); err != nil {
		t.Fatalf("Unexpected publish error: %v", err)
	}

	<-done
	sub.Unsubscribe()

	if err := b.Disconnect(); err != nil {
		t.Fatalf("Unexpected disconnect error: %v", err)
	}
}

func TestConcurrentSubBroker(t *testing.T) {
	m := newTestRegistry()
	b := NewBroker(broker.Registry(m))

	if err := b.Init(); err != nil {
		t.Fatalf("Unexpected init error: %v", err)
	}

	if err := b.Connect(); err != nil {
		t.Fatalf("Unexpected connect error: %v", err)
	}

	msg := &broker.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "Hello World"}`),
	}

	ctx := context.TODO()
	var subs []broker.Subscriber
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		sub, err := b.Subscribe("test", func(p broker.Event) error {
			defer wg.Done()

			m := p.Message()

			if string(m.Body) != string(msg.Body) {
				t.Fatalf("Unexpected msg %s, expected %s", string(m.Body), string(msg.Body))
			}

			return nil
		})
		if err != nil {
			t.Fatalf("Unexpected subscribe error: %v", err)
		}

		wg.Add(1)
		subs = append(subs, sub)
	}

	if err := b.Publish(ctx, "test", msg); err != nil {
		t.Fatalf("Unexpected publish error: %v", err)
	}

	wg.Wait()

	for _, sub := range subs {
		sub.Unsubscribe()
	}

	if err := b.Disconnect(); err != nil {
		t.Fatalf("Unexpected disconnect error: %v", err)
	}
}

func TestConcurrentPubBroker(t *testing.T) {
	m := newTestRegistry()
	b := NewBroker(broker.Registry(m))

	if err := b.Init(); err != nil {
		t.Fatalf("Unexpected init error: %v", err)
	}

	if err := b.Connect(); err != nil {
		t.Fatalf("Unexpected connect error: %v", err)
	}

	msg := &broker.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "Hello World"}`),
	}

	ctx := context.TODO()
	var wg sync.WaitGroup

	sub, err := b.Subscribe("test", func(p broker.Event) error {
		defer wg.Done()

		m := p.Message()

		if string(m.Body) != string(msg.Body) {
			t.Fatalf("Unexpected msg %s, expected %s", string(m.Body), string(msg.Body))
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Unexpected subscribe error: %v", err)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)

		if err := b.Publish(ctx, "test", msg); err != nil {
			t.Fatalf("Unexpected publish error: %v", err)
		}
	}

	wg.Wait()

	sub.Unsubscribe()

	if err := b.Disconnect(); err != nil {
		t.Fatalf("Unexpected disconnect error: %v", err)
	}
}

func BenchmarkSub1(b *testing.B) {
	sub(b, 1)
}
func BenchmarkSub8(b *testing.B) {
	sub(b, 8)
}

func BenchmarkSub32(b *testing.B) {
	sub(b, 32)
}

func BenchmarkSub64(b *testing.B) {
	sub(b, 64)
}

func BenchmarkSub128(b *testing.B) {
	sub(b, 128)
}

func BenchmarkPub1(b *testing.B) {
	pub(b, 1)
}

func BenchmarkPub8(b *testing.B) {
	pub(b, 8)
}

func BenchmarkPub32(b *testing.B) {
	pub(b, 32)
}

func BenchmarkPub64(b *testing.B) {
	pub(b, 64)
}

func BenchmarkPub128(b *testing.B) {
	pub(b, 128)
}
