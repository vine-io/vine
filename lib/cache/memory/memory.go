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

// Package memory is a in-memory cache cache
package memory

import (
	"context"
	"errors"
	"path/filepath"
	"sort"
	"strings"
	"time"

	gcache "github.com/patrickmn/go-cache"

	"github.com/vine-io/vine/lib/cache"
)

// NewCache returns a memory cache
func NewCache(opts ...cache.Option) cache.Cache {
	s := &memoryCache{
		options: cache.Options{
			Database: "vine",
			Table:    "vine",
		},
		cache: gcache.New(gcache.NoExpiration, 5*time.Minute),
	}
	for _, o := range opts {
		o(&s.options)
	}
	return s
}

type memoryCache struct {
	options cache.Options

	cache *gcache.Cache
}

type cacheRecord struct {
	key       string
	value     []byte
	metadata  map[string]interface{}
	expiresAt time.Time
}

func (m *memoryCache) key(prefix, key string) string {
	return filepath.Join(prefix, key)
}

func (m *memoryCache) prefix(database, table string) string {
	if len(database) == 0 {
		database = m.options.Database
	}
	if len(table) == 0 {
		table = m.options.Table
	}
	return filepath.Join(database, table)
}

func (m *memoryCache) get(prefix, key string) (*cache.Record, error) {
	key = m.key(prefix, key)

	var cachedRecord *cacheRecord
	r, found := m.cache.Get(key)
	if !found {
		return nil, cache.ErrNotFound
	}

	cachedRecord, ok := r.(*cacheRecord)
	if !ok {
		return nil, errors.New("retrieved a non *cacheRecord from the cache")
	}

	// Copy the record on the way out
	newRecord := &cache.Record{}
	newRecord.Key = strings.TrimPrefix(cachedRecord.key, prefix+"/")
	newRecord.Value = make([]byte, len(cachedRecord.value))
	newRecord.Metadata = make(map[string]interface{})

	// copy the value into the new record
	copy(newRecord.Value, cachedRecord.value)

	// check if we need to set the expiry
	if !cachedRecord.expiresAt.IsZero() {
		newRecord.Expiry = time.Until(cachedRecord.expiresAt)
	}

	// copy in the metadata
	for k, v := range cachedRecord.metadata {
		newRecord.Metadata[k] = v
	}

	return newRecord, nil
}

func (m *memoryCache) set(prefix string, r *cache.Record) {
	key := m.key(prefix, r.Key)

	// copy the incoming record and then
	// convert the expiry in to a hard timestamp
	i := &cacheRecord{}
	i.key = r.Key
	i.value = make([]byte, len(r.Value))
	i.metadata = make(map[string]interface{})

	// copy the the value
	copy(i.value, r.Value)

	// set the expiry
	if r.Expiry != 0 {
		i.expiresAt = time.Now().Add(r.Expiry)
	}

	// set the metadata
	for k, v := range r.Metadata {
		i.metadata[k] = v
	}

	m.cache.Set(key, i, r.Expiry)
}

func (m *memoryCache) delete(prefix, key string) {
	key = m.key(prefix, key)
	m.cache.Delete(key)
}

func (m *memoryCache) list(prefix string, limit, offset uint) []string {
	allItems := m.cache.Items()
	allKeys := make([]string, len(allItems))
	i := 0

	for k := range allItems {
		if !strings.HasPrefix(k, prefix+"/") {
			continue
		}
		allKeys[i] = strings.TrimPrefix(k, prefix+"/")
		i++
	}

	if limit != 0 || offset != 0 {
		sort.Slice(allKeys, func(i, j int) bool { return allKeys[i] < allKeys[j] })
		min := func(i, j uint) uint {
			if i < j {
				return i
			}
			return j
		}
		return allKeys[offset:min(limit, uint(len(allKeys)))]
	}

	return allKeys
}

func (m *memoryCache) Init(opts ...cache.Option) error {
	for _, o := range opts {
		o(&m.options)
	}
	return nil
}

func (m *memoryCache) Options() cache.Options {
	return m.options
}

func (m *memoryCache) Get(ctx context.Context, key string, opts ...cache.GetOption) ([]*cache.Record, error) {
	readOpts := cache.GetOptions{}
	for _, o := range opts {
		o(&readOpts)
	}

	prefix := m.prefix(readOpts.Database, readOpts.Table)

	var keys []string

	// Handle Prefix / suffix
	if readOpts.Prefix || readOpts.Suffix {
		k := m.list(prefix, readOpts.Limit, readOpts.Offset)

		for _, kk := range k {
			if readOpts.Prefix && !strings.HasPrefix(kk, key) {
				continue
			}

			if readOpts.Suffix && !strings.HasSuffix(kk, key) {
				continue
			}

			keys = append(keys, kk)
		}
	} else {
		keys = []string{key}
	}

	var results []*cache.Record

	for _, k := range keys {
		r, err := m.get(prefix, k)
		if err != nil {
			return results, err
		}
		results = append(results, r)
	}

	return results, nil
}

func (m *memoryCache) Put(ctx context.Context, r *cache.Record, opts ...cache.PutOption) error {
	writeOpts := cache.PutOptions{}
	for _, o := range opts {
		o(&writeOpts)
	}

	prefix := m.prefix(writeOpts.Database, writeOpts.Table)

	if len(opts) > 0 {
		// Copy the record before applying options, or the incoming record will be mutated
		newRecord := cache.Record{}
		newRecord.Key = r.Key
		newRecord.Value = make([]byte, len(r.Value))
		newRecord.Metadata = make(map[string]interface{})
		copy(newRecord.Value, r.Value)
		newRecord.Expiry = r.Expiry

		if !writeOpts.Expiry.IsZero() {
			newRecord.Expiry = time.Until(writeOpts.Expiry)
		}
		if writeOpts.TTL != 0 {
			newRecord.Expiry = writeOpts.TTL
		}

		for k, v := range r.Metadata {
			newRecord.Metadata[k] = v
		}

		m.set(prefix, &newRecord)
		return nil
	}

	// set
	m.set(prefix, r)

	return nil
}

func (m *memoryCache) Del(ctx context.Context, key string, opts ...cache.DelOption) error {
	deleteOptions := cache.DelOptions{}
	for _, o := range opts {
		o(&deleteOptions)
	}

	prefix := m.prefix(deleteOptions.Database, deleteOptions.Table)
	m.delete(prefix, key)
	return nil
}

func (m *memoryCache) List(ctx context.Context, opts ...cache.ListOption) ([]string, error) {
	listOptions := cache.ListOptions{}

	for _, o := range opts {
		o(&listOptions)
	}

	prefix := m.prefix(listOptions.Database, listOptions.Table)
	keys := m.list(prefix, listOptions.Limit, listOptions.Offset)

	if len(listOptions.Prefix) > 0 {
		var prefixKeys []string
		for _, k := range keys {
			if strings.HasPrefix(k, listOptions.Prefix) {
				prefixKeys = append(prefixKeys, k)
			}
		}
		keys = prefixKeys
	}

	if len(listOptions.Suffix) > 0 {
		var suffixKeys []string
		for _, k := range keys {
			if strings.HasSuffix(k, listOptions.Suffix) {
				suffixKeys = append(suffixKeys, k)
			}
		}
		keys = suffixKeys
	}

	return keys, nil
}

func (m *memoryCache) Close() error {
	m.cache.Flush()
	return nil
}

func (m *memoryCache) String() string {
	return "memory"
}
