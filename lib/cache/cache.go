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

// Package cache is an interface for distributed data cache.
package cache

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrNotFound is returned when a key doesn't exist
	ErrNotFound = errors.New("not found")
	// DefaultCache is the memory cache.
	DefaultCache Cache
)

// Cache is a data cache interface
type Cache interface {
	// Init initialises the cache. It must perform any required setup on the backing storage implementation and check that it is ready for use, returning any errors.
	Init(...Option) error
	// Options allows you to view the current options.
	Options() Options
	// Get takes a single key name and optional GetOptions. It returns matching []*Record or an error.
	Get(ctx context.Context, key string, opts ...GetOption) ([]*Record, error)
	// Put writes a record to the cache, and returns an error if the record was not written.
	Put(ctx context.Context, r *Record, opts ...PutOption) error
	// Del removes the record with the corresponding key from the cache.
	Del(ctx context.Context, key string, opts ...DelOption) error
	// List returns any keys that match, or an empty list with no error if none matched.
	List(ctx context.Context, opts ...ListOption) ([]string, error)
	// Close the cache
	Close() error
	// String returns the name of the implementation.
	String() string
}

// Record is an item cached or retrieved from a Cache
type Record struct {
	// The key to cache the record
	Key string `json:"key"`
	// The value within the record
	Value []byte `json:"value"`
	// Any associated metadata for indexing
	Metadata map[string]interface{} `json:"metadata"`
	// Time to expire a record: TODO: change to timestamp
	Expiry time.Duration `json:"expiry,omitempty"`
}

// Get takes a single key to DefaultCache
func Get(ctx context.Context, key string, opts ...GetOption) ([]*Record, error) {
	return DefaultCache.Get(ctx, key, opts...)
}

// Put writes a record to the DefaultCache,
func Put(ctx context.Context, r *Record, opts ...PutOption) error {
	return DefaultCache.Put(ctx, r, opts...)
}

// Del removes the record with the corresponding key from the DefaultCache.
func Del(ctx context.Context, key string, opts ...DelOption) error {
	return DefaultCache.Del(ctx, key, opts...)
}

// List returns any keys from the DefaultCache
func List(ctx context.Context, opts ...ListOption) ([]string, error) {
	return DefaultCache.List(ctx, opts...)
}
