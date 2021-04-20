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

	rr "github.com/lack-io/vine/core/router"
	"github.com/lack-io/vine/proto/apis/errors"
	pb "github.com/lack-io/vine/proto/services/router"
)

type Table struct {
	Router rr.Router
}

func (t *Table) Create(ctx context.Context, route *pb.Route, resp *pb.CreateResponse) error {
	err := t.Router.Table().Create(rr.Route{
		Service: route.Service,
		Address: route.Address,
		Gateway: route.Gateway,
		Network: route.Network,
		Router:  route.Router,
		Link:    route.Link,
		Metric:  route.Metric,
	})
	if err != nil {
		return errors.InternalServerError("go.vine.router", "failed to create route: %s", err)
	}

	return nil
}

func (t *Table) Update(ctx context.Context, route *pb.Route, resp *pb.UpdateResponse) error {
	err := t.Router.Table().Update(rr.Route{
		Service: route.Service,
		Address: route.Address,
		Gateway: route.Gateway,
		Network: route.Network,
		Router:  route.Router,
		Link:    route.Link,
		Metric:  route.Metric,
	})
	if err != nil {
		return errors.InternalServerError("go.vine.router", "failed to update route: %s", err)
	}

	return nil
}

func (t *Table) Delete(ctx context.Context, route *pb.Route, resp *pb.DeleteResponse) error {
	err := t.Router.Table().Delete(rr.Route{
		Service: route.Service,
		Address: route.Address,
		Gateway: route.Gateway,
		Network: route.Network,
		Router:  route.Router,
		Link:    route.Link,
		Metric:  route.Metric,
	})
	if err != nil {
		return errors.InternalServerError("go.vine.router", "failed to delete route: %s", err)
	}

	return nil
}

// List returns all routes in the routing table
func (t *Table) List(ctx context.Context, req *pb.Request, resp *pb.ListResponse) error {
	routes, err := t.Router.Table().List()
	if err != nil {
		return errors.InternalServerError("go.vine.router", "failed to list routes: %s", err)
	}

	respRoutes := make([]*pb.Route, 0, len(routes))
	for _, route := range routes {
		respRoute := &pb.Route{
			Service: route.Service,
			Address: route.Address,
			Gateway: route.Gateway,
			Network: route.Network,
			Router:  route.Router,
			Link:    route.Link,
			Metric:  route.Metric,
		}
		respRoutes = append(respRoutes, respRoute)
	}

	resp.Routes = respRoutes

	return nil
}

func (t *Table) Query(ctx context.Context, req *pb.QueryRequest, resp *pb.QueryResponse) error {
	routes, err := t.Router.Table().Query(rr.QueryService(req.Query.Service))
	if err != nil {
		return errors.InternalServerError("go.vine.router", "failed to lookup routes: %s", err)
	}

	respRoutes := make([]*pb.Route, 0, len(routes))
	for _, route := range routes {
		respRoute := &pb.Route{
			Service: route.Service,
			Address: route.Address,
			Gateway: route.Gateway,
			Network: route.Network,
			Router:  route.Router,
			Link:    route.Link,
			Metric:  route.Metric,
		}
		respRoutes = append(respRoutes, respRoute)
	}

	resp.Routes = respRoutes

	return nil
}
