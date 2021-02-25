// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package source

import (
	"context"

	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/config/encoder"
	"github.com/lack-io/vine/service/config/encoder/json"
)

type Options struct {
	// Encoder
	Encoder encoder.Encoder

	// for alternative data
	Context context.Context

	// Client to use for RPC
	Client client.Client
}

type Option func(o *Options)

func NewOptions(opts ...Option) Options {
	options := Options{
		Encoder: json.NewEncoder(),
		Context: context.Background(),
		Client:  client.DefaultClient,
	}

	for _, o := range opts {
		o(&options)
	}

	return options
}

// WithEncoder sets the source encoder
func WithEncoder(e encoder.Encoder) Option {
	return func(o *Options) {
		o.Encoder = e
	}
}

// WithClient sets the source client
func WithClient(c client.Client) Option {
	return func(o *Options) {
		o.Client = c
	}
}
