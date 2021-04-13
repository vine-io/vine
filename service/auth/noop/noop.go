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

package noop

import (
	"github.com/google/uuid"
	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/service/auth/provider/basic"
)

type noop struct {
	opts auth.Options
}

// String returns the name of the implementation
func (n *noop) String() string {
	return "noop"
}

// Init the auth
func (n *noop) Init(opts ...auth.Option) {
	for _, o := range opts {
		o(&n.opts)
	}
}

// Options set for auth
func (n *noop) Options() auth.Options {
	return n.opts
}

// Generate a new account
func (n *noop) Generate(id string, opts ...auth.GenerateOption) (*auth.Account, error) {
	options := auth.NewGenerateOptions(opts...)

	return &auth.Account{
		ID:       id,
		Secret:   options.Secret,
		Metadata: options.Metadata,
		Scopes:   options.Scopes,
		Issuer:   n.Options().Namespace,
	}, nil
}

// Grant access to a resource
func (n *noop) Grant(rule *auth.Rule) error {
	return nil
}

// Revoke access to a resource
func (n *noop) Revoke(rule *auth.Rule) error {
	return nil
}

// Rules used to verify requests
func (n *noop) Rules(opts ...auth.RulesOption) ([]*auth.Rule, error) {
	return []*auth.Rule{}, nil
}

// Verify an account has access to a resource
func (n *noop) Verify(acc *auth.Account, res *auth.Resource, opts ...auth.VerifyOption) error {
	return nil
}

// Inspect a token
func (n *noop) Inspect(token string) (*auth.Account, error) {
	return &auth.Account{ID: uuid.New().String(), Issuer: n.Options().Namespace}, nil
}

// Token generation using an account id and secret
func (n *noop) Token(opts ...auth.TokenOption) (*auth.Token, error) {
	return &auth.Token{}, nil
}

func NewAuth(opts ...auth.Option) auth.Auth {
	options := auth.Options{
		Provider: basic.NewProvider(),
	}

	for _, o := range opts {
		o(&options)
	}

	return &noop{
		opts: options,
	}
}
