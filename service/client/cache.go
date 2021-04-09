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
