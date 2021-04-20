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

package cache

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lack-io/vine/lib/store"
	"github.com/lack-io/vine/lib/store/memory"
)

func TestCache(t *testing.T) {
	l0, l1, l2 := memory.NewStore(store.Database("l0")), memory.NewStore(store.Table("l1")), memory.NewStore()
	_, _, _ = l0.Init(), l1.Init(), l2.Init()

	assert := assert.New(t)

	nonCache := NewCache(nil)
	assert.Equal(len(nonCache.(*cache).stores), 1, "Expected a cache initialised with just 1 store to fail")

	// Basic functionality
	cachedStore := NewCache(l0, l1, l2)
	assert.Equal(cachedStore.Options(), l0.Options(), "Options on store/cache are nonsensical")
	expectedString := "cache [memory memory memory]"
	assert.Equal(cachedStore.String(), expectedString, "Cache couldn't describe itself as expected")

	// Read/Write tests
	_, err := cachedStore.Read("test")
	assert.Equal(store.ErrNotFound, err, "Read non existant key")
	r1 := &store.Record{
		Key:      "aaa",
		Value:    []byte("bbb"),
		Metadata: map[string]interface{}{},
	}
	r2 := &store.Record{
		Key:      "aaaa",
		Value:    []byte("bbbb"),
		Metadata: map[string]interface{}{},
	}
	r3 := &store.Record{
		Key:      "aaaaa",
		Value:    []byte("bbbbb"),
		Metadata: map[string]interface{}{},
	}
	// Write 3 records directly to l2
	l2.Write(r1)
	l2.Write(r2)
	l2.Write(r3)
	// Ensure it's not in l0
	assert.Equal(store.ErrNotFound, func() error { _, err := l0.Read(r1.Key); return err }())
	// Read from cache, ensure it's in all 3 stores
	results, err := cachedStore.Read(r1.Key)
	assert.Nil(err, "cachedStore.Read() returned error")
	assert.Len(results, 1, "cachedStore.Read() should only return 1 result")
	assert.Equal(r1, results[0], "Cached read didn't return the record that was put in")
	results, err = l0.Read(r1.Key)
	assert.Nil(err)
	assert.Equal(r1, results[0], "l0 not coherent")
	results, err = l1.Read(r1.Key)
	assert.Nil(err)
	assert.Equal(r1, results[0], "l1 not coherent")
	results, err = l2.Read(r1.Key)
	assert.Nil(err)
	assert.Equal(r1, results[0], "l2 not coherent")
	// Multiple read
	results, err = cachedStore.Read("aa", store.ReadPrefix())
	assert.Nil(err, "Cachedstore multiple read errored")
	assert.Len(results, 3, "ReadPrefix should have read all records")
	// l1 should now have all 3 records
	l1results, err := l1.Read("aa", store.ReadPrefix())
	assert.Nil(err, "l1.Read failed")
	assert.Len(l1results, 3, "l1 didn't contain a full cache")
	sort.Slice(results, func(i, j int) bool { return results[i].Key < results[j].Key })
	sort.Slice(l1results, func(i, j int) bool { return l1results[i].Key < l1results[j].Key })
	assert.Equal(results[0], l1results[0], "l1 cache not coherent")
	assert.Equal(results[1], l1results[1], "l1 cache not coherent")
	assert.Equal(results[2], l1results[2], "l1 cache not coherent")

	// Test List and Delete
	keys, err := cachedStore.List(store.ListPrefix("a"))
	assert.Nil(err, "List should not error")
	assert.Len(keys, 3, "List should return 3 keys")
	for _, k := range keys {
		err := cachedStore.Delete(k)
		assert.Nil(err, "Delete should not error")
		_, err = cachedStore.Read(k)
		// N.B. - this may not pass on stores that are eventually consistent
		assert.Equal(store.ErrNotFound, err, "record should be gone")
	}

	// Test Write
	err = cachedStore.Write(r1)
	assert.Nil(err, "Write shouldn't fail")
	l2result, err := l2.Read(r1.Key)
	assert.Nil(err)
	assert.Len(l2result, 1)
	assert.Equal(r1, l2result[0], "Write didn't make it all the way through to l2")

}
