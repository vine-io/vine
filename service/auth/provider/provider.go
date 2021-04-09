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

// Package provider is an external auth provider e.g oauth
package provider

import (
	"time"
)

// Provider is an auth provider
type Provider interface {
	// String returns the name of the provider
	String() string
	// Options returns the options of a provider
	Options() Options
	// Endpoint for the provider
	Endpoint(...EndpointOption) string
	// Redirect url incase of UI
	Redirect() string
}

// Grant is a granted authorisation
type Grant struct {
	// token for reuse
	Token string
	// Expiry of the token
	Expiry time.Time
	// Scopes associated with grant
	Scopes []string
}

type EndpointOptions struct {
	// State is a code to verify the req
	State string
	// LoginHint prefils the user id on oauth clients
	LoginHint string
}

type EndpointOption func(*EndpointOptions)

func WithState(c string) EndpointOption {
	return func(o *EndpointOptions) {
		o.State = c
	}
}

func WithLoginHint(hint string) EndpointOption {
	return func(o *EndpointOptions) {
		o.LoginHint = hint
	}
}
