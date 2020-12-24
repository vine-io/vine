// Copyright 2020 The vine Authors
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

package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/cmd/vine/service/config/handler"
	proto "github.com/lack-io/vine/proto/config"
	"github.com/lack-io/vine/service"
	"github.com/lack-io/vine/service/config/cmd"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/util/client"
	"github.com/lack-io/vine/util/helper"
)

var (
	// Service name
	Name = "go.vine.config"
	// Default database store
	Database = "store"
	// Default key
	Namespace = "global"
)

func Run(c *cli.Context, srvOpts ...service.Option) {
	if len(c.String("server-name")) > 0 {
		Name = c.String("server-name")
	}

	if len(c.String("watch-topic")) > 0 {
		handler.WatchTopic = c.String("watch-topic")
	}

	srvOpts = append(srvOpts, service.Name(Name))

	srv := service.NewService(srvOpts...)

	h := &handler.Config{
		Store: *cmd.DefaultCmd.Options().Store,
	}

	proto.RegisterConfigHandler(srv.Server(), h)
	service.RegisterSubscriber(handler.WatchTopic, srv.Server(), handler.Watcher)

	if err := srv.Run(); err != nil {
		log.Fatalf("config Run the service error: ", err)
	}
}

func setConfig(ctx *cli.Context) error {
	pb := proto.NewConfigService("go.vine.config", client.New(ctx))

	args := ctx.Args()

	if args.Len() == 0 {
		fmt.Println("Required usage: vine config set key val")
		os.Exit(1)
	}

	// key val
	key := args.Get(0)
	val := args.Get(1)

	// TODO: allow the specifying of a config.Key. This will be service name
	// The actuall key-val set is a path e.g vine/accounts/key
	_, err := pb.Update(context.TODO(), &proto.UpdateRequest{
		Change: &proto.Change{
			// global key
			Namespace: Namespace,
			// actual key for the value
			Path: key,
			// The value
			ChangeSet: &proto.ChangeSet{
				Data:      string(val),
				Format:    "json",
				Source:    "cli",
				Timestamp: time.Now().Unix(),
			},
		},
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}

func getConfig(ctx *cli.Context) error {
	pb := proto.NewConfigService("go.vine.config", client.New(ctx))

	args := ctx.Args()

	if args.Len() == 0 {
		fmt.Println("Required usage: vine config get key")
		os.Exit(1)
	}

	// key val
	key := args.Get(0)

	if len(key) == 0 {
		log.Fatal("key cannot be blank")
	}

	// TODO: allow the specifying of a config.Key. This will be service name
	// The actuall key-val set is a path e.g vine/accounts/key

	rsp, err := pb.Read(context.TODO(), &proto.ReadRequest{
		// The global key,
		Namespace: Namespace,
		// The actual key for the val
		Path: key,
	})

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fmt.Println("not found")
			os.Exit(1)
		}
		fmt.Println(err)
		os.Exit(1)
	}

	if rsp.Change == nil || rsp.Change.ChangeSet == nil {
		fmt.Println("not found")
		os.Exit(1)
	}

	// don't do it
	if v := rsp.Change.ChangeSet.Data; len(v) == 0 || string(v) == "null" {
		fmt.Println("not found")
		os.Exit(1)
	}

	fmt.Println(string(rsp.Change.ChangeSet.Data))

	return nil
}

func delConfig(ctx *cli.Context) error {
	pb := proto.NewConfigService("go.vine.config", client.New(ctx))

	args := ctx.Args()

	if args.Len() == 0 {
		fmt.Println("Required usage: vine config get key")
		os.Exit(1)
	}

	// key val
	key := args.Get(0)

	if len(key) == 0 {
		log.Fatal("key cannot be blank")
	}

	// TODO: allow the specifying of a config.Key. This will be service name
	// The actuall key-val set is a path e.g vine/accounts/key

	_, err := pb.Delete(context.TODO(), &proto.DeleteRequest{
		Change: &proto.Change{
			// The global key,
			Namespace: Namespace,
			// The actual key for the val
			Path: key,
		},
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}

func Commands(options ...service.Option) []*cli.Command {
	command := &cli.Command{
		Name:  "config",
		Usage: "Manage configuration values",
		Subcommands: []*cli.Command{
			{
				Name:   "get",
				Usage:  "Get a value; vine config get key",
				Action: getConfig,
			},
			{
				Name:   "set",
				Usage:  "Set a key-val; vine config set key val",
				Action: setConfig,
			},
			{
				Name:   "del",
				Usage:  "Delete a value; vine config del key",
				Action: delConfig,
			},
		},
		Action: func(ctx *cli.Context) error {
			if err := helper.UnexpectedSubcommand(ctx); err != nil {
				return err
			}
			Run(ctx, options...)
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "namespace",
				EnvVars: []string{"VINE_CONFIG_NAMESPACE"},
				Usage:   "Set the namespace used by the Config Service e.g. go.vine.srv.config",
			},
			&cli.StringFlag{
				Name:    "watch-topic",
				EnvVars: []string{"VINE_CONFIG_WATCH_TOPIC"},
				Usage:   "watch the change event.",
			},
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
