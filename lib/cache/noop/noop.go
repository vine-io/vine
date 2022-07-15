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

package noop

import (
	"context"

	"github.com/vine-io/vine/lib/cache"
)

type noopCache struct{}

func (n *noopCache) Init(opts ...cache.Option) error {
	return nil
}

func (n *noopCache) Options() cache.Options {
	return cache.Options{}
}

func (n *noopCache) String() string {
	return "noop"
}

func (n *noopCache) Get(ctx context.Context, key string, opts ...cache.GetOption) ([]*cache.Record, error) {
	return []*cache.Record{}, nil
}

func (n *noopCache) Put(ctx context.Context, r *cache.Record, opts ...cache.PutOption) error {
	return nil
}

func (n *noopCache) Del(ctx context.Context, key string, opts ...cache.DelOption) error {
	return nil
}

func (n *noopCache) List(ctx context.Context, opts ...cache.ListOption) ([]string, error) {
	return []string{}, nil
}

func (n *noopCache) Close() error {
	return nil
}

func NewCache(opts ...cache.Option) cache.Cache {
	return new(noopCache)
}
