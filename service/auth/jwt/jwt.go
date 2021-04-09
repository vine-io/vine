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

package jwt

import (
	"sync"
	"time"

	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/service/auth/rules"
	"github.com/lack-io/vine/service/auth/token"
	jwtToken "github.com/lack-io/vine/service/auth/token/jwt"
)

// NewAuth returns a new instance of the Auth service
func NewAuth(opts ...auth.Option) auth.Auth {
	j := new(jwt)
	j.Init(opts...)
	return j
}

type jwt struct {
	options auth.Options
	jwt     token.Provider
	rules   []*auth.Rule

	sync.Mutex
}

func (j *jwt) String() string {
	return "jwt"
}

func (j *jwt) Init(opts ...auth.Option) {
	j.Lock()
	defer j.Unlock()

	for _, o := range opts {
		o(&j.options)
	}

	j.jwt = jwtToken.NewTokenProvider(
		token.WithPrivateKey(j.options.PrivateKey),
		token.WithPublicKey(j.options.PublicKey),
	)
}

func (j *jwt) Options() auth.Options {
	j.Lock()
	defer j.Unlock()
	return j.options
}

func (j *jwt) Generate(id string, opts ...auth.GenerateOption) (*auth.Account, error) {
	options := auth.NewGenerateOptions(opts...)
	account := &auth.Account{
		ID:       id,
		Type:     options.Type,
		Scopes:   options.Scopes,
		Metadata: options.Metadata,
		Issuer:   j.Options().Namespace,
	}

	// generate a JWT secret which can be provided to the Token() method
	// and exchanged for an access token
	secret, err := j.jwt.Generate(account)
	if err != nil {
		return nil, err
	}
	account.Secret = secret.Token

	// return the account
	return account, nil
}

func (j *jwt) Grant(rule *auth.Rule) error {
	j.Lock()
	defer j.Unlock()
	j.rules = append(j.rules, rule)
	return nil
}

func (j *jwt) Revoke(rule *auth.Rule) error {
	j.Lock()
	defer j.Unlock()

	rules := []*auth.Rule{}
	for _, r := range j.rules {
		if r.ID != rule.ID {
			rules = append(rules, r)
		}
	}

	j.rules = rules
	return nil
}

func (j *jwt) Verify(acc *auth.Account, res *auth.Resource, opts ...auth.VerifyOption) error {
	j.Lock()
	defer j.Unlock()

	var options auth.VerifyOptions
	for _, o := range opts {
		o(&options)
	}

	return rules.Verify(j.rules, acc, res)
}

func (j *jwt) Rules(opts ...auth.RulesOption) ([]*auth.Rule, error) {
	j.Lock()
	defer j.Unlock()
	return j.rules, nil
}

func (j *jwt) Inspect(token string) (*auth.Account, error) {
	return j.jwt.Inspect(token)
}

func (j *jwt) Token(opts ...auth.TokenOption) (*auth.Token, error) {
	options := auth.NewTokenOptions(opts...)

	secret := options.RefreshToken
	if len(options.Secret) > 0 {
		secret = options.Secret
	}

	account, err := j.jwt.Inspect(secret)
	if err != nil {
		return nil, err
	}

	access, err := j.jwt.Generate(account, token.WithExpiry(options.Expiry))
	if err != nil {
		return nil, err
	}

	refresh, err := j.jwt.Generate(account, token.WithExpiry(options.Expiry+time.Hour))
	if err != nil {
		return nil, err
	}

	return &auth.Token{
		Created:      access.Created,
		Expiry:       access.Expiry,
		AccessToken:  access.Token,
		RefreshToken: refresh.Token,
	}, nil
}
