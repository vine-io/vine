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

	pb "github.com/lack-io/vine/proto/store"
	"github.com/lack-io/vine/service"
	log "github.com/lack-io/vine/service/logger"
)

var (
	// name of the store service
	name = "store"
	// address is the store address
	address = ":8002"
)

// Run vine store
func Run(ctx *cli.Context) error {
	if len(ctx.String("server-name")) > 0 {
		name = ctx.String("server-name")
	}
	if len(ctx.String("address")) > 0 {
		address = ctx.String("address")
	}

	// Initialise service
	service := service.New(
		service.Name(name),
		service.Address(address),
	)

	// the store handler
	pb.RegisterStoreHandler(service.Server(), &handler{
		stores: make(map[string]bool),
	})

	// the blob store handler
	pb.RegisterBlobStoreHandler(service.Server(), new(blobHandler))

	// start the service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
	return nil
}
