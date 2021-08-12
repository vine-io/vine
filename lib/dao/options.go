// Copyright 2021 lack
//
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

package dao

import (
	"context"
	"sync"
	"time"

	"github.com/vine-io/vine/lib/dao/clause"
	"github.com/vine-io/vine/lib/dao/logger"
	"github.com/vine-io/vine/lib/dao/schema"
)

// Options DAO configuration
type Options struct {
	DSN string
	// You can disable it by setting `SkipDefaultTransaction` to true
	SkipDefaultTransaction bool
	// NamingStrategy tables, columns naming strategy
	NamingStrategy schema.Namer
	// FullSaveAssociations full save associations
	FullSaveAssociations bool
	// Logger
	Logger logger.Interface
	// NowFunc the function to be used when creating a new timestamp
	NowFunc func() time.Time
	// DryRun generate sql without execute
	DryRun bool
	// PrepareStmt executes the given query in cached statement
	PrepareStmt bool
	// DisableAutomaticPing
	DisableAutomaticPing bool
	// DisableForeignKeyConstraintWhenMigrating
	DisableForeignKeyConstraintWhenMigrating bool
	// DisableNestedTransaction disable nested transaction
	DisableNestedTransaction bool
	// AllowGlobalUpdate allow global update
	AllowGlobalUpdate bool
	// QueryFields executes the SQL query with all fields of the table
	QueryFields bool
	// CreateBatchSize default create batch size
	CreateBatchSize int

	// ClauseBuilders clause builder
	ClauseBuilders map[string]clause.ClauseBuilder
	// ConnPool db conn pool
	ConnPool   ConnPool
	callbacks  *callbacks
	cacheStore *sync.Map

	Context context.Context
}

func NewOptions(opts ...Option) Options {
	var options Options

	for _, opt := range opts {
		opt(&options)
	}

	if options.NamingStrategy == nil {
		options.NamingStrategy = schema.NamingStrategy{}
	}

	if options.Logger == nil {
		options.Logger = logger.Default
	}

	if options.NowFunc == nil {
		options.NowFunc = func() time.Time { return time.Now().Local() }
	}

	if options.cacheStore == nil {
		options.cacheStore = &sync.Map{}
	}

	options.Context = context.Background()

	return options
}

type Option func(*Options)

func DSN(dsn string) Option {
	return func(o *Options) {
		o.DSN = dsn
	}
}

func Namer(namer schema.Namer) Option {
	return func(o *Options) {
		o.NamingStrategy = namer
	}
}

func Logger(l logger.Interface) Option {
	return func(o *Options) {
		o.Logger = l
	}
}
