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

package nop

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lack-io/vine/service/dao"
)

type nopDao struct {
	options dao.Options
}

func (n nopDao) Init(option ...dao.Option) error {
	return nil
}

func (n nopDao) Options() dao.Options {
	return n.options
}

func (n nopDao) Migrate(models ...dao.Model) error {
	return fmt.Errorf("nop dao")
}

func (n nopDao) DB() dao.DB {
	return &nopDB{err: fmt.Errorf("nop DB")}
}

func (n nopDao) String() string {
	panic("implement me")
}

type nopDB struct {
	err error
}

func (db nopDB) JSONQuery(predicate *dao.Predicate) dao.DB {
	panic("implement me")
}

func (db nopDB) WithContext(ctx context.Context) dao.DB {
	return db
}

func (db nopDB) Table(name string) dao.DB {
	return db
}

func (db nopDB) Distinct(args ...interface{}) dao.DB {
	return db
}

func (db nopDB) Select(columns ...string) dao.DB {
	return db
}

func (db nopDB) Where(query interface{}, args ...interface{}) dao.DB {
	return db
}

func (db nopDB) Or(query interface{}, args ...interface{}) dao.DB {
	return db
}

func (db nopDB) Take(dest interface{}, args ...interface{}) dao.DB {
	return db
}

func (db nopDB) First(dest interface{}, conds ...interface{}) dao.DB {
	return db
}

func (db nopDB) Last(dest interface{}, conds ...interface{}) dao.DB {
	return db
}

func (db nopDB) Find(dest interface{}, conds ...interface{}) dao.DB {
	return db
}

func (db nopDB) Limit(limit int32) dao.DB {
	return db
}

func (db nopDB) Offset(offset int32) dao.DB {
	return db
}

func (db nopDB) Group(name string) dao.DB {
	return db
}

func (db nopDB) Having(query interface{}, args ...interface{}) dao.DB {
	return db
}

func (db nopDB) Joins(query interface{}, args ...interface{}) dao.DB {
	return db
}

func (db nopDB) Omit(columns ...string) dao.DB {
	return db
}

func (db nopDB) Order(value interface{}) dao.DB {
	return db
}

func (db nopDB) Not(query interface{}, args ...interface{}) dao.DB {
	return db
}

func (db nopDB) Count(count *int64) dao.DB {
	return db
}

func (db nopDB) Create(value dao.Schema) dao.DB {
	return db
}

func (db nopDB) Updates(value dao.Schema) dao.DB {
	return db
}

func (db nopDB) Delete(value dao.Schema, conds ...interface{}) dao.DB {
	return db
}

func (db nopDB) Exec(sql string, values ...interface{}) dao.DB {
	return db
}

func (db nopDB) Scan(dest interface{}) dao.DB {
	return db
}

func (db nopDB) Row() *sql.Row {
	return nil
}

func (db nopDB) Rows() (*sql.Rows, error) {
	return nil, fmt.Errorf("nop DB")
}

func (db nopDB) Begin(opts ...*sql.TxOptions) dao.DB {
	return db
}

func (db nopDB) Rollback() dao.DB {
	return db
}

func (db nopDB) Commit() dao.DB {
	return db
}

func (db nopDB) Err() error {
	return db.err
}

var (
	_ dao.Dao = (*nopDao)(nil)
	_ dao.DB  = (*nopDB)(nil)
)

func newNopDao(opts ...dao.Option) *nopDao {
	options := dao.NewOptions(opts...)

	for _, o := range opts {
		o(&options)
	}

	return &nopDao{options: options}
}

func NewDao(opts ...dao.Option) dao.Dao {
	return newNopDao(opts...)
}
