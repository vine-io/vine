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

package client

import (
	"context"

	ccli "github.com/lack-io/cli"

	cliutil "github.com/lack-io/vine/cmd/vine/client/cli/util"
	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/client/grpc"
	"github.com/lack-io/vine/util/config"
	"github.com/lack-io/vine/util/context/metadata"
)

// New returns a wrapped grpc client which will inject the
// token found in config into each request
func New(ctx *ccli.Context) client.Client {
	env := cliutil.GetEnv(ctx)
	token, _ := config.Get("vine", "auth", env.Name, "token")
	return &wrapper{grpc.NewClient(), token, env.Name, ctx}
}

type wrapper struct {
	client.Client
	token string
	env   string
	ctx   *ccli.Context
}

func (a *wrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
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
