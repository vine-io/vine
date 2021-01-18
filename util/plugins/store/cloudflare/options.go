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

package cloudflare

import (
	"context"
	"time"

	"github.com/lack-io/vine/service/store"
)

func getOption(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}
	val, ok := ctx.Value(key).(string)
	if !ok {
		return ""
	}
	return val
}

func getToken(ctx context.Context) string {
	return getOption(ctx, "CF_API_TOKEN")
}

func getAccount(ctx context.Context) string {
	return getOption(ctx, "CF_ACCOUNT_ID")
}

// Token sets the cloudflare api token
func Token(t string) store.Option {
	return func(o *store.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, "CF_API_TOKEN", t)
	}
}

// Account sets the cloudflare account id
func Account(id string) store.Option {
	return func(o *store.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, "CF_ACCOUNT_ID", id)
	}
}

// Namespace sets the KV namespace
func Namespace(ns string) store.Option {
	return func(o *store.Options) {
		o.Database = ns
	}
}

// CacheTTL sets the timeout in nanoseconds of the read/write cache
func CacheTTL(ttl time.Duration) store.Option {
	return func(o *store.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, "STORE_CACHE_TTL", ttl)
	}
}
