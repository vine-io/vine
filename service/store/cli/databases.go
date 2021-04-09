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

package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/lack-io/cli"

	storeproto "github.com/lack-io/vine/proto/services/store"
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
