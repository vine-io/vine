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

package token

import (
	"time"

	"github.com/lack-io/vine/service/dao"
)

type Options struct {
	// Dialect to persist the tokens
	Dialect dao.Dialect
	// PublicKey base64 encoded, used by JWT
	PublicKey string
	// PrivateKey base64 encoded, used by JWT
	PrivateKey string
}

type Option func(o *Options)

// WithDialect sets the token providers store
func WithDialect(d dao.Dialect) Option {
	return func(o *Options) {
		o.Dialect = d
	}
}

// WithPublicKey sets the JWT public key
func WithPublicKey(key string) Option {
	return func(o *Options) {
		o.PublicKey = key
	}
}

// WithPrivateKey sets the JWT private key
func WithPrivateKey(key string) Option {
	return func(o *Options) {
		o.PrivateKey = key
	}
}

func NewOptions(opts ...Option) Options {
	var options Options
	for _, o := range opts {
		o(&options)
	}
	//set default store
	if options.Dialect == nil {
		options.Dialect = dao.DefaultDialect
	}
	return options
}

type GenerateOptions struct {
	// Expiry for the token
	Expiry time.Duration
}

type GenerateOption func(o *GenerateOptions)

// WithExpiry for the generated account's token expires
func WithExpiry(d time.Duration) GenerateOption {
	return func(o *GenerateOptions) {
		o.Expiry = d
	}
}

// NewGenerateOptions from a slice of options
func NewGenerateOptions(opts ...GenerateOption) GenerateOptions {
	var options GenerateOptions
	for _, o := range opts {
		o(&options)
	}
	//set default Expiry of token
	if options.Expiry == 0 {
		options.Expiry = time.Minute * 15
	}
	return options
}
