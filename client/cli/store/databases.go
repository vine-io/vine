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
	"fmt"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/client/cli/namespace"
	"github.com/lack-io/vine/client/cli/util"
	pb "github.com/lack-io/vine/proto/store"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/context"
)

// databases is the entrypoint for vine store databases
func databases(ctx *cli.Context) error {
	dbReq := client.NewRequest(ctx.String("store"), "Store.Databases", &pb.DatabasesRequest{})
	dbRsp := &pb.DatabasesResponse{}
	if err := client.DefaultClient.Call(context.DefaultContext, dbReq, dbRsp, client.WithAuthToken()); err != nil {
		return err
	}
	for _, db := range dbRsp.Databases {
		fmt.Println(db)
	}
	return nil
}

// tables is the entrypoint for vine store tables
func tables(ctx *cli.Context) error {
	env, err := util.GetEnv(ctx)
	if err != nil {
		return err
	}
	ns, err := namespace.Get(env.Name)
	if err != nil {
		return err
	}

	tReq := client.NewRequest(ctx.String("store"), "Store.Tables", &pb.TablesRequest{
		Database: ns,
	})
	tRsp := &pb.TablesResponse{}
	if err := client.DefaultClient.Call(context.DefaultContext, tReq, tRsp, client.WithAuthToken()); err != nil {
		return err
	}
	for _, table := range tRsp.Tables {
		fmt.Println(table)
	}
	return nil
}
