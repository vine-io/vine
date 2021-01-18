// Copyright 2020 lack
//
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

package env

import (
	"context"
	"strings"

	"github.com/lack-io/vine/service/config/source"
)

type strippedPrefixKey struct{}
type prefixKey struct{}

// WithStrippedPrefix sets the environment variable prefixes to scope to.
// These prefixes will be removed from the actual config entries.
func WithStrippedPrefix(p ...string) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}

		o.Context = context.WithValue(o.Context, strippedPrefixKey{}, appendUnderscore(p))
	}
}

// WithPrefix sets the environment variable prefixes to scope to.
// These prefixes will not be removed. Each prefix will be considered a top level config entry.
func WithPrefix(p ...string) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, prefixKey{}, appendUnderscore(p))
	}
}

func appendUnderscore(prefixes []string) []string {
	//nolint:prealloc
	var result []string
	for _, p := range prefixes {
		if !strings.HasSuffix(p, "_") {
			result = append(result, p+"_")
			continue
		}

		result = append(result, p)
	}

	return result
}
