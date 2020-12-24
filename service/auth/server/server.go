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

	"github.com/lack-io/vine/internal/auth/token"
	"github.com/lack-io/vine/internal/auth/token/jwt"
	pb "github.com/lack-io/vine/proto/auth"
	"github.com/lack-io/vine/service"
	"github.com/lack-io/vine/service/auth"
	authHandler "github.com/lack-io/vine/service/auth/server/auth"
	rulesHandler "github.com/lack-io/vine/service/auth/server/rules"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/store"
	mustore "github.com/lack-io/vine/service/store"
)

// Flags specific to the router
var Flags = []cli.Flag{
	&cli.BoolFlag{
		Name:    "disable-admin",
		EnvVars: []string{"VINE_AUTH_DISABLE_ADMIN"},
		Usage:   "Prevent generation of default accounts in namespaces",
	},
}

const (
	name    = "auth"
	address = ":8010"
)

// Run the auth service
func Run(ctx *cli.Context) error {
	srv := service.New(
		service.Name(name),
		service.Address(address),
	)

	// setup the handlers
	ruleH := &rulesHandler.Rules{}
	authH := &authHandler.Auth{
		DisableAdmin: ctx.Bool("disable-admin"),
	}

	// setup the auth handler to use JWTs
	authH.TokenProvider = jwt.NewTokenProvider(
		token.WithPublicKey(auth.DefaultAuth.Options().PublicKey),
		token.WithPrivateKey(auth.DefaultAuth.Options().PrivateKey),
	)

	// set the handlers store
	mustore.DefaultStore.Init(store.Table("auth"))
	authH.Init(auth.Store(mustore.DefaultStore))
	ruleH.Init(auth.Store(mustore.DefaultStore))

	// register handlers
	pb.RegisterAuthHandler(srv.Server(), authH)
	pb.RegisterRulesHandler(srv.Server(), ruleH)
	pb.RegisterAccountsHandler(srv.Server(), authH)

	// run service
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
	return nil
}
