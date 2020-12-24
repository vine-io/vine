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

package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/lack-io/cli"

	storeproto "github.com/lack-io/vine/proto/store"
	"github.com/lack-io/vine/service/config/cmd"
)

// Databases is the entrypoint for vine store databases
func Databases(ctx *cli.Context) error {
	client := *cmd.DefaultOptions().Client
	dbReq := client.NewRequest(ctx.String("store"), "Store.Databases", &storeproto.DatabasesRequest{})
	dbRsp := &storeproto.DatabasesResponse{}
	if err := client.Call(context.TODO(), dbReq, dbRsp); err != nil {
		return err
	}
	for _, db := range dbRsp.Databases {
		fmt.Println(db)
	}
	return nil
}

// Tables is the entrypoint for vine store tables
func Tables(ctx *cli.Context) error {
	if len(ctx.String("database")) == 0 {
		return errors.New("database flag is required")
	}
	client := *cmd.DefaultOptions().Client
	tReq := client.NewRequest(ctx.String("store"), "Store.Tables", &storeproto.TablesRequest{
		Database: ctx.String("database"),
	})
	tRsp := &storeproto.TablesResponse{}
	if err := client.Call(context.TODO(), tReq, tRsp); err != nil {
		return err
	}
	for _, table := range tRsp.Tables {
		fmt.Println(table)
	}
	return nil
}
