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

package basic

import (
	"github.com/lack-io/vine/service/auth/provider"
)

// NewProvider returns an initialised basic provider
func NewProvider(opts ...provider.Option) provider.Provider {
	var options provider.Options
	for _, o := range opts {
		o(&options)
	}
	return &basic{options}
}

type basic struct {
	opts provider.Options
}

func (b *basic) String() string {
	return "basic"
}

func (b *basic) Options() provider.Options {
	return b.opts
}

func (b *basic) Endpoint(...provider.EndpointOption) string {
	return ""
}

func (b *basic) Redirect() string {
	return ""
}
