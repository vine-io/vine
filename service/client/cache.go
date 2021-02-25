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

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"time"

	cache "github.com/patrickmn/go-cache"

	"github.com/lack-io/vine/util/context/metadata"
)

// NewCache returns an initialised cache.
func NewCache() *Cache {
	return &Cache{
		cache: cache.New(cache.NoExpiration, 30*time.Second),
	}
}

// Cache for responses
type Cache struct {
	cache *cache.Cache
}

// Get a response from the cache
func (c *Cache) Get(ctx context.Context, req *Request) (interface{}, bool) {
	return c.cache.Get(key(ctx, req))
}

// Set a response in the cache
func (c *Cache) Set(ctx context.Context, req *Request, rsp interface{}, expiry time.Duration) {
	c.cache.Set(key(ctx, req), rsp, expiry)
}

// List the key value pairs in the cache
func (c *Cache) List() map[string]string {
	items := c.cache.Items()

	rsp := make(map[string]string, len(items))
	for k, v := range items {
		bytes, _ := json.Marshal(v.Object)
		rsp[k] = string(bytes)
	}

	return rsp
}

// key returns a hash for the context and request
func key(ctx context.Context, req *Request) string {
	ns, _ := metadata.Get(ctx, "Vine-Namespace")

	bytes, _ := json.Marshal(map[string]interface{}{
		"namespace": ns,
		"request": map[string]interface{}{
			"service":  (*req).Service(),
			"endpoint": (*req).Endpoint(),
			"method":   (*req).Method(),
			"body":     (*req).Body(),
		},
	})

	h := fnv.New64()
	h.Write(bytes)
	return fmt.Sprintf("%x", h.Sum(nil))
}
