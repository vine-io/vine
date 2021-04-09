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
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/proto/apis/errors"
	pb "github.com/lack-io/vine/proto/services/auth"
	"github.com/lack-io/vine/util/client"
)

func listRules(ctx *cli.Context) {
	client := rulesFromContext(ctx)

	rsp, err := client.List(context.TODO(), &pb.ListRequest{})
	if err != nil {
		fmt.Printf("Error listing rules: %v\n", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	defer w.Flush()

	formatResource := func(r *pb.Resource) string {
		return strings.Join([]string{r.Type, r.Name, r.Endpoint}, ":")
	}

	// sort rules using resource name and priority to keep the list consistent
	sort.Slice(rsp.Rules, func(i, j int) bool {
		resI := formatResource(rsp.Rules[i].Resource) + string(rsp.Rules[i].Priority)
		resJ := formatResource(rsp.Rules[j].Resource) + string(rsp.Rules[j].Priority)
		return sort.StringsAreSorted([]string{resJ, resI})
	})

	fmt.Fprintln(w, strings.Join([]string{"ID", "Scope", "Access", "Resource", "Priority"}, "\t\t"))
	for _, r := range rsp.Rules {
		res := formatResource(r.Resource)
		if r.Scope == "" {
			r.Scope = "<public>"
		}
		fmt.Fprintln(w, strings.Join([]string{r.Id, r.Scope, r.Access.String(), res, fmt.Sprintf("%d", r.Priority)}, "\t\t"))
	}
}

func createRule(ctx *cli.Context) {
	_, err := rulesFromContext(ctx).Create(context.TODO(), &pb.CreateRequest{
		Rule: constructRule(ctx),
	})
	if verr, ok := err.(*errors.Error); ok {
		fmt.Printf("Error: %v\n", verr.Detail)
		return
	} else if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Rule created")
}

func deleteRule(ctx *cli.Context) {
	if ctx.Args().Len() != 1 {
		fmt.Println("Expected one argument: ID")
		os.Exit(1)
	}

	_, err := rulesFromContext(ctx).Delete(context.TODO(), &pb.DeleteRequest{
		Id: ctx.Args().First(),
	})
	if verr, ok := err.(*errors.Error); ok {
		fmt.Printf("Error: %v\n", verr.Detail)
		return
	} else if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Rule deleted")
}

func constructRule(ctx *cli.Context) *pb.Rule {
	if ctx.Args().Len() != 1 {
		fmt.Println("Too many arguments, expected one argument: ID")
		os.Exit(1)
	}

	var access pb.Access
	switch ctx.String("access") {
	case "granted":
		access = pb.Access_GRANTED
	case "denied":
		access = pb.Access_DENIED
	default:
		fmt.Printf("Invalid access: %v, must be granted or denied\n", ctx.String("access"))
		os.Exit(1)
	}

	resComps := strings.Split(ctx.String("resource"), ":")
	if len(resComps) != 3 {
		fmt.Println("Invalid resource, must be in the format type:name:endpoint")
		os.Exit(1)
	}

	return &pb.Rule{
		Id:       ctx.Args().First(),
		Access:   access,
		Scope:    ctx.String("scope"),
		Priority: int32(ctx.Int("priority")),
		Resource: &pb.Resource{
			Type:     resComps[0],
			Name:     resComps[1],
			Endpoint: resComps[2],
		},
	}
}

func rulesFromContext(ctx *cli.Context) pb.RulesService {
	return pb.NewRulesService("go.vine.auth", client.New(ctx))
}
