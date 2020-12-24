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

package oauth

import (
	"fmt"
	"net/url"

	"github.com/lack-io/vine/util/auth/provider"
)

// NewProvider returns an initialised oauth provider
func NewProvider(opts ...provider.Option) provider.Provider {
	var options provider.Options
	for _, o := range opts {
		o(&options)
	}
	return &oauth{options}
}

type oauth struct {
	opts provider.Options
}

func (o *oauth) String() string {
	return "oauth"
}

func (o *oauth) Options() provider.Options {
	return o.opts
}

func (o *oauth) Endpoint(opts ...provider.EndpointOption) string {
	var options provider.EndpointOptions
	for _, o := range opts {
		o(&options)
	}

	params := make(url.Values)
	params.Add("response_type", "code")

	if len(options.State) > 0 {
		params.Add("state", options.State)
	}

	if len(options.LoginHint) > 0 {
		params.Add("login_hint", options.LoginHint)
	}

	if clientID := o.opts.ClientID; len(clientID) > 0 {
		params.Add("client_id", clientID)
	}

	if scope := o.opts.Scope; len(scope) > 0 {
		params.Add("scope", scope)
	}

	if redir := o.Redirect(); len(redir) > 0 {
		params.Add("redirect_uri", redir)
	}

	return fmt.Sprintf("%v?%v", o.opts.Endpoint, params.Encode())
}

func (o *oauth) Redirect() string {
	return o.opts.Redirect
}
