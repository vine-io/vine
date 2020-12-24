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

package provider

// Option returns a function which sets an option
type Option func(*Options)

// Options a provider can have
type Options struct {
	// ClientID is the application's ID.
	ClientID string
	// ClientSecret is the application's secret.
	ClientSecret string
	// Endpoint for the provider
	Endpoint string
	// Redirect url incase of UI
	Redirect string
	// Scope of the oauth request
	Scope string
}

// Credentials is an option which sets the client id and secret
func Credentials(id, secret string) Option {
	return func(o *Options) {
		o.ClientID = id
		o.ClientSecret = secret
	}
}

// Endpoint sets the endpoint option
func Endpoint(e string) Option {
	return func(o *Options) {
		o.Endpoint = e
	}
}

// Redirect sets the Redirect option
func Redirect(r string) Option {
	return func(o *Options) {
		o.Redirect = r
	}
}

// Scope sets the oauth scope
func Scope(s string) Option {
	return func(o *Options) {
		o.Scope = s
	}
}
