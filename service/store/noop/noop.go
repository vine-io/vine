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

import "github.com/lack-io/vine/service/store"

type noopStore struct{}

func (n *noopStore) Init(opts ...store.Option) error {
	return nil
}

func (n *noopStore) Options() store.Options {
	return store.Options{}
}

func (n *noopStore) String() string {
	return "noop"
}

func (n *noopStore) Read(key string, opts ...store.ReadOption) ([]*store.Record, error) {
	return []*store.Record{}, nil
}

func (n *noopStore) Write(r *store.Record, opts ...store.WriteOption) error {
	return nil
}

func (n *noopStore) Delete(key string, opts ...store.DeleteOption) error {
	return nil
}

func (n *noopStore) List(opts ...store.ListOption) ([]string, error) {
	return []string{}, nil
}

func (n *noopStore) Close() error {
	return nil
}

func NewStore() store.Store {
	return new(noopStore)
}
