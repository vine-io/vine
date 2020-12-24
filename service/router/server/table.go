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

	pb "github.com/lack-io/vine/proto/router"
	"github.com/lack-io/vine/service/errors"
	"github.com/lack-io/vine/service/router"
)

type Table struct {
	Router router.Router
}

func (t *Table) Create(ctx context.Context, route *pb.Route, resp *pb.CreateResponse) error {
	err := t.Router.Table().Create(router.Route{
		Service:  route.Service,
		Address:  route.Address,
		Gateway:  route.Gateway,
		Network:  route.Network,
		Router:   route.Router,
		Link:     route.Link,
		Metric:   route.Metric,
		Metadata: route.Metadata,
	})
	if err != nil {
		return errors.InternalServerError("router.Table.Create", "failed to create route: %s", err)
	}

	return nil
}

func (t *Table) Update(ctx context.Context, route *pb.Route, resp *pb.UpdateResponse) error {
	err := t.Router.Table().Update(router.Route{
		Service:  route.Service,
		Address:  route.Address,
		Gateway:  route.Gateway,
		Network:  route.Network,
		Router:   route.Router,
		Link:     route.Link,
		Metric:   route.Metric,
		Metadata: route.Metadata,
	})
	if err != nil {
		return errors.InternalServerError("router.Table.Update", "failed to update route: %s", err)
	}

	return nil
}

func (t *Table) Delete(ctx context.Context, route *pb.Route, resp *pb.DeleteResponse) error {
	err := t.Router.Table().Delete(router.Route{
		Service:  route.Service,
		Address:  route.Address,
		Gateway:  route.Gateway,
		Network:  route.Network,
		Router:   route.Router,
		Link:     route.Link,
		Metric:   route.Metric,
		Metadata: route.Metadata,
	})
	if err != nil {
		return errors.InternalServerError("route.Table.Delete", "failed to delete route: %s", err)
	}

	return nil
}

func (t *Table) Read(ctx context.Context, req *pb.ReadRequest, resp *pb.ReadResponse) error {
	routes, err := t.Router.Table().Read(router.ReadService(req.Service))
	if err != nil {
		return errors.InternalServerError("router.Table.Read", "failed to lookup routes: %s", err)
	}

	respRoutes := make([]*pb.Route, 0, len(routes))
	for _, route := range routes {
		respRoute := &pb.Route{
			Service:  route.Service,
			Address:  route.Address,
			Gateway:  route.Gateway,
			Network:  route.Network,
			Router:   route.Router,
			Link:     route.Link,
			Metric:   route.Metric,
			Metadata: route.Metadata,
		}
		respRoutes = append(respRoutes, respRoute)
	}

	resp.Routes = respRoutes

	return nil
}
