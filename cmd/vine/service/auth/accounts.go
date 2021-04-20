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

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/lib/auth"
	pb "github.com/lack-io/vine/proto/services/auth"
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
