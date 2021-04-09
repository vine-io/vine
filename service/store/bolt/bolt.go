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

// Package local is a bolt system backed store
package bolt

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	json "github.com/json-iterator/go"
	bolt "go.etcd.io/bbolt"

	"github.com/lack-io/vine/service/store"
)

var (
	// DefaultDatabase is the namespace that the bbolt store
	// will use if no namespace is provided.
	DefaultDatabase = "vine"
	// DefaultTable when none is specified
	DefaultTable = "vine"
	// DefaultDir is the default directory for bbolt files
	DefaultDir = filepath.Join(os.TempDir(), "vine", "store")

	// bucket used for data storage
	dataBucket = "data"
)

// NewStore returns a memory store
func NewStore(opts ...store.Option) store.Store {
	s := &boltStore{
		handles: make(map[string]*boltHandle),
	}
	s.init(opts...)
	return s
}

type boltStore struct {
	options store.Options
	dir     string

	// the database handle
	sync.RWMutex
	handles map[string]*boltHandle
}

type boltHandle struct {
	key string
	db  *bolt.DB
}

// record stored by us
type record struct {
	Key       string
	Value     []byte
	Metadata  map[string]interface{}
	ExpiresAt time.Time
}

func key(database, table string) string {
	return database + ":" + table
}

func (s *boltStore) delete(fd *boltHandle, key string) error {
	return fd.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dataBucket))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(key))
	})
}

func (s *boltStore) init(opts ...store.Option) error {
	for _, o := range opts {
		o(&s.options)
	}

	if s.options.Database == "" {
		s.options.Database = DefaultDatabase
	}

	if s.options.Table == "" {
		// bbolt requires bucketname to not be empty
		s.options.Table = DefaultTable
	}

	// create a directory /tmp/vine
	dir := filepath.Join(DefaultDir, s.options.Database)
	// Ignoring this as the folder might exist.
	// Reads/Writes updates will return with sensible error messages
	// about the dir not existing in case this cannot create the path anyway
	os.MkdirAll(dir, 0700)

	return nil
}

func (s *boltStore) getDB(database, table string) (*boltHandle, error) {
	if len(database) == 0 {
		database = s.options.Database
	}
	if len(table) == 0 {
		table = s.options.Table
	}

	k := key(database, table)
	s.RLock()
	fd, ok := s.handles[k]
	s.RUnlock()

	// return the bolt handle
	if ok {
		return fd, nil
	}

	// double check locking
	s.Lock()
	defer s.Unlock()
	if fd, ok := s.handles[k]; ok {
		return fd, nil
	}

	// create a directory /tmp/vine
	dir := filepath.Join(DefaultDir, database)
	// create the database handle
	fname := table + ".db"
	// make the dir
	os.MkdirAll(dir, 0700)
	// database path
	dbPath := filepath.Join(dir, fname)

	// create new db handle
	// Bolt DB only allows one process to open the bolt R/W so make sure we're doing this under a lock
	db, err := bolt.Open(dbPath, 0700, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, err
	}
	fd = &boltHandle{
		key: k,
		db:  db,
	}
	s.handles[k] = fd

	return fd, nil
}

func (s *boltStore) list(fd *boltHandle, limit, offset uint) []string {
	var allItems []string

	fd.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dataBucket))
		// nothing to read
		if b == nil {
			return nil
		}

		// @todo very inefficient
		if err := b.ForEach(func(k, v []byte) error {
			storedRecord := &record{}

			if err := json.Unmarshal(v, storedRecord); err != nil {
				return err
			}

			if !storedRecord.ExpiresAt.IsZero() {
				if storedRecord.ExpiresAt.Before(time.Now()) {
					return nil
				}
			}

			allItems = append(allItems, string(k))

			return nil
		}); err != nil {
			return err
		}

		return nil
	})

	allKeys := make([]string, len(allItems))

	for i, k := range allItems {
		allKeys[i] = k
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

func (s *boltStore) get(fd *boltHandle, k string) (*store.Record, error) {
	var value []byte

	fd.db.View(func(tx *bolt.Tx) error {
		// @todo this is still very experimental...
		b := tx.Bucket([]byte(dataBucket))
		if b == nil {
			return nil
		}

		value = b.Get([]byte(k))
		return nil
	})

	if value == nil {
		return nil, store.ErrNotFound
	}

	storedRecord := &record{}

	if err := json.Unmarshal(value, storedRecord); err != nil {
		return nil, err
	}

	newRecord := &store.Record{}
	newRecord.Key = storedRecord.Key
	newRecord.Value = storedRecord.Value
	newRecord.Metadata = make(map[string]interface{})

	for k, v := range storedRecord.Metadata {
		newRecord.Metadata[k] = v
	}

	if !storedRecord.ExpiresAt.IsZero() {
		if storedRecord.ExpiresAt.Before(time.Now()) {
			return nil, store.ErrNotFound
		}
		newRecord.Expiry = time.Until(storedRecord.ExpiresAt)
	}

	return newRecord, nil
}

func (s *boltStore) set(fd *boltHandle, r *store.Record) error {
	// copy the incoming record and then
	// convert the expiry in to a hard timestamp
	item := &record{}
	item.Key = r.Key
	item.Value = r.Value
	item.Metadata = make(map[string]interface{})

	if r.Expiry != 0 {
		item.ExpiresAt = time.Now().Add(r.Expiry)
	}

	for k, v := range r.Metadata {
		item.Metadata[k] = v
	}

	// marshal the data
	data, _ := json.Marshal(item)

	return fd.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dataBucket))
		if b == nil {
			var err error
			b, err = tx.CreateBucketIfNotExists([]byte(dataBucket))
			if err != nil {
				return err
			}
		}
		return b.Put([]byte(r.Key), data)
	})
}

func (s *boltStore) Close() error {
	s.Lock()
	defer s.Unlock()
	for k, v := range s.handles {
		v.db.Close()
		delete(s.handles, k)
	}
	return nil
}

func (s *boltStore) Init(opts ...store.Option) error {
	return s.init(opts...)
}

func (s *boltStore) Delete(key string, opts ...store.DeleteOption) error {
	var deleteOptions store.DeleteOptions
	for _, o := range opts {
		o(&deleteOptions)
	}

	fd, err := s.getDB(deleteOptions.Database, deleteOptions.Table)
	if err != nil {
		return err
	}

	return s.delete(fd, key)
}

func (s *boltStore) Read(key string, opts ...store.ReadOption) ([]*store.Record, error) {
	var readOpts store.ReadOptions
	for _, o := range opts {
		o(&readOpts)
	}

	fd, err := s.getDB(readOpts.Database, readOpts.Table)
	if err != nil {
		return nil, err
	}

	var keys []string

	// Handle Prefix / suffix
	// TODO: do range scan here rather than listing all keys
	if readOpts.Prefix || readOpts.Suffix {
		// list the keys
		k := s.list(fd, readOpts.Limit, readOpts.Offset)

		// check for prefix and suffix
		for _, v := range k {
			if readOpts.Prefix && !strings.HasPrefix(v, key) {
				continue
			}
			if readOpts.Suffix && !strings.HasSuffix(v, key) {
				continue
			}
			keys = append(keys, v)
		}
	} else {
		keys = []string{key}
	}

	var results []*store.Record

	for _, k := range keys {
		r, err := s.get(fd, k)
		if err != nil {
			return results, err
		}
		results = append(results, r)
	}

	return results, nil
}

func (s *boltStore) Write(r *store.Record, opts ...store.WriteOption) error {
	var writeOpts store.WriteOptions
	for _, o := range opts {
		o(&writeOpts)
	}

	fd, err := s.getDB(writeOpts.Database, writeOpts.Table)
	if err != nil {
		return err
	}

	if len(opts) > 0 {
		// Copy the record before applying options, or the incoming record will be mutated
		newRecord := store.Record{}
		newRecord.Key = r.Key
		newRecord.Value = r.Value
		newRecord.Metadata = make(map[string]interface{})
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

		return s.set(fd, &newRecord)
	}

	return s.set(fd, r)
}

func (s *boltStore) Options() store.Options {
	return s.options
}

func (s *boltStore) List(opts ...store.ListOption) ([]string, error) {
	var listOptions store.ListOptions

	for _, o := range opts {
		o(&listOptions)
	}

	fd, err := s.getDB(listOptions.Database, listOptions.Table)
	if err != nil {
		return nil, err
	}

	// TODO apply prefix/suffix in range query
	allKeys := s.list(fd, listOptions.Limit, listOptions.Offset)

	if len(listOptions.Prefix) > 0 {
		var prefixKeys []string
		for _, k := range allKeys {
			if strings.HasPrefix(k, listOptions.Prefix) {
				prefixKeys = append(prefixKeys, k)
			}
		}
		allKeys = prefixKeys
	}

	if len(listOptions.Suffix) > 0 {
		var suffixKeys []string
		for _, k := range allKeys {
			if strings.HasSuffix(k, listOptions.Suffix) {
				suffixKeys = append(suffixKeys, k)
			}
		}
		allKeys = suffixKeys
	}

	return allKeys, nil
}

func (s *boltStore) String() string {
	return "bolt"
}
