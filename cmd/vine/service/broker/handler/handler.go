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

package handler

import (
	"context"

	broker2 "github.com/lack-io/vine/core/broker"
	log "github.com/lack-io/vine/lib/logger"
	"github.com/lack-io/vine/proto/apis/errors"
	pb "github.com/lack-io/vine/proto/services/broker"
	"github.com/lack-io/vine/util/namespace"
)

type Broker struct {
	Broker broker2.Broker
}

func (b *Broker) Publish(ctx context.Context, req *pb.PublishRequest, rsp *pb.Empty) error {
	ns := namespace.FromContext(ctx)

	log.Debugf("Publishing message to %s topic in the %v namespace", req.Topic, ns)
	err := b.Broker.Publish(ns+"."+req.Topic, &broker2.Message{
		Header: req.Message.Header,
		Body:   req.Message.Body,
	})
	log.Debugf("Published message to %s topic in the %v namespace", req.Topic, ns)
	if err != nil {
		return errors.InternalServerError("go.vine.broker", err.Error())
	}
	return nil
}

func (b *Broker) Subscribe(ctx context.Context, req *pb.SubscribeRequest, stream pb.Broker_SubscribeStream) error {
	ns := namespace.FromContext(ctx)
	errChan := make(chan error, 1)

	// message handler to stream back messages from broker
	handler := func(p broker2.Event) error {
		if err := stream.Send(&pb.Message{
			Header: p.Message().Header,
			Body:   p.Message().Body,
		}); err != nil {
			select {
			case errChan <- err:
				return err
			default:
				return err
			}
		}
		return nil
	}

	log.Debugf("Subscribing to %s topic in namespace %v", req.Topic, ns)
	sub, err := b.Broker.Subscribe(ns+"."+req.Topic, handler, broker2.Queue(ns+"."+req.Queue))
	if err != nil {
		return errors.InternalServerError("go.vine.broker", err.Error())
	}
	defer func() {
		log.Debugf("Unsubscribing from topic %s in namespace %v", req.Topic, ns)
		sub.Unsubscribe()
	}()

	select {
	case <-ctx.Done():
		log.Debugf("Context done for subscription to topic %s", req.Topic)
		return nil
	case err := <-errChan:
		log.Debugf("Subscription error for topic %s: %v", req.Topic, err)
		return err
	}
}
