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

package oauth

import (
	"fmt"
	"net/url"

	"github.com/lack-io/vine/service/auth/provider"
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
