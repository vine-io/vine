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

package client

import (
	"context"

	ccli "github.com/lack-io/cli"
	client2 "github.com/lack-io/vine/core/client"
	grpc2 "github.com/lack-io/vine/core/client/grpc"

	cliutil "github.com/lack-io/vine/cmd/vine/client/cli/util"
	"github.com/lack-io/vine/lib/auth"
	"github.com/lack-io/vine/util/config"
	"github.com/lack-io/vine/util/context/metadata"
)

// New returns a wrapped grpc client which will inject the
// token found in config into each request
func New(ctx *ccli.Context) client2.Client {
	env := cliutil.GetEnv(ctx)
	token, _ := config.Get("vine", "auth", env.Name, "token")
	return &wrapper{grpc2.NewClient(), token, env.Name, ctx}
}

type wrapper struct {
	client2.Client
	token string
	env   string
	ctx   *ccli.Context
}

func (a *wrapper) Call(ctx context.Context, req client2.Request, rsp interface{}, opts ...client2.CallOption) error {
	if len(a.token) > 0 {
		ctx = metadata.Set(ctx, "Authorization", auth.BearerScheme+a.token)
	}
	if len(a.env) > 0 && !cliutil.IsLocal(a.ctx) && !cliutil.IsServer(a.ctx) {
		// @todo this is temporarily removed because multi tenancy is not there yet
		// and the moment core and non core services run in different environments, we
		// get issues. To test after `vine env add mine 127.0.0.1:8081` do,
		//
		// env := strings.ReplaceAll(a.env, "/", "-")
		// ctx = metadata.Set(ctx, "Vine-Namespace", env)
	}
	return a.Client.Call(ctx, req, rsp, opts...)
}
