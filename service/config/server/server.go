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
	"github.com/lack-io/cli"

	pb "github.com/lack-io/vine/proto/config"
	"github.com/lack-io/vine/service"
	"github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/store"
	mustore "github.com/lack-io/vine/service/store"
)

const (
	name    = "config"
	address = ":8001"
)

var (
	// Flags specific to the config service
	Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "watch-topic",
			EnvVars: []string{"VINE_CONFIG_SECRET_KEY"},
			Usage:   "watch the change event.",
		},
	}
)

// Run vine config
func Run(c *cli.Context) error {
	srv := service.New(
		service.Name(name),
		service.Address(address),
	)

	mustore.DefaultStore.Init(store.Table("config"))

	// register the handler
	pb.RegisterConfigHandler(srv.Server(), NewConfig(c.String("config-secret-key")))
	// register the subscriber
	//srv.Subscribe(watchTopic, new(watcher))

	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
	return nil
}
