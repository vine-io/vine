// Copyright 2021 lack
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

import (
	"context"
	"sync"
	"time"

	"github.com/lack-io/vine/service/dao/clause"
	"github.com/lack-io/vine/service/dao/schema"
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
