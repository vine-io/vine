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

package auth

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine"
	cliutil "github.com/lack-io/vine/cmd/vine/client/cli/util"
	"github.com/lack-io/vine/cmd/vine/service/auth/api"
	authHandler "github.com/lack-io/vine/cmd/vine/service/auth/handler/auth"
	rulesHandler "github.com/lack-io/vine/cmd/vine/service/auth/handler/rules"
	"github.com/lack-io/vine/proto/apis/errors"
	pb "github.com/lack-io/vine/proto/auth"
	"github.com/lack-io/vine/service/auth"
	svcAuth "github.com/lack-io/vine/service/auth/grpc"
	"github.com/lack-io/vine/service/auth/token"
	"github.com/lack-io/vine/service/auth/token/jwt"
	"github.com/lack-io/vine/service/config/cmd"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/util/client"
	"github.com/lack-io/vine/util/config"
	"github.com/lack-io/vine/util/helper"
)

var (
	// Name of the service
	Name = "go.vine.auth"
	// Address of the service
	Address = ":8010"
	// ServiceFlags are provided to commands which run vine services
	ServiceFlags = []cli.Flag{
		&cli.StringFlag{
			Name:    "address",
			Usage:   "Set the auth http address e.g 0.0.0.0:8010",
			EnvVars: []string{"VINE_SERVER_ADDRESS"},
		},
		&cli.StringFlag{
			Name:    "auth-provider",
			EnvVars: []string{"VINE_AUTH_PROVIDER"},
			Usage:   "Auth provider enables account generation",
		},
		&cli.StringFlag{
			Name:    "auth-public-key",
			EnvVars: []string{"VINE_AUTH_PUBLIC_KEY"},
			Usage:   "Public key for JWT auth (base64 encoded PEM)",
		},
		&cli.StringFlag{
			Name:    "auth-private-key",
			EnvVars: []string{"VINE_AUTH_PRIVATE_KEY"},
			Usage:   "Private key for JWT auth (base64 encoded PEM)",
		},
	}
	// RuleFlags are provided to commands which create or delete rules
	RuleFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  "scope",
			Usage: "The scope to amend, e.g. 'user' or '*', leave blank to make public",
		},
		&cli.StringFlag{
			Name:  "resource",
			Usage: "The resource to amend in the format type:name:endpoint, e.g. service:go.vine.auth:*",
		},
		&cli.StringFlag{
			Name:  "access",
			Usage: "The access level, must be granted or denied",
			Value: "granted",
		},
		&cli.IntFlag{
			Name:  "priority",
			Usage: "The priority level, default is 0, the greater the number the higher the priority",
			Value: 0,
		},
	}
	// AccountFlags are provided to the create account command
	AccountFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  "secret",
			Usage: "The account secret (password)",
		},
		&cli.StringSliceFlag{
			Name:  "scopes",
			Usage: "Comma seperated list of scopes to give the account",
		},
	}
)

// run the auth service
func Run(ctx *cli.Context, svcOpts ...vine.Option) {

	if len(ctx.String("address")) > 0 {
		Address = ctx.String("address")
	}
	if len(Address) > 0 {
		svcOpts = append(svcOpts, vine.Address(Address))
	}

	// Init plugins
	for _, p := range Plugins() {
		p.Init(ctx)
	}

	// setup the handlers
	ruleH := &rulesHandler.Rules{}
	authH := &authHandler.Auth{}

	// setup the auth handler to use JWTs
	pubKey := ctx.String("auth-public-key")
	privKey := ctx.String("auth-private-key")
	if len(pubKey) > 0 || len(privKey) > 0 {
		authH.TokenProvider = jwt.NewTokenProvider(
			token.WithPublicKey(pubKey),
			token.WithPrivateKey(privKey),
		)
	}

	st := *cmd.DefaultCmd.Options().Store

	// set the handlers store
	authH.Init(auth.Store(st))
	ruleH.Init(auth.Store(st))

	// setup service
	svcOpts = append(svcOpts, vine.Name(Name))
	service := vine.NewService(svcOpts...)

	// register handlers
	pb.RegisterAuthHandler(service.Server(), authH)
	pb.RegisterRulesHandler(service.Server(), ruleH)
	pb.RegisterAccountsHandler(service.Server(), authH)

	// run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}

func authFromContext(ctx *cli.Context) auth.Auth {
	if cliutil.IsLocal(ctx) {
		return *cmd.DefaultCmd.Options().Auth
	}
	return svcAuth.NewAuth(
		auth.WithClient(client.New(ctx)),
	)
}

// login using a token
func login(ctx *cli.Context) {
	// check for the token flag
	env := cliutil.GetEnv(ctx)
	if tok := ctx.String("token"); len(tok) > 0 {
		_, err := authFromContext(ctx).Inspect(tok)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := config.Set(tok, "vine", "auth", env.Name, "token"); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("You have been logged in")
		return
	}

	if ctx.Args().Len() != 2 {
		fmt.Println("Usage: `vine login {id} {secret} OR vine login --token {token}`")
		os.Exit(1)
	}
	id := ctx.Args().Get(0)
	secret := ctx.Args().Get(1)

	// Execute the request
	tok, err := authFromContext(ctx).Token(auth.WithCredentials(id, secret), auth.WithExpiry(time.Hour*24))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Store the access token in vine config
	if err := config.Set(tok.AccessToken, "vine", "auth", env.Name, "token"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Store the refresh token in vine config
	if err := config.Set(tok.RefreshToken, "vine", "auth", env.Name, "refresh-token"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Inform the user
	fmt.Println("You have been logged in")
}

// whoami returns info about the logged in user
func whoami(ctx *cli.Context) {
	// Get the token from vine config
	env, _ := config.Get("env")
	tok, err := config.Get("vine", "auth", env, "token")
	if err != nil {
		fmt.Println("You are not logged in")
		os.Exit(1)
	}

	// Inspect the token
	acc, err := authFromContext(ctx).Inspect(tok)
	if verr, ok := err.(*errors.Error); ok {
		fmt.Println("Error: " + verr.Detail)
		return
	} else if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("ID: %v; Scopes: %v\n", acc.ID, strings.Join(acc.Scopes, ", "))
}

//Commands for auth
func Commands(svcOpts ...vine.Option) []*cli.Command {
	commands := []*cli.Command{
		{
			Name:  "auth",
			Usage: "Manage authentication related resources",
			Action: func(ctx *cli.Context) error {
				if err := helper.UnexpectedSubcommand(ctx); err != nil {
					return err
				}
				Run(ctx)
				return nil
			},
			Subcommands: append([]*cli.Command{
				{
					Name:  "list",
					Usage: "List auth resources",
					Subcommands: append([]*cli.Command{
						{
							Name:  "rules",
							Usage: "List auth rules",
							Action: func(ctx *cli.Context) error {
								listRules(ctx)
								return nil
							},
						},
						{
							Name:  "accounts",
							Usage: "List auth accounts",
							Action: func(ctx *cli.Context) error {
								listAccounts(ctx)
								return nil
							},
						},
					}),
				},
				{
					Name:  "create",
					Usage: "Create an auth resource",
					Subcommands: append([]*cli.Command{
						{
							Name:  "rule",
							Usage: "Create an auth rule",
							Flags: append(RuleFlags),
							Action: func(ctx *cli.Context) error {
								createRule(ctx)
								return nil
							},
						},
						{
							Name:  "account",
							Usage: "Create an auth account",
							Flags: append(AccountFlags),
							Action: func(ctx *cli.Context) error {
								createAccount(ctx)
								return nil
							},
						},
					}),
				},
				{
					Name:  "delete",
					Usage: "Delete a auth resource",
					Subcommands: append([]*cli.Command{
						{
							Name:  "rule",
							Usage: "Delete an auth rule",
							Flags: RuleFlags,
							Action: func(ctx *cli.Context) error {
								deleteRule(ctx)
								return nil
							},
						},
					}),
				},
				{
					Name:        "api",
					Usage:       "Run the auth api",
					Description: "Run the auth api",
					Flags:       ServiceFlags,
					Action: func(ctx *cli.Context) error {
						api.Run(ctx, svcOpts...)
						return nil
					},
				},
			}),
		},
		{
			Name:  "login",
			Usage: "Login using a token",
			Action: func(ctx *cli.Context) error {
				login(ctx)
				return nil
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "token",
					Usage: "The token to set",
				},
			},
		},
		{
			Name:  "whoami",
			Usage: "Account information",
			Action: func(ctx *cli.Context) error {
				whoami(ctx)
				return nil
			},
		},
	}

	for _, c := range commands {
		for _, p := range Plugins() {
			if cmds := p.Commands(); len(cmds) > 0 {
				c.Subcommands = append(c.Subcommands, cmds...)
			}

			if flags := p.Flags(); len(flags) > 0 {
				c.Flags = append(c.Flags, flags...)
			}
		}
	}

	return commands
}
