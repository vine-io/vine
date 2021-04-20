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
	"io"
	"time"

	rr "github.com/lack-io/vine/core/router"
	"github.com/lack-io/vine/proto/apis/errors"
	pb "github.com/lack-io/vine/proto/services/router"
)

// Router implements router handler
type Router struct {
	Router rr.Router
}

// Lookup looks up routes in the routing table and returns them
func (r *Router) Lookup(ctx context.Context, req *pb.LookupRequest, resp *pb.LookupResponse) error {
	routes, err := r.Router.Lookup(rr.QueryService(req.Query.Service))
	if err != nil {
		return errors.InternalServerError("go.vine.router", "failed to lookup routes: %v", err)
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

// Advertise streams router advertisements
func (r *Router) Advertise(ctx context.Context, req *pb.Request, stream pb.Router_AdvertiseStream) error {
	advertChan, err := r.Router.Advertise()
	if err != nil {
		return errors.InternalServerError("go.vine.router", "failed to get adverts: %v", err)
	}

	for advert := range advertChan {
		var events []*pb.Event
		for _, event := range advert.Events {
			route := &pb.Route{
				Service: event.Route.Service,
				Address: event.Route.Address,
				Gateway: event.Route.Gateway,
				Network: event.Route.Network,
				Router:  event.Route.Router,
				Link:    event.Route.Link,
				Metric:  event.Route.Metric,
			}
			e := &pb.Event{
				Id:        event.Id,
				Type:      pb.EventType(event.Type),
				Timestamp: event.Timestamp.UnixNano(),
				Route:     route,
			}
			events = append(events, e)
		}

		advert := &pb.Advert{
			Id:        advert.Id,
			Type:      pb.AdvertType(advert.Type),
			Timestamp: advert.Timestamp.UnixNano(),
			Events:    events,
		}

		// send the advert
		err := stream.Send(advert)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return errors.InternalServerError("go.vine.router", "error sending message %v", err)
		}
	}

	return nil
}

// Process processes advertisements
func (r *Router) Process(ctx context.Context, req *pb.Advert, rsp *pb.ProcessResponse) error {
	events := make([]*rr.Event, len(req.Events))
	for i, event := range req.Events {
		route := rr.Route{
			Service: event.Route.Service,
			Address: event.Route.Address,
			Gateway: event.Route.Gateway,
			Network: event.Route.Network,
			Router:  event.Route.Router,
			Link:    event.Route.Link,
			Metric:  event.Route.Metric,
		}

		events[i] = &rr.Event{
			Id:        event.Id,
			Type:      rr.EventType(event.Type),
			Timestamp: time.Unix(0, event.Timestamp),
			Route:     route,
		}
	}

	advert := &rr.Advert{
		Id:        req.Id,
		Type:      rr.AdvertType(req.Type),
		Timestamp: time.Unix(0, req.Timestamp),
		TTL:       time.Duration(req.Ttl),
		Events:    events,
	}

	if err := r.Router.Process(advert); err != nil {
		return errors.InternalServerError("go.vine.router", "error publishing advert: %v", err)
	}

	return nil
}

// Watch streans routing table events
func (r *Router) Watch(ctx context.Context, req *pb.WatchRequest, stream pb.Router_WatchStream) error {
	watcher, err := r.Router.Watch()
	if err != nil {
		return errors.InternalServerError("go.vine.router", "failed creating event watcher: %v", err)
	}
	defer watcher.Stop()
	defer stream.Close()

	for {
		event, err := watcher.Next()
		if err == rr.ErrWatcherStopped {
			return errors.InternalServerError("go.vine.router", "watcher stopped")
		}

		if err != nil {
			return errors.InternalServerError("go.vine.router", "error watching events: %v", err)
		}

		route := &pb.Route{
			Service: event.Route.Service,
			Address: event.Route.Address,
			Gateway: event.Route.Gateway,
			Network: event.Route.Network,
			Router:  event.Route.Router,
			Link:    event.Route.Link,
			Metric:  event.Route.Metric,
		}

		tableEvent := &pb.Event{
			Id:        event.Id,
			Type:      pb.EventType(event.Type),
			Timestamp: event.Timestamp.UnixNano(),
			Route:     route,
		}

		if err := stream.Send(tableEvent); err != nil {
			return err
		}
	}
}
