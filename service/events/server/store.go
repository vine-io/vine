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

package server

import (
	"context"

	"github.com/lack-io/vine/internal/auth/namespace"
	pb "github.com/lack-io/vine/proto/events"
	"github.com/lack-io/vine/service/errors"
	"github.com/lack-io/vine/service/events"
	goevents "github.com/lack-io/vine/service/events"
	"github.com/lack-io/vine/service/events/util"
)

type Store struct{}

func (s *Store) Read(ctx context.Context, req *pb.ReadRequest, rsp *pb.ReadResponse) error {
	// authorize the request
	if err := namespace.Authorize(ctx, namespace.DefaultNamespace); err == namespace.ErrForbidden {
		return errors.Forbidden("events.Store.Read", err.Error())
	} else if err == namespace.ErrUnauthorized {
		return errors.Unauthorized("events.Store.Read", err.Error())
	} else if err != nil {
		return errors.InternalServerError("events.Store.Read", err.Error())
	}

	// validate the request
	if len(req.Topic) == 0 {
		return errors.BadRequest("events.Store.Read", goevents.ErrMissingTopic.Error())
	}

	// parse options
	var opts []goevents.ReadOption
	if req.Limit > 0 {
		opts = append(opts, goevents.ReadLimit(uint(req.Limit)))
	}
	if req.Offset > 0 {
		opts = append(opts, goevents.ReadOffset(uint(req.Offset)))
	}

	// read from the store
	result, err := events.DefaultStore.Read(req.Topic, opts...)
	if err != nil {
		return errors.InternalServerError("events.Store.Read", err.Error())
	}

	// serialize the result
	rsp.Events = make([]*pb.Event, len(result))
	for i, r := range result {
		rsp.Events[i] = util.SerializeEvent(r)
	}

	return nil
}

func (s *Store) Write(ctx context.Context, req *pb.WriteRequest, rsp *pb.WriteResponse) error {
	return errors.NotImplemented("events.Store.Write", "Writing to the store directly is not supported")
}
