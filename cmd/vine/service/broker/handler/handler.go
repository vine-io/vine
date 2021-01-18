// Copyright 2020 lack
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

package handler

import (
	"context"

	pb "github.com/lack-io/vine/proto/broker"
	"github.com/lack-io/vine/proto/errors"
	"github.com/lack-io/vine/service/broker"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/util/namespace"
)

type Broker struct {
	Broker broker.Broker
}

func (b *Broker) Publish(ctx context.Context, req *pb.PublishRequest, rsp *pb.Empty) error {
	ns := namespace.FromContext(ctx)

	log.Debugf("Publishing message to %s topic in the %v namespace", req.Topic, ns)
	err := b.Broker.Publish(ns+"."+req.Topic, &broker.Message{
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
	handler := func(p broker.Event) error {
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
	sub, err := b.Broker.Subscribe(ns+"."+req.Topic, handler, broker.Queue(ns+"."+req.Queue))
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
