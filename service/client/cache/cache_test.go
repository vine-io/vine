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

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/lack-io/vine/service/context/metadata"
)

func TestCache(t *testing.T) {
	ctx := context.TODO()
	req := &testRequest{service: "go.vine.service.foo", method: "Foo.Bar"}

	t.Run("CacheMiss", func(t *testing.T) {
		if _, ok := New().Get(ctx, req); ok {
			t.Errorf("Expected to get no result from Get")
		}
	})

	t.Run("CacheHit", func(t *testing.T) {
		c := New()

		rsp := "theresponse"
		c.Set(ctx, req, rsp, time.Minute)

		if res, ok := c.Get(ctx, req); !ok {
			t.Errorf("Expected a result, got nothing")
		} else if res != rsp {
			t.Errorf("Expected '%v' result, got '%v'", rsp, res)
		}
	})
}

func TestCacheKey(t *testing.T) {
	ctx := context.TODO()

	req1 := &testRequest{service: "go.vine.service.foo", method: "Foo.Bar"}
	req2 := &testRequest{service: "go.vine.service.foo", method: "Foo.Baz"}
	req3 := &testRequest{service: "go.vine.service.foo", method: "Foo.Bar", body: "customquery"}

	t.Run("IdenticalRequests", func(t *testing.T) {
		key1 := key(ctx, req1)
		key2 := key(ctx, req1)
		if key1 != key2 {
			t.Errorf("Expected the keys to match for identical requests and context")
		}
	})

	t.Run("DifferentRequestEndpoints", func(t *testing.T) {
		key1 := key(ctx, req1)
		key2 := key(ctx, req2)

		if key1 == key2 {
			t.Errorf("Expected the keys to differ for different request endpoints")
		}
	})

	t.Run("DifferentRequestBody", func(t *testing.T) {
		key1 := key(ctx, req2)
		key2 := key(ctx, req3)

		if key1 == key2 {
			t.Errorf("Expected the keys to differ for different request bodies")
		}
	})

	t.Run("DifferentMetadata", func(t *testing.T) {
		mdCtx := metadata.Set(context.TODO(), "Vine-Namespace", "bar")
		key1 := key(mdCtx, req1)
		key2 := key(ctx, req1)

		if key1 == key2 {
			t.Errorf("Expected the keys to differ for different metadata")
		}
	})
}
