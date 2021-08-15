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
	"context"
	"time"

	"github.com/vine-io/vine/core/client"
)

// Options contains configuration for the Cache
type Options struct {
	// Nodes contains the addresses or other connection information of the backing storage.
	// For example, an etcd implementation would contain the nodes of the cluster.
	// A SQL implementation could contain one or more connection strings.
	Nodes []string
	// Database allows multiple isolated stores to be kept in one backend, if supported.
	Database string
	// Table is analagous to a table in database backends or a key prefix in KV backends
	Table string
	// Context should contain all implementation specific options, using context.WithValue.
	Context context.Context
	// Client to use for RPC
	Client client.Client
}

// Option sets values in Options
type Option func(o *Options)

// Nodes contains the addresses or other connection information of the backing storage.
// For example, an etcd implementation would contain the nodes of the cluster.
// A SQL implementation could contain one or more connection strings.
func Nodes(a ...string) Option {
	return func(o *Options) {
		o.Nodes = a
	}
}

// Database allows multiple isolated stores to be kept in one backend, if supported.
func Database(db string) Option {
	return func(o *Options) {
		o.Database = db
	}
}

// Table is analagous to a table in database backends or a key prefix in KV backends
func Table(t string) Option {
	return func(o *Options) {
		o.Table = t
	}
}

// WithContext sets the stores context, for any extra configuration
func WithContext(c context.Context) Option {
	return func(o *Options) {
		o.Context = c
	}
}

// WithClient sets the stores client to use for RPC
func WithClient(c client.Client) Option {
	return func(o *Options) {
		o.Client = c
	}
}

// GetOptions configures an individual Get operation
type GetOptions struct {
	Database, Table string
	// Prefix returns all records that are prefixed with key
	Prefix bool
	// Suffix returns all records that have the suffix key
	Suffix bool
	// Limit limits the number of returned records
	Limit uint
	// Offset when combined with Limit supports pagination
	Offset uint
}

// GetOption sets values in GetOptions
type GetOption func(r *GetOptions)

// GetFrom the database and table
func GetFrom(database, table string) GetOption {
	return func(r *GetOptions) {
		r.Database = database
		r.Table = table
	}
}

// GetPrefix returns all records that are prefixed with key
func GetPrefix() GetOption {
	return func(r *GetOptions) {
		r.Prefix = true
	}
}

// GetSuffix returns all records that have the suffix key
func GetSuffix() GetOption {
	return func(r *GetOptions) {
		r.Suffix = true
	}
}

// GetLimit limits the number of responses to l
func GetLimit(l uint) GetOption {
	return func(r *GetOptions) {
		r.Limit = l
	}
}

// GetOffset starts returning responses from o. Use in conjunction with Limit for pagination
func GetOffset(o uint) GetOption {
	return func(r *GetOptions) {
		r.Offset = o
	}
}

// PutOptions configures an individual Put operation
// If Expiry and TTL are set TTL takes precedence
type PutOptions struct {
	Database, Table string
	// Expiry is the time the record expires
	Expiry time.Time
	// TTL is the time until the record expires
	TTL time.Duration
}

// PutOption sets values in PutOptions
type PutOption func(w *PutOptions)

// PutTo the database and table
func PutTo(database, table string) PutOption {
	return func(w *PutOptions) {
		w.Database = database
		w.Table = table
	}
}

// PutExpiry is the time the record expires
func PutExpiry(t time.Time) PutOption {
	return func(w *PutOptions) {
		w.Expiry = t
	}
}

// PutTTL is the time the record expires
func PutTTL(d time.Duration) PutOption {
	return func(w *PutOptions) {
		w.TTL = d
	}
}

// DelOptions configures an individual Del operation
type DelOptions struct {
	Database, Table string
}

// DelOption sets values in DelOptions
type DelOption func(d *DelOptions)

// DelFrom the database and table
func DelFrom(database, table string) DelOption {
	return func(d *DelOptions) {
		d.Database = database
		d.Table = table
	}
}

// ListOptions configures an individual List operation
type ListOptions struct {
	// List from the following
	Database, Table string
	// Prefix returns all keys that are prefixed with key
	Prefix string
	// Suffix returns all keys that end with key
	Suffix string
	// Limit limits the number of returned keys
	Limit uint
	// Offset when combined with Limit supports pagination
	Offset uint
}

// ListOption sets values in ListOptions
type ListOption func(l *ListOptions)

// ListFrom the database and table
func ListFrom(database, table string) ListOption {
	return func(l *ListOptions) {
		l.Database = database
		l.Table = table
	}
}

// ListPrefix returns all keys that are prefixed with key
func ListPrefix(p string) ListOption {
	return func(l *ListOptions) {
		l.Prefix = p
	}
}

// ListSuffix returns all keys that end with key
func ListSuffix(s string) ListOption {
	return func(l *ListOptions) {
		l.Suffix = s
	}
}

// ListLimit limits the number of returned keys to l
func ListLimit(l uint) ListOption {
	return func(lo *ListOptions) {
		lo.Limit = l
	}
}

// ListOffset starts returning responses from o. Use in conjunction with Limit for pagination.
func ListOffset(o uint) ListOption {
	return func(l *ListOptions) {
		l.Offset = o
	}
}
