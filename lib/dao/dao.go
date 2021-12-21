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
	"database/sql"
	"fmt"
	"strings"

	"github.com/vine-io/vine/lib/dao/clause"
	"github.com/vine-io/vine/lib/dao/schema"
)

// Dialect DAO database dialect
type Dialect interface {
	Init(...Option) error
	Options() Options
	NewTx() *DB
	Migrator() Migrator
	DataTypeOf(*schema.Field) string
	DefaultValueOf(*schema.Field) clause.Expression
	BindVarTo(writer clause.Writer, stmt *Statement, v interface{})
	QuoteTo(clause.Writer, string)
	Explain(sql string, vars ...interface{}) string
	JSONBuild(column string) JSONQuery
	JSONDataType() string
	String() string
}

// ConnPool db conns pool interface
type ConnPool interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// SavePointerDialectorInterface save pointer interface
type SavePointerDialectorInterface interface {
	SavePoint(tx *DB, name string) error
	RollbackTo(tx *DB, name string) error
}

type TxBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type ConnPoolBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (ConnPool, error)
}

type TxCommitter interface {
	Commit() error
	Rollback() error
}

// Valuer dao valuer interface
type Valuer interface {
	DaoValue(context.Context, *DB) clause.Expr
}

type JSONOp int32

const (
	JSONHasKey JSONOp = iota + 1 // JSON Key
	JSONEq                       // ==
	JSONNeq                      // <>
	JSONGt                       // >
	JSONGte                      // >=
	JSONLt                       // <
	JSONLte                      // <=
	JSONLike                     // like
)

func (j JSONOp) String() string {
	switch j {
	case JSONHasKey:
		return " IS NOT NULL"
	case JSONEq:
		return "="
	case JSONNeq:
		return "<>"
	case JSONGte:
		return ">="
	case JSONGt:
		return ">"
	case JSONLte:
		return "<="
	case JSONLt:
		return "<"
	case JSONLike:
		return "LIKE"
	default:
		return fmt.Sprintf("%d", j)
	}
}

func ParseOp(v interface{}) (op JSONOp) {
	switch v.(type) {
	case string:
		vv := v.(string)
		if strings.HasPrefix(vv, "%") || strings.HasSuffix(vv, "%") {
			op = JSONLike
		} else {
			op = JSONEq
		}
	default:
		op = JSONEq
	}
	return
}

// JSONQuery query column as json
type JSONQuery interface {
	Op(op JSONOp, value interface{}, keys ...string) JSONQuery
	Contains(op JSONOp, values interface{}, keys ...string) JSONQuery
	Tx(tx *DB) JSONQuery
	Build(builder clause.Builder)
}

var (
	DefaultDialect Dialect
)
