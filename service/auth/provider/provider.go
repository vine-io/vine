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
