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

package grpc

import (
	"context"
	"time"

	pb "github.com/lack-io/vine/proto/services/broker"
	"github.com/lack-io/vine/service/broker"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/logger"
)

type gRPCBroker struct {
	Addrs   []string
	Client  pb.BrokerService
	options broker.Options
}

var (
	DefaultName = "go.vine.broker"
)

func (b *gRPCBroker) Address() string {
	return b.Addrs[0]
}

func (b *gRPCBroker) Connect() error {
	return nil
}

func (b *gRPCBroker) Disconnect() error {
	return nil
}

func (b *gRPCBroker) Init(opts ...broker.Option) error {
	for _, o := range opts {
		o(&b.options)
	}
	return nil
}

func (b *gRPCBroker) Options() broker.Options {
	return b.options
}

func (b *gRPCBroker) Publish(topic string, msg *broker.Message, opts ...broker.PublishOption) error {
	logger.Debugf("Publishing to topic %s broker %v", topic, b.Addrs)

	_, err := b.Client.Publish(context.TODO(), &pb.PublishRequest{
		Topic: topic,
		Message: &pb.Message{
			Header: msg.Header,
			Body:   msg.Body,
		},
	}, client.WithAddress(b.Addrs...))
	return err
}

func (b *gRPCBroker) Subscribe(topic string, handler broker.Handler, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	var options broker.SubscribeOptions
	for _, o := range opts {
		o(&options)
	}

	logger.Debugf("Subscribing to topic %s queue %s broker %v", topic, options.Queue, b.Addrs)
	stream, err := b.Client.Subscribe(context.TODO(), &pb.SubscribeRequest{
		Topic: topic,
		Queue: options.Queue,
	}, client.WithAddress(b.Addrs...), client.WithRequestTimeout(time.Hour))
	if err != nil {
		return nil, err
	}

	sub := &serviceSub{
		topic:   topic,
		queue:   options.Queue,
		handler: handler,
		stream:  stream,
		closed:  make(chan bool),
		options: options,
	}

	go func() {
		for {
			select {
			case <-sub.closed:
				logger.Debugf("Unsubscribed from topic %s", topic)
				return
			default:
				// run the subscriber
				logger.Debugf("Streaming from broker %v to topic [%s] queue [%s]", b.Addrs, topic, options.Queue)
				if err := sub.run(); err != nil {
					logger.Debugf("Resubscribing to topic %s broker %v", topic, b.Addrs)

					stream, err := b.Client.Subscribe(context.TODO(), &pb.SubscribeRequest{
						Topic: topic,
						Queue: options.Queue,
					}, client.WithAddress(b.Addrs...), client.WithRequestTimeout(time.Hour))
					if err != nil {
						logger.Debugf("Failed to resubscribe to topic %s: %v", topic, err)
						time.Sleep(time.Second)
						continue
					}

					// new stream
					sub.stream = stream
				}
			}
		}
	}()

	return sub, nil
}

func (b *gRPCBroker) String() string {
	return "grpc"
}

func NewBroker(opts ...broker.Option) broker.Broker {
	var options broker.Options
	for _, o := range opts {
		o(&options)
	}

	addrs := options.Addrs
	if len(addrs) == 0 {
		addrs = []string{"127.0.0.1:8001"}
	}

	cli := client.DefaultClient

	return &gRPCBroker{
		Addrs:   addrs,
		Client:  pb.NewBrokerService(DefaultName, cli),
		options: options,
	}
}
