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

package memory

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kr/pretty"
	"github.com/vine-io/vine/lib/cache"
)

func TestMemoryReInit(t *testing.T) {
	s := NewCache(cache.Table("aaa"))
	s.Init(cache.Table(""))
	if len(s.Options().Table) > 0 {
		t.Error("Init didn't reinitialise the store")
	}
}

func TestMemoryBasic(t *testing.T) {
	s := NewCache()
	s.Init()
	basictest(s, t)
}

func TestMemoryPrefix(t *testing.T) {
	s := NewCache()
	s.Init(cache.Table("some-prefix"))
	basictest(s, t)
}

func TestMemoryNamespace(t *testing.T) {
	s := NewCache()
	s.Init(cache.Database("some-namespace"))
	basictest(s, t)
}

func TestMemoryNamespacePrefix(t *testing.T) {
	s := NewCache()
	s.Init(cache.Table("some-prefix"), cache.Database("some-namespace"))
	basictest(s, t)
}

func basictest(s cache.Cache, t *testing.T) {
	if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
		t.Logf("Testing store %s, with options %# v\n", s.String(), pretty.Formatter(s.Options()))
	}
	ctx := context.TODO()
	// Get and Put an expiring Record
	if err := s.Put(ctx, &cache.Record{
		Key:    "Hello",
		Value:  []byte("World"),
		Expiry: time.Millisecond * 100,
	}); err != nil {
		t.Error(err)
	}
	if r, err := s.Get(ctx, "Hello"); err != nil {
		t.Error(err)
	} else {
		if len(r) != 1 {
			t.Error("Get returned multiple records")
		}
		if r[0].Key != "Hello" {
			t.Errorf("Expected %s, got %s", "Hello", r[0].Key)
		}
		if string(r[0].Value) != "World" {
			t.Errorf("Expected %s, got %s", "World", r[0].Value)
		}
	}
	time.Sleep(time.Millisecond * 200)
	if _, err := s.Get(ctx, "Hello"); err != cache.ErrNotFound {
		t.Errorf("Expected %# v, got %# v", cache.ErrNotFound, err)
	}

	// Put 3 records with various expiry and get with prefix
	records := []*cache.Record{
		&cache.Record{
			Key:   "foo",
			Value: []byte("foofoo"),
		},
		&cache.Record{
			Key:    "foobar",
			Value:  []byte("foobarfoobar"),
			Expiry: time.Millisecond * 100,
		},
		&cache.Record{
			Key:    "foobarbaz",
			Value:  []byte("foobarbazfoobarbaz"),
			Expiry: 2 * time.Millisecond * 100,
		},
	}
	for _, r := range records {
		if err := s.Put(ctx, r); err != nil {
			t.Errorf("Couldn't write k: %s, v: %# v (%s)", r.Key, pretty.Formatter(r.Value), err)
		}
	}
	if results, err := s.Get(ctx, "foo", cache.GetPrefix()); err != nil {
		t.Errorf("Couldn't read all \"foo\" keys, got %# v (%s)", pretty.Formatter(results), err)
	} else {
		if len(results) != 3 {
			t.Errorf("Expected 3 items, got %d", len(results))
		}
		if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
			t.Logf("Prefix test: %v\n", pretty.Formatter(results))
		}
	}
	time.Sleep(time.Millisecond * 100)
	if results, err := s.Get(ctx, "foo", cache.GetPrefix()); err != nil {
		t.Errorf("Couldn't read all \"foo\" keys, got %# v (%s)", pretty.Formatter(results), err)
	} else {
		if len(results) != 2 {
			t.Errorf("Expected 2 items, got %d", len(results))
		}
		if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
			t.Logf("Prefix test: %v\n", pretty.Formatter(results))
		}
	}
	time.Sleep(time.Millisecond * 100)
	if results, err := s.Get(ctx, "foo", cache.GetPrefix()); err != nil {
		t.Errorf("Couldn't read all \"foo\" keys, got %# v (%s)", pretty.Formatter(results), err)
	} else {
		if len(results) != 1 {
			t.Errorf("Expected 1 item, got %d", len(results))
		}
		if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
			t.Logf("Prefix test: %# v\n", pretty.Formatter(results))
		}
	}
	if err := s.Del(ctx, "foo", func(d *cache.DelOptions) {}); err != nil {
		t.Errorf("Del failed (%v)", err)
	}
	if results, err := s.Get(ctx, "foo", cache.GetPrefix()); err != nil {
		t.Errorf("Couldn't read all \"foo\" keys, got %# v (%s)", pretty.Formatter(results), err)
	} else {
		if len(results) != 0 {
			t.Errorf("Expected 0 items, got %d (%# v)", len(results), pretty.Formatter(results))
		}
	}

	// Put 3 records with various expiry and get with Suffix
	records = []*cache.Record{
		&cache.Record{
			Key:   "foo",
			Value: []byte("foofoo"),
		},
		&cache.Record{
			Key:    "barfoo",
			Value:  []byte("barfoobarfoo"),
			Expiry: time.Millisecond * 100,
		},
		&cache.Record{
			Key:    "bazbarfoo",
			Value:  []byte("bazbarfoobazbarfoo"),
			Expiry: 2 * time.Millisecond * 100,
		},
	}
	for _, r := range records {
		if err := s.Put(ctx, r); err != nil {
			t.Errorf("Couldn't write k: %s, v: %# v (%s)", r.Key, pretty.Formatter(r.Value), err)
		}
	}
	if results, err := s.Get(ctx, "foo", cache.GetSuffix()); err != nil {
		t.Errorf("Couldn't read all \"foo\" keys, got %# v (%s)", pretty.Formatter(results), err)
	} else {
		if len(results) != 3 {
			t.Errorf("Expected 3 items, got %d", len(results))
		}
		if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
			t.Logf("Prefix test: %v\n", pretty.Formatter(results))
		}
	}
	time.Sleep(time.Millisecond * 100)
	if results, err := s.Get(ctx, "foo", cache.GetSuffix()); err != nil {
		t.Errorf("Couldn't read all \"foo\" keys, got %# v (%s)", pretty.Formatter(results), err)
	} else {
		if len(results) != 2 {
			t.Errorf("Expected 2 items, got %d", len(results))
		}
		if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
			t.Logf("Prefix test: %v\n", pretty.Formatter(results))
		}
	}
	time.Sleep(time.Millisecond * 100)
	if results, err := s.Get(ctx, "foo", cache.GetSuffix()); err != nil {
		t.Errorf("Couldn't read all \"foo\" keys, got %# v (%s)", pretty.Formatter(results), err)
	} else {
		if len(results) != 1 {
			t.Errorf("Expected 1 item, got %d", len(results))
		}
		if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
			t.Logf("Prefix test: %# v\n", pretty.Formatter(results))
		}
	}
	if err := s.Del(ctx, "foo"); err != nil {
		t.Errorf("Del failed (%v)", err)
	}
	if results, err := s.Get(ctx, "foo", cache.GetSuffix()); err != nil {
		t.Errorf("Couldn't read all \"foo\" keys, got %# v (%s)", pretty.Formatter(results), err)
	} else {
		if len(results) != 0 {
			t.Errorf("Expected 0 items, got %d (%# v)", len(results), pretty.Formatter(results))
		}
	}

	// Test Prefix, Suffix and PutOptions
	if err := s.Put(ctx, &cache.Record{
		Key:   "foofoobarbar",
		Value: []byte("something"),
	}, cache.PutTTL(time.Millisecond*100)); err != nil {
		t.Error(err)
	}
	if err := s.Put(ctx, &cache.Record{
		Key:   "foofoo",
		Value: []byte("something"),
	}, cache.PutExpiry(time.Now().Add(time.Millisecond*100))); err != nil {
		t.Error(err)
	}
	if err := s.Put(ctx, &cache.Record{
		Key:   "barbar",
		Value: []byte("something"),
		// TTL has higher precedence than expiry
	}, cache.PutExpiry(time.Now().Add(time.Hour)), cache.PutTTL(time.Millisecond*100)); err != nil {
		t.Error(err)
	}
	if results, err := s.Get(ctx, "foo", cache.GetPrefix(), cache.GetSuffix()); err != nil {
		t.Error(err)
	} else {
		if len(results) != 1 {
			t.Errorf("Expected 1 results, got %d: %# v", len(results), pretty.Formatter(results))
		}
	}
	time.Sleep(time.Millisecond * 100)
	if results, err := s.List(ctx); err != nil {
		t.Errorf("List failed: %s", err)
	} else {
		if len(results) != 0 {
			t.Error("Expiry options were not effective")
		}
	}
	s.Put(ctx, &cache.Record{Key: "a", Value: []byte("a")})
	s.Put(ctx, &cache.Record{Key: "aa", Value: []byte("aa")})
	s.Put(ctx, &cache.Record{Key: "aaa", Value: []byte("aaa")})
	if results, err := s.Get(ctx, "b", cache.GetPrefix()); err != nil {
		t.Error(err)
	} else {
		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(results))
		}
	}

	s.Close() // reset the store
	for i := 0; i < 10; i++ {
		s.Put(ctx, &cache.Record{
			Key:   fmt.Sprintf("a%d", i),
			Value: []byte{},
		})
	}
	if results, err := s.Get(ctx, "a", cache.GetLimit(5), cache.GetPrefix()); err != nil {
		t.Error(err)
	} else {
		if len(results) != 5 {
			t.Fatal("Expected 5 results, got ", len(results))
		}
		if results[0].Key != "a0" {
			t.Fatalf("Expected a0, got %s", results[0].Key)
		}
		if results[4].Key != "a4" {
			t.Fatalf("Expected a4, got %s", results[4].Key)
		}
	}
	if results, err := s.Get(ctx, "a", cache.GetLimit(30), cache.GetOffset(5), cache.GetPrefix()); err != nil {
		t.Error(err)
	} else {
		if len(results) != 5 {
			t.Error("Expected 5 results, got ", len(results))
		}
	}
}
