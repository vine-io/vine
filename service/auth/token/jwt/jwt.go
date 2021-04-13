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
	"encoding/base64"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/service/auth/token"
)

// authClaims to be encoded in the JWT
type authClaims struct {
	Type     auth.Type         `json:"type"`
	Scopes   []string          `json:"scopes"`
	Metadata map[string]string `json:"metadata"`

	jwt.StandardClaims
}

// JWT implementation of token provider
type JWT struct {
	opts token.Options
}

// NewTokenProvider returns an initialized basic provider
func NewTokenProvider(opts ...token.Option) token.Provider {
	return &JWT{
		opts: token.NewOptions(opts...),
	}
}

// Generate a new JWT
func (j *JWT) Generate(acc *auth.Account, opts ...token.GenerateOption) (*token.Token, error) {
	// decode the private key
	priv, err := base64.StdEncoding.DecodeString(j.opts.PrivateKey)
	if err != nil {
		return nil, err
	}

	// parse the private key
	key, err := jwt.ParseRSAPrivateKeyFromPEM(priv)
	if err != nil {
		return nil, token.ErrEncodingToken
	}

	// parse the options
	options := token.NewGenerateOptions(opts...)

	// generate the JWT
	expiry := time.Now().Add(options.Expiry)
	t := jwt.NewWithClaims(jwt.SigningMethodRS256, authClaims{
		acc.Type, acc.Scopes, acc.Metadata, jwt.StandardClaims{
			Subject:   acc.ID,
			Issuer:    acc.Issuer,
			ExpiresAt: expiry.Unix(),
		},
	})
	tok, err := t.SignedString(key)
	if err != nil {
		return nil, err
	}

	// return the token
	return &token.Token{
		Token:   tok,
		Expiry:  expiry,
		Created: time.Now(),
	}, nil
}

// Inspect a JWT
func (j *JWT) Inspect(t string) (*auth.Account, error) {
	// decode the public key
	pub, err := base64.StdEncoding.DecodeString(j.opts.PublicKey)
	if err != nil {
		return nil, err
	}

	// parse the public key
	res, err := jwt.ParseWithClaims(t, &authClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwt.ParseRSAPublicKeyFromPEM(pub)
	})
	if err != nil {
		return nil, token.ErrInvalidToken
	}

	// validate the token
	if !res.Valid {
		return nil, token.ErrInvalidToken
	}
	claims, ok := res.Claims.(*authClaims)
	if !ok {
		return nil, token.ErrInvalidToken
	}

	// return the token
	return &auth.Account{
		ID:       claims.Subject,
		Issuer:   claims.Issuer,
		Type:     claims.Type,
		Scopes:   claims.Scopes,
		Metadata: claims.Metadata,
	}, nil
}

// String returns JWT
func (j *JWT) String() string {
	return "jwt"
}
