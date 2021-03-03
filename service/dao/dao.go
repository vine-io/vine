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
	"database/sql"
	"time"
)

type Dao interface {
	Init(...Option) error
	Options() Options
	Migrate(...Model) error
	DB() DB
	String() string
}

type Model interface {
	GetSchema() Schema
	Create(ctx context.Context) error
	FindOne(ctx context.Context, dest interface{}, ps ...P) error
	FindAll(ctx context.Context, dest interface{}, ps ...P) error
	Update(ctx context.Context, ps ...P) error
	BatchUpdate(ctx context.Context, ps ...P) error
	Delete(ctx context.Context, ps ...P) error
	BatchDelete(ctx context.Context, ps ...P) error
}

type Schema interface {
	TableName() string
	Fields() []string
	PrimaryKey() string
	From(Model) Schema
	To() Model
}

type DB interface {
	WithContext(ctx context.Context) DB
	Table(name string) DB
	Distinct(args ...interface{}) DB
	Select(columns ...string) DB
	Where(query interface{}, args ...interface{}) DB
	Or(query interface{}, args ...interface{}) DB
	First(dest interface{}, conds ...interface{}) DB
	Last(dest interface{}, conds ...interface{}) DB
	Find(dest interface{}, conds ...interface{}) DB
	Take(dest interface{}, args ...interface{}) DB
	Limit(limit int32) DB
	Offset(offset int32) DB
	Group(name string) DB
	Having(query interface{}, args ...interface{}) DB
	Joins(query interface{}, args ...interface{}) DB
	Omit(columns ...string) DB
	Order(value interface{}) DB
	Not(query interface{}, args ...interface{}) DB
	Count(count *int64) DB
	Create(Schema) DB
	Updates(Schema) DB
	Delete(value Schema, conds ...interface{}) DB
	Exec(sql string, values ...interface{}) DB
	Scan(dest interface{}) DB
	Row() *sql.Row
	Rows() (*sql.Rows, error)
	Begin(opts ...*sql.TxOptions) DB
	Rollback() DB
	Commit() DB
	JSONQuery(predicate *Predicate) DB
	Err() error
}

var (
	DefaultDao             Dao
	DefaultMaxIdleConns    int32         = 10
	DefaultMaxOpenConns    int32         = 100
	DefaultConnMaxLifetime time.Duration = time.Hour
)
