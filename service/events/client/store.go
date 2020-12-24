// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	pb "github.com/lack-io/vine/proto/events"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/context"
	"github.com/lack-io/vine/service/events"
	"github.com/lack-io/vine/service/events/util"
)

// NewStore returns an initialized store handler
func NewStore() events.Store {
	return new(store)
}

type store struct {
	Client pb.StoreService
}

func (s *store) Read(topic string, opts ...events.ReadOption) ([]*events.Event, error) {
	// parse the options
	var options events.ReadOptions
	for _, o := range opts {
		o(&options)
	}

	// execute the RPC
	rsp, err := s.client().Read(context.DefaultContext, &pb.ReadRequest{
		Topic:  topic,
		Limit:  uint64(options.Limit),
		Offset: uint64(options.Offset),
	}, client.WithAuthToken())
	if err != nil {
		return nil, err
	}

	// serialize the response
	result := make([]*events.Event, len(rsp.Events))
	for i, r := range rsp.Events {
		ev := util.DeserializeEvent(r)
		result[i] = &ev
	}

	return result, nil
}

func (s *store) Write(ev *events.Event, opts ...events.WriteOption) error {
	// parse options
	var options events.WriteOptions
	for _, o := range opts {
		o(&options)
	}

	// start the stream
	_, err := s.client().Write(context.DefaultContext, &pb.WriteRequest{
		Event: &pb.Event{
			Id:        ev.ID,
			Topic:     ev.Topic,
			Metadata:  ev.Metadata,
			Payload:   ev.Payload,
			Timestamp: ev.Timestamp.Unix(),
		},
	}, client.WithAuthToken())

	return err
}

// this is a tmp solution since the client isn't initialized when NewStream is called. There is a
// fix in the works in another PR.
func (s *store) client() pb.StoreService {
	if s.Client == nil {
		s.Client = pb.NewStoreService("events", client.DefaultClient)
	}
	return s.Client
}
