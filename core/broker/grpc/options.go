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
	"github.com/lack-io/vine/core/broker"
	"github.com/lack-io/vine/lib/logger"
	pb "github.com/lack-io/vine/proto/services/broker"
)

type serviceSub struct {
	topic   string
	queue   string
	handler broker.Handler
	stream  pb.Broker_SubscribeService
	closed  chan bool
	options broker.SubscribeOptions
}

type serviceEvent struct {
	topic   string
	err     error
	message *broker.Message
}

func (s *serviceEvent) Topic() string {
	return s.topic
}

func (s *serviceEvent) Message() *broker.Message {
	return s.message
}

func (s *serviceEvent) Ack() error {
	return nil
}

func (s *serviceEvent) Error() error {
	return s.err
}

func (s *serviceSub) isClosed() bool {
	select {
	case <-s.closed:
		return true
	default:
		return false
	}
}

func (s *serviceSub) run() error {
	exit := make(chan bool)
	go func() {
		select {
		case <-exit:
		case <-s.closed:
		}

		// close the stream
		s.stream.Close()
	}()

	for {
		// TODO: do not fail silently
		msg, err := s.stream.Recv()
		if err != nil {
			logger.Debugf("Streaming error for subscription to topic %s: %v", s.Topic(), err)

			// close the exit channel
			close(exit)

			// don't return an error if we unsubscribed
			if s.isClosed() {
				return nil
			}

			// return stream error
			return err
		}

		p := &serviceEvent{
			topic: s.topic,
			message: &broker.Message{
				Header: msg.Header,
				Body:   msg.Body,
			},
		}

		p.err = s.handler(p)
	}
}

func (s *serviceSub) Options() broker.SubscribeOptions {
	return s.options
}

func (s *serviceSub) Topic() string {
	return s.topic
}

func (s *serviceSub) Unsubscribe() error {
	select {
	case <-s.closed:
		return nil
	default:
		close(s.closed)
	}
	return nil
}
