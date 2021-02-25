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

package manager

import "github.com/lack-io/vine/service/store"

// Options for the runtime manager
type Options struct {
	// Profile contains the env vars to set when
	// running a service
	Profile []string
	// Store to persist state
	Store store.Store
}

// Option sets an option
type Option func(*Options)

// Profile to use when running services
func Profile(p []string) Option {
	return func(o *Options) {
		o.Profile = p
	}
}

// Store to persist services and sync events
func Store(s store.Store) Option {
	return func(o *Options) {
		o.Store = s
	}
}
