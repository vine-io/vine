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
	pb "github.com/lack-io/vine/proto/router"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/context"
	"github.com/lack-io/vine/service/router"
)

type table struct {
	table    pb.TableService
	callOpts []client.CallOption
}

// Create new route in the routing table
func (t *table) Create(r router.Route) error {
	route := &pb.Route{
		Service: r.Service,
		Address: r.Address,
		Gateway: r.Gateway,
		Network: r.Network,
		Link:    r.Link,
		Metric:  r.Metric,
	}

	if _, err := t.table.Create(context.DefaultContext, route, t.callOpts...); err != nil {
		return err
	}

	return nil
}

// Delete deletes existing route from the routing table
func (t *table) Delete(r router.Route) error {
	route := &pb.Route{
		Service: r.Service,
		Address: r.Address,
		Gateway: r.Gateway,
		Network: r.Network,
		Link:    r.Link,
		Metric:  r.Metric,
	}

	if _, err := t.table.Delete(context.DefaultContext, route, t.callOpts...); err != nil {
		return err
	}

	return nil
}

// Update updates route in the routing table
func (t *table) Update(r router.Route) error {
	route := &pb.Route{
		Service: r.Service,
		Address: r.Address,
		Gateway: r.Gateway,
		Network: r.Network,
		Link:    r.Link,
		Metric:  r.Metric,
	}

	if _, err := t.table.Update(context.DefaultContext, route, t.callOpts...); err != nil {
		return err
	}

	return nil
}

// Read looks up routes in the routing table and returns them
func (t *table) Read(opts ...router.ReadOption) ([]router.Route, error) {
	var options router.ReadOptions
	for _, o := range opts {
		o(&options)
	}
	// call the router
	resp, err := t.table.Read(context.DefaultContext, &pb.ReadRequest{
		Service: options.Service,
	}, t.callOpts...)
	// errored out
	if err != nil {
		return nil, err
	}

	routes := make([]router.Route, len(resp.Routes))
	for i, route := range resp.Routes {
		routes[i] = router.Route{
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
