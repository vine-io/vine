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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/lack-io/cli"

	pb "github.com/lack-io/vine/proto/auth"
	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/util/client"
)

func listAccounts(ctx *cli.Context) {
	client := accountsFromContext(ctx)

	rsp, err := client.List(context.TODO(), &pb.ListAccountsRequest{})
	if err != nil {
		fmt.Printf("Error listing accounts: %v\n", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	defer w.Flush()

	fmt.Fprintln(w, strings.Join([]string{"ID", "Scopes", "Metadata"}, "\t\t"))
	for _, r := range rsp.Accounts {
		var metadata string
		for k, v := range r.Metadata {
			metadata = fmt.Sprintf("%v %v=%v ", metadata, k, v)
		}
		scopes := strings.Join(r.Scopes, ", ")

		if len(metadata) == 0 {
			metadata = "n/a"
		}
		if len(scopes) == 0 {
			scopes = "n/a"
		}

		fmt.Fprintln(w, strings.Join([]string{r.Id, scopes, metadata}, "\t\t"))
	}
}

func createAccount(ctx *cli.Context) {
	if ctx.Args().Len() == 0 {
		fmt.Println("Missing argument: ID")
		return
	}

	var options []auth.GenerateOption
	if len(ctx.StringSlice("scopes")) > 0 {
		options = append(options, auth.WithScopes(ctx.StringSlice("scopes")...))
	}
	if len(ctx.String("secret")) > 0 {
		options = append(options, auth.WithSecret(ctx.String("secret")))
	}

	acc, err := authFromContext(ctx).Generate(ctx.Args().First(), options...)
	if err != nil {
		fmt.Printf("Error creating account: %v\n", err)
		os.Exit(1)
	}

	json, _ := json.Marshal(acc)
	fmt.Printf("Account created: %v\n", string(json))
}

func accountsFromContext(ctx *cli.Context) pb.AccountsService {
	return pb.NewAccountsService("go.vine.auth", client.New(ctx))
}
