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

// Package broker is the vine broker
package broker

import (
	"time"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine"
	"github.com/lack-io/vine/cmd/vine/service/broker/handler"
	pb "github.com/lack-io/vine/proto/services/broker"
)

var (
	// Name of the broker
	Name = "go.vine.broker"
	// The address of the broker
	Address = ":8001"
)

func Run(ctx *cli.Context, svcOpts ...vine.Option) {

	if len(ctx.String("server-name")) > 0 {
		Name = ctx.String("server-name")
	}
	if len(ctx.String("address")) > 0 {
		Address = ctx.String("address")
	}

	// service opts
	svcOpts = append(svcOpts, vine.Name(Name))
	if i := time.Duration(ctx.Int("register-ttl")); i > 0 {
		svcOpts = append(svcOpts, vine.RegisterTTL(i*time.Second))
	}
	if i := time.Duration(ctx.Int("register-interval")); i > 0 {
		svcOpts = append(svcOpts, vine.RegisterInterval(i*time.Second))
	}

	// set address
	if len(Address) > 0 {
		svcOpts = append(svcOpts, vine.Address(Address))
	}

	// new service
	svc := vine.NewService(svcOpts...)

	// connect to the broker
	svc.Options().Broker.Connect()

	// register the broker handler
	pb.RegisterBrokerHandler(svc.Server(), &handler.Broker{
		// using the mdns broker
		Broker: svc.Options().Broker,
	})

	// run the service
	svc.Run()
}

func Commands(options ...vine.Option) []*cli.Command {
	command := &cli.Command{
		Name:  "broker",
		Usage: "Run the message broker",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "address",
				Usage:   "Set the broker http address e.g 0.0.0.0:8001",
				EnvVars: []string{"VINE_SERVER_ADDRESS"},
			},
		},
		Action: func(ctx *cli.Context) error {
			Run(ctx, options...)
			return nil
		},
	}

	return []*cli.Command{command}
}
