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

	"github.com/vine-io/vine/core/client"
	rr "github.com/vine-io/vine/core/router"
	pb "github.com/vine-io/vine/proto/services/router"
)

type table struct {
	table    pb.TableService
	callOpts []client.CallOption
}

// Create new route in the routing table
func (t *table) Create(r rr.Route) error {
	route := &pb.Route{
		Service: r.Service,
		Address: r.Address,
		Gateway: r.Gateway,
		Network: r.Network,
		Link:    r.Link,
		Metric:  r.Metric,
	}

	if _, err := t.table.Create(context.Background(), route, t.callOpts...); err != nil {
		return err
	}

	return nil
}

// Delete deletes existing route from the routing table
func (t *table) Delete(r rr.Route) error {
	route := &pb.Route{
		Service: r.Service,
		Address: r.Address,
		Gateway: r.Gateway,
		Network: r.Network,
		Link:    r.Link,
		Metric:  r.Metric,
	}

	if _, err := t.table.Delete(context.Background(), route, t.callOpts...); err != nil {
		return err
	}

	return nil
}

// Update updates route in the routing table
func (t *table) Update(r rr.Route) error {
	route := &pb.Route{
		Service: r.Service,
		Address: r.Address,
		Gateway: r.Gateway,
		Network: r.Network,
		Link:    r.Link,
		Metric:  r.Metric,
	}

	if _, err := t.table.Update(context.Background(), route, t.callOpts...); err != nil {
		return err
	}

	return nil
}

// List returns the list of all routes in the table
func (t *table) List() ([]rr.Route, error) {
	resp, err := t.table.List(context.Background(), &pb.Request{}, t.callOpts...)
	if err != nil {
		return nil, err
	}

	routes := make([]rr.Route, len(resp.Routes))
	for i, route := range resp.Routes {
		routes[i] = rr.Route{
			Service: route.Service,
			Address: route.Address,
			Gateway: route.Gateway,
			Network: route.Network,
			Link:    route.Link,
			Metric:  route.Metric,
		}
	}

	return routes, nil
}

// Query lookup looks up routes in the routing table and returns them
func (t *table) Query(q ...rr.QueryOption) ([]rr.Route, error) {
	query := rr.NewQuery(q...)

	// call the router
	resp, err := t.table.Query(context.Background(), &pb.QueryRequest{
		Query: &pb.Query{
			Service: query.Service,
			Gateway: query.Gateway,
			Network: query.Network,
		},
	}, t.callOpts...)

	// errored out
	if err != nil {
		return nil, err
	}

	routes := make([]rr.Route, len(resp.Routes))
	for i, route := range resp.Routes {
		routes[i] = rr.Route{
			Service: route.Service,
			Address: route.Address,
			Gateway: route.Gateway,
			Network: route.Network,
			Link:    route.Link,
			Metric:  route.Metric,
		}
	}

	return routes, nil
}
