// Copyright 2021 lack
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dialect

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"

	log "github.com/lack-io/vine/service/logger"
)

// ExecQuerier wraps the 2 database operations.
type ExecQuerier interface {
	// Exec executes a query that doesn't return rows. For example, in SQL, INSERT or UPDATE.
	// It scans the result into the pointer v. In SQL, you it's usually sql.Result.
	Exec(ctx context.Context, query string, args, v interface{}) error
	// Query executes a query that returns rows, typically a SELECT in SQL.
	// It scans the result into the pointer v. In SQL, you it's usually *sql.Rows.
	Query(ctx context.Context, query string, args, v interface{}) error
}

// Driver is the interface that wraps all necessary operations for ent.Ent
type Driver interface {
	ExecQuerier
	// Tx starts and returns a new transaction.
	// The provided context is used until the transaction is committed or rolled back.
	Tx(context.Context) (Tx, error)
	// Close closes the underlying connection.
	Close() error
	// Dialect returns the dialect name of the driver.
	Dialect() string
}

// Tx wraps the Exec and Query interface in transaction.
type Tx interface {
	ExecQuerier
	driver.Tx
}

type nopTx struct {
	Driver
}

func (nopTx) Commit() error   { return nil }
func (nopTx) Rollback() error { return nil }

// NopTx returns a Tx with a no-op Commit / Rollback methods wrapping
// the provided Driver d.
func NopTx(d Driver) Tx {
	return nopTx{d}
}

// DebugDriver is a driver that logs all driver operations.
type DebugDriver struct {
	// underlying driver.
	Driver
	// log function. Defaults to log.Println
	log func(context.Context, ...interface{})
}

// Debug gets a driver and an optional logging function, and returns
// a new debugged-driver that prints all outgoing operations.
func Debug(d Driver, logger ...func(...interface{})) Driver {
	logf := log.Info
	if len(logger) == 1 {
		logf = logger[0]
	}
	drv := &DebugDriver{d, func(_ context.Context, v ...interface{}) { logf(v...) }}
	return drv
}

// DebugWithContext gets a driver and a logging function, and returns
// a new debugged-driver that prints all outgoing operations with context.
func DebugWithContext(d Driver, logger func(context.Context, ...interface{})) Driver {
	drv := &DebugDriver{d, logger}
	return drv
}

// Exec logs its params and calls the underlying driver Exec method.
func (d *DebugDriver) Exec(ctx context.Context, query string, args, v interface{}) error {
	d.log(ctx, fmt.Sprintf("driver.Exec: query=%v args=%v", query, args))
	return d.Driver.Exec(ctx, query, args, v)
}

// Query logs its params and calls the underlying driver Query method.
func (d *DebugDriver) Query(ctx context.Context, query string, args, v interface{}) error {
	d.log(ctx, fmt.Sprintf("driver.Query: query=%v args=%v", query, args))
	return d.Driver.Query(ctx, query, args, v)
}

// Tx adds an log-id for the transaction and calls the underlying driver Tx command
func (d *DebugDriver) Tx(ctx context.Context) (Tx, error) {
	tx, err := d.Driver.Tx(ctx)
	if err != nil {
		return nil, err
	}
	id := uuid.New().String()
	d.log(ctx, fmt.Sprintf("driver.Tx(%s): started", id))
	return &DebugTx{tx, id, d.log, ctx}, nil
}

// BeginTx adds an log-id for the transaction and calls the underlying driver BeginTx command if it's supported.
func (d *DebugDriver) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	drv, ok := d.Driver.(interface {
		BeginTx(context.Context, *sql.TxOptions) (Tx, error)
	})
	if !ok {
		return nil, fmt.Errorf("Driver.BeginTx is not supported")
	}
	tx, err := drv.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	id := uuid.New().String()
	d.log(ctx, fmt.Sprintf("driver.BeginTx(%s): started", id))
	return &DebugTx{tx, id, d.log, ctx}, nil
}

// DebugTx is a transaction implementation that logs all transaction operations.
type DebugTx struct {
	Tx                                        // underlying transaction
	id  string                                // transaction logging id.
	log func(context.Context, ...interface{}) // log function. Defaults to fmt.Println
	ctx context.Context
}

// Exec logs its params and calls the underlying transaction Exec method.
func (d *DebugTx) Exec(ctx context.Context, query string, args, v interface{}) error {
	d.log(ctx, fmt.Sprintf("Tx(%s).Exec: query=%v args=%v", d.id, query, args))
	return d.Tx.Exec(ctx, query, args, v)
}

// Query logs its params and calls the underlying transaction Query method.
func (d *DebugTx) Query(ctx context.Context, query string, args, v interface{}) error {
	d.log(ctx, fmt.Sprintf("Tx(%s).Query: query=%v args=%v", d.id, query, args))
	return d.Tx.Query(ctx, query, args, v)
}

// Commit logs this step and calls the underlying transaction Commit method.
func (d *DebugTx) Commit() error {
	d.log(d.ctx, fmt.Sprintf("Tx(%s): commited", d.id))
	return d.Tx.Commit()
}

// Rollback logs this step and calls the underlying transaction Rollback method.
func (d *DebugTx) Rollback() error {
	d.log(d.ctx, fmt.Sprintf("Tx(%s): rollbacked", d.id))
	return d.Tx.Rollback()
}
