// Copyright 2020 The vine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpc

import (
	pb "github.com/lack-io/vine/proto/broker"
	"github.com/lack-io/vine/service/broker"
	"github.com/lack-io/vine/service/logger"
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
