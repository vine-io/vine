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

package basic

import (
	"time"

	"github.com/google/uuid"
	"github.com/lack-io/vine/lib/dao"

	"github.com/lack-io/vine/lib/auth"
	"github.com/lack-io/vine/lib/auth/token"
)

// Basic implementation of token provider, backed by the store
type Basic struct {
	dialect dao.Dialect
}

var (
	// StorePrefix to isolate tokens
	StorePrefix = "tokens/"
)

// NewTokenProvider returns an initialized basic provider
func NewTokenProvider(opts ...token.Option) token.Provider {
	options := token.NewOptions(opts...)

	//if options.Dialect == nil {
	//	options.Dialect = sqlite.NewDialect()
	//}

	return &Basic{
		dialect: options.Dialect,
	}
}

// Generate a token for an account
func (b *Basic) Generate(acc *auth.Account, opts ...token.GenerateOption) (*token.Token, error) {
	options := token.NewGenerateOptions(opts...)

	// marshal the account to bytes
	//bytes, err := json.Marshal(acc)
	//if err != nil {
	//	return nil, err
	//}

	// write to the store
	key := uuid.New().String()
	//err = b.store.Write(&store.Record{
	//	Key:    fmt.Sprintf("%v%v", StorePrefix, key),
	//	Value:  bytes,
	//	Expiry: options.Expiry,
	//})
	//if err != nil {
	//	return nil, err
	//}

	// return the token
	return &token.Token{
		Token:   key,
		Created: time.Now(),
		Expiry:  time.Now().Add(options.Expiry),
	}, nil
}

// Inspect a token
func (b *Basic) Inspect(t string) (*auth.Account, error) {
	// lookup the token in the store
	//recs, err := b.store.Read(StorePrefix + t)
	//if err == store.ErrNotFound {
	//	return nil, token.ErrInvalidToken
	//} else if err != nil {
	//	return nil, err
	//}
	//bytes := recs[0].Value
	//
	//// unmarshal the bytes
	//var acc *auth.Account
	//if err := json.Unmarshal(bytes, &acc); err != nil {
	//	return nil, err
	//}

	//return acc, nil
	return nil, nil
}

// String returns basic
func (b *Basic) String() string {
	return "basic"
}
