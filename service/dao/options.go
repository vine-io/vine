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

//
//type Options struct {
//	// Dialect holds the value of database dialect
//	Dialect string
//	// DSN holds the value of database driver name
//	DSN string
//	// MaxIdleConns holds the maximum number of connections in the idle connection pool.
//	MaxIdleConns int32
//	// MaxOpenConns holds the maximum number of open connections to the database.
//	MaxOpenConns int32
//	// ConnMaxLifeTime holds the maximum amount of time a connection may be reused.
//	ConnMaxLifeTime time.Duration
//	// Other options for implementations of the interface
//	// can be stored in a context
//	Context context.Context
//}
//
//func NewOptions(opts ...Option) Options {
//	options := Options{
//		MaxIdleConns:    DefaultMaxIdleConns,
//		MaxOpenConns:    DefaultMaxOpenConns,
//		ConnMaxLifeTime: DefaultConnMaxLifetime,
//	}
//
//	for _, opt := range opts {
//		opt(&options)
//	}
//	return options
//}
//
//type Option func(*Options)
//
//// WithDialect sets database dialect, like MySQL, Postgres, SQLite, etc
//func WithDialect(dialect string) Option {
//	return func(o *Options) {
//		o.Dialect = dialect
//	}
//}
//
//// WithDSN sets DSN
//func WithDSN(dsn string) Option {
//	return func(o *Options) {
//		o.DSN = dsn
//	}
//}
//
//// WithMaxIdleConns sets the maximum number of connections in the idle connection pool.
//func WithMaxIdleConns(conns int32) Option {
//	return func(o *Options) {
//		o.MaxIdleConns = conns
//	}
//}
//
//// WithMaxOpenConns sets the maximum number of open connections to the database.
//func WithMaxOpenConns(conns int32) Option {
//	return func(o *Options) {
//		o.MaxOpenConns = conns
//	}
//}
//
//// WithConnMaxLifeTime sets the maximum amount of time a connection may be reused.
//func WithConnMaxLifeTime(duration time.Duration) Option {
//	return func(o *Options) {
//		o.ConnMaxLifeTime = duration
//	}
//}
