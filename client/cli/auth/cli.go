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

package cli

import (
	"github.com/lack-io/cli"

	"github.com/lack-io/vine/cmd"
	"github.com/lack-io/vine/internal/helper"
)

var (
	// ruleFlags are provided to commands which create or delete rules
	ruleFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  "scope",
			Usage: "The scope to amend, e.g. 'user' or '*', leave blank to make public",
		},
		&cli.StringFlag{
			Name:  "resource",
			Usage: "The resource to amend in the format type:name:endpoint, e.g. service:auth:*",
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
	// accountFlags are provided to the create account command
	accountFlags = []cli.Flag{
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

func init() {
	cmd.Register(
		&cli.Command{
			Name:   "auth",
			Usage:  "Manage authentication, accounts and rules",
			Action: helper.UnexpectedSubcommand,
			Subcommands: []*cli.Command{
				{
					Name:  "list",
					Usage: "List auth resources",
					Subcommands: []*cli.Command{
						{
							Name:   "rules",
							Usage:  "List auth rules",
							Action: listRules,
						},
						{
							Name:   "accounts",
							Usage:  "List auth accounts",
							Action: listAccounts,
						},
					},
				},
				{
					Name:  "create",
					Usage: "Create an auth resource",
					Subcommands: []*cli.Command{
						{
							Name:   "rule",
							Usage:  "Create an auth rule",
							Flags:  ruleFlags,
							Action: createRule,
						},
						{
							Name:  "account",
							Usage: "Create an auth account",
							Flags: append(accountFlags, &cli.StringFlag{
								Name:  "namespace",
								Usage: "Namespace to use when creating the account",
							}),
							Action: createAccount,
						},
					},
				},
				{
					Name:  "delete",
					Usage: "Delete a auth resource",
					Subcommands: []*cli.Command{
						{
							Name:   "rule",
							Usage:  "Delete an auth rule",
							Flags:  ruleFlags,
							Action: deleteRule,
						},
						{
							Name:   "account",
							Usage:  "Delete an auth account",
							Flags:  ruleFlags,
							Action: deleteAccount,
						},
					},
				},
			},
		},
		&cli.Command{
			Name:        "login",
			Usage:       `Interactive login flow.`,
			Description: "Run 'vine login' for vine servers or 'vine login --otp' for the Vine Platform.",
			Action:      login,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "otp",
					Usage: "Login/signup with a One Time Password.",
				},
				&cli.StringFlag{
					Name:  "password",
					Usage: "Password to use for login. If not provided, will be asked for during login. Useful for automated scripts",
				},
				&cli.StringFlag{
					Name:    "username",
					Usage:   "Username to use for login",
					Aliases: []string{"email"},
				},
			},
		},
		&cli.Command{
			Name:        "logout",
			Usage:       `Logout.`,
			Description: "Use 'vine logout' to delete your token in your current environment.",
			Action:      logout,
		},
	)
}
