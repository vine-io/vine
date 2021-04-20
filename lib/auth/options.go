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
	"time"

	"github.com/lack-io/vine/core/client"
	"github.com/lack-io/vine/lib/auth/provider"
	"github.com/lack-io/vine/lib/dao"
)

func NewOptions(opts ...Option) Options {
	var options Options
	for _, o := range opts {
		o(&options)
	}
	if options.Client == nil {
		options.Client = client.DefaultClient
	}

	return options
}

type Options struct {
	// Namespace the service belongs to
	Namespace string
	// ID is the services auth ID
	ID string
	// Secret is used to authenticate the service
	Secret string
	// Token is the services token used to authenticate itself
	Token *Token
	// PublicKey for decoding JWTs
	PublicKey string
	// PrivateKey for encoding JWTs
	PrivateKey string
	// Provider is an auth provider
	Provider provider.Provider
	// LoginURL is the relative url path where a user can login
	LoginURL string
	// Dialect to back auth
	Dialect dao.Dialect
	// Client to use for RPC
	Client client.Client
	// Addrs sets the addresses of auth
	Addrs []string
}

type Option func(o *Options)

// Addrs is the auth addresses to use
func Addrs(addrs ...string) Option {
	return func(o *Options) {
		o.Addrs = addrs
	}
}

// Namespace the service belongs to
func Namespace(n string) Option {
	return func(o *Options) {
		o.Namespace = n
	}
}

// Dialect to back auth
func Dialect(d dao.Dialect) Option {
	return func(o *Options) {
		o.Dialect = d
	}
}

// PublicKey is the JWT public key
func PublicKey(key string) Option {
	return func(o *Options) {
		o.PublicKey = key
	}
}

// PrivateKey is the JWT private key
func PrivateKey(key string) Option {
	return func(o *Options) {
		o.PrivateKey = key
	}
}

// Credentials sets the auth credentials
func Credentials(id, secret string) Option {
	return func(o *Options) {
		o.ID = id
		o.Secret = secret
	}
}

// ClientToken sets the auth token to use when making requests
func ClientToken(token *Token) Option {
	return func(o *Options) {
		o.Token = token
	}
}

// Provider set the auth provider
func Provider(p provider.Provider) Option {
	return func(o *Options) {
		o.Provider = p
	}
}

// LoginURL sets the auth LoginURL
func LoginURL(url string) Option {
	return func(o *Options) {
		o.LoginURL = url
	}
}

// WithClient sets the client to use when making requests
func WithClient(c client.Client) Option {
	return func(o *Options) {
		o.Client = c
	}
}

type GenerateOptions struct {
	// Metadata associated with the account
	Metadata map[string]string
	// Scopes the account has access to
	Scopes []string
	// Provider of the account, e.g. oauth
	Provider string
	// Type of the account, e.g. user
	Type Type
	// Secret used to authenticate the account
	Secret string

	Context context.Context
}

type GenerateOption func(o *GenerateOptions)

// WithSecret for the generated account
func WithSecret(s string) GenerateOption {
	return func(o *GenerateOptions) {
		o.Secret = s
	}
}

// WithType for the generated account
func WithType(t Type) GenerateOption {
	return func(o *GenerateOptions) {
		o.Type = t
	}
}

// WithMetadata for the generated account
func WithMetadata(md map[string]string) GenerateOption {
	return func(o *GenerateOptions) {
		o.Metadata = md
	}
}

// WithProvider for the generated account
func WithProvider(p string) GenerateOption {
	return func(o *GenerateOptions) {
		o.Provider = p
	}
}

// WithScopes for the generated account
func WithScopes(s ...string) GenerateOption {
	return func(o *GenerateOptions) {
		o.Scopes = s
	}
}

// WithGenerateContext fot the generated account
func WithGenerateContext(ctx context.Context) GenerateOption {
	return func(o *GenerateOptions) {
		o.Context = ctx
	}
}

// NewGenerateOptions from a slice of options
func NewGenerateOptions(opts ...GenerateOption) GenerateOptions {
	var options GenerateOptions
	for _, o := range opts {
		o(&options)
	}

	if options.Context == nil {
		options.Context = context.Background()
	}

	return options
}

type TokenOptions struct {
	// ID for the account
	ID string
	// Secret for the account
	Secret string
	// RefreshToken is used to refresh a token
	RefreshToken string
	// Expiry is the time the token should live for
	Expiry time.Duration

	Context context.Context
}

type TokenOption func(o *TokenOptions)

// WithExpiry for the token
func WithExpiry(ex time.Duration) TokenOption {
	return func(o *TokenOptions) {
		o.Expiry = ex
	}
}

func WithCredentials(id, secret string) TokenOption {
	return func(o *TokenOptions) {
		o.ID = id
		o.Secret = secret
	}
}

func WithToken(rt string) TokenOption {
	return func(o *TokenOptions) {
		o.RefreshToken = rt
	}
}

func WithTokenContext(ctx context.Context) TokenOption {
	return func(o *TokenOptions) {
		o.Context = ctx
	}
}

// NewTokenOptions from a slice of options
func NewTokenOptions(opts ...TokenOption) TokenOptions {
	var options TokenOptions
	for _, o := range opts {
		o(&options)
	}

	// set defualt expiry of token
	if options.Expiry == 0 {
		options.Expiry = time.Minute
	}

	if options.Context == nil {
		options.Context = context.Background()
	}

	return options
}

type VerifyOptions struct {
	Context context.Context
}

type VerifyOption func(o *VerifyOptions)

func VerifyContext(ctx context.Context) VerifyOption {
	return func(o *VerifyOptions) {
		o.Context = ctx
	}
}

type RulesOptions struct {
	Context context.Context
}

type RulesOption func(o *RulesOptions)

func RulesContext(ctx context.Context) RulesOption {
	return func(o *RulesOptions) {
		o.Context = ctx
	}
}
