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

package router

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine"
	pb "github.com/lack-io/vine/proto/router"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/router"
	"github.com/lack-io/vine/service/router/handler"
	regRouter "github.com/lack-io/vine/service/router/registry"
)

var (
	// Name of the router vine service
	Name = "go.vine.router"
	// Address is the router vine service bind address
	Address = ":8084"
	// Network is the network name
	Network = router.DefaultNetwork
	// Topic is router adverts topic
	Topic = "go.vine.router.adverts"
)

// Sub processes router events
type sub struct {
	router router.Router
}

// Process processes router adverts
func (s *sub) Process(ctx context.Context, advert *pb.Advert) error {
	log.Debugf("received advert from: %s", advert.Id)
	if advert.Id == s.router.Options().Id {
		log.Debug("skipping advert")
		return nil
	}

	var events []*router.Event
	for _, event := range advert.Events {
		route := router.Route{
			Service: event.Route.Service,
			Address: event.Route.Address,
			Gateway: event.Route.Gateway,
			Network: event.Route.Network,
			Link:    event.Route.Link,
			Metric:  event.Route.Metric,
		}

		e := &router.Event{
			Type:      router.EventType(event.Type),
			Timestamp: time.Unix(0, advert.Timestamp),
			Route:     route,
		}

		events = append(events, e)
	}

	a := &router.Advert{
		Id:        advert.Id,
		Type:      router.AdvertType(advert.Type),
		Timestamp: time.Unix(0, advert.Timestamp),
		TTL:       time.Duration(advert.Ttl),
		Events:    events,
	}

	if err := s.router.Process(a); err != nil {
		return fmt.Errorf("failed processing advert: %s", err)
	}

	return nil
}

// rtr is vine router
type rtr struct {
	// router is the vine router
	router.Router
	// publisher to publish router adverts
	vine.Event
}

// newRouter creates new vine router and returns it
func newRouter(srv vine.Service, router router.Router) *rtr {
	s := &sub{
		router: router,
	}

	// register subscriber
	if err := vine.RegisterSubscriber(Topic, srv.Server(), s); err != nil {
		log.Errorf("failed to subscribe to adverts: %s", err)
		os.Exit(1)
	}

	return &rtr{
		Router: router,
		Event:  vine.NewEvent(Topic, srv.Client()),
	}
}

// PublishAdverts publishes adverts for other routers to consume
func (r *rtr) PublishAdverts(ch <-chan *router.Advert) error {
	for advert := range ch {
		var events []*pb.Event
		for _, event := range advert.Events {
			route := &pb.Route{
				Service: event.Route.Service,
				Address: event.Route.Address,
				Gateway: event.Route.Gateway,
				Network: event.Route.Network,
				Link:    event.Route.Link,
				Metric:  int64(event.Route.Metric),
			}
			e := &pb.Event{
				Type:      pb.EventType(event.Type),
				Timestamp: event.Timestamp.UnixNano(),
				Route:     route,
			}
			events = append(events, e)
		}

		a := &pb.Advert{
			Id:        r.Options().Id,
			Type:      pb.AdvertType(advert.Type),
			Timestamp: advert.Timestamp.UnixNano(),
			Events:    events,
		}

		if err := r.Publish(context.Background(), a); err != nil {
			log.Debugf("error publishing advert: %v", err)
			return fmt.Errorf("error publishing advert: %v", err)
		}
	}

	return nil
}

// Start starts the vine router
func (r *rtr) Start() error {
	// start the router
	if err := r.Router.Start(); err != nil {
		return fmt.Errorf("failed to start router: %v", err)
	}

	return nil
}

// Stop stops the vine router
func (r *rtr) Stop() error {
	// stop the router
	if err := r.Router.Stop(); err != nil {
		return fmt.Errorf("failed to stop router: %v", err)
	}

	return nil
}

// Run runs the vine server
func Run(ctx *cli.Context, srvOpts ...vine.Option) {

	// Init plugins
	for _, p := range Plugins() {
		p.Init(ctx)
	}

	if len(ctx.String("server-name")) > 0 {
		Name = ctx.String("server-name")
	}
	if len(ctx.String("address")) > 0 {
		Address = ctx.String("address")
	}
	if len(ctx.String("network")) > 0 {
		Network = ctx.String("network")
	}
	// default gateway address
	var gateway string
	if len(ctx.String("gateway")) > 0 {
		gateway = ctx.String("gateway")
	}

	// advertise the best routes
	strategy := router.AdvertiseLocal

	if a := ctx.String("advertise-strategy"); len(a) > 0 {
		switch a {
		case "all":
			strategy = router.AdvertiseAll
		case "best":
			strategy = router.AdvertiseBest
		case "local":
			strategy = router.AdvertiseLocal
		case "none":
			strategy = router.AdvertiseNone
		}
	}

	// Initialise service
	service := vine.NewService(
		vine.Name(Name),
		vine.Address(Address),
		vine.RegisterTTL(time.Duration(ctx.Int("register-ttl"))*time.Second),
		vine.RegisterInterval(time.Duration(ctx.Int("register-interval"))*time.Second),
	)

	r := regRouter.NewRouter(
		router.Id(service.Server().Options().Id),
		router.Address(service.Server().Options().Id),
		router.Network(Network),
		router.Registry(service.Client().Options().Registry),
		router.Gateway(gateway),
		router.Advertise(strategy),
	)

	// register router handler
	pb.RegisterRouterHandler(
		service.Server(),
		&handler.Router{
			Router: r,
		},
	)

	// register the table handler
	pb.RegisterTableHandler(
		service.Server(),
		&handler.Table{
			Router: r,
		},
	)

	// create new vine router and start advertising routes
	rtr := newRouter(service, r)

	log.Info("starting vine router")

	if err := rtr.Start(); err != nil {
		log.Errorf("failed to start: %s", err)
		os.Exit(1)
	}

	log.Info("starting to advertise")

	advertChan, err := rtr.Advertise()
	if err != nil {
		log.Errorf("failed to advertise: %s", err)
		log.Info("attempting to stop the router")
		if err := rtr.Stop(); err != nil {
			log.Errorf("failed to stop: %s", err)
			os.Exit(1)
		}
		os.Exit(1)
	}

	var wg sync.WaitGroup
	// error channel to collect router errors
	errChan := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		errChan <- rtr.PublishAdverts(advertChan)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		errChan <- service.Run()
	}()

	// we block here until either service or server fails
	if err := <-errChan; err != nil {
		log.Errorf("error running the router: %v", err)
	}

	log.Info("attempting to stop the router")

	// stop the router
	if err := r.Stop(); err != nil {
		log.Errorf("failed to stop: %s", err)
		os.Exit(1)
	}

	wg.Wait()

	log.Info("successfully stopped")
}

func Commands(options ...vine.Option) []*cli.Command {
	command := &cli.Command{
		Name:  "router",
		Usage: "Run the vine network router",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "address",
				Usage:   "Set the vine router address :9093",
				EnvVars: []string{"VINE_SERVER_ADDRESS"},
			},
			&cli.StringFlag{
				Name:    "network",
				Usage:   "Set the vine network name: local",
				EnvVars: []string{"VINE_NETWORK_NAME"},
			},
			&cli.StringFlag{
				Name:    "gateway",
				Usage:   "Set the vine default gateway address. Defaults to none.",
				EnvVars: []string{"VINE_GATEWAY_ADDRESS"},
			},
			&cli.StringFlag{
				Name:    "advertise-strategy",
				Usage:   "Set the advertise strategy; all, best, local, none",
				EnvVars: []string{"VINE_ROUTER_ADVERTISE_STRATEGY"},
			},
		},
		Action: func(ctx *cli.Context) error {
			Run(ctx, options...)
			return nil
		},
	}

	for _, p := range Plugins() {
		if cmds := p.Commands(); len(cmds) > 0 {
			command.Subcommands = append(command.Subcommands, cmds...)
		}

		if flags := p.Flags(); len(flags) > 0 {
			command.Flags = append(command.Flags, flags...)
		}
	}

	return []*cli.Command{command}
}
