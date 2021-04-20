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
	"errors"
	"time"

	"github.com/lack-io/vine/lib/auth"
)

var (
	// ErrNotFound is returned when a token cannot be found
	ErrNotFound = errors.New("token not found")
	// ErrEncodingToken is returned when the service encounters an error during encoding
	ErrEncodingToken = errors.New("error encoding the token")
	// ErrInvalidToken is returned when the token provided is not valid
	ErrInvalidToken = errors.New("invalid token provided")
)

// Provider generates and inspects tokens
type Provider interface {
	Generate(account *auth.Account, opts ...GenerateOption) (*Token, error)
	Inspect(token string) (*auth.Account, error)
	String() string
}

type Token struct {
	// The actual token
	Token string `json:"token"`
	// Time of token creation
	Created time.Time `json:"created"`
	// Time of token expiry
	Expiry time.Time `json:"expiry"`
}
