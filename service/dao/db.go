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

package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/lack-io/vine/service/dao/clause"
	"github.com/lack-io/vine/service/dao/schema"
)

// DB DAO DB definition
type DB struct {
	// Dialect database dialect
	Dialect
	Options
	Error        error
	RowsAffected int64
	Statement    *Statement
	clone        int
}

// Open initialize db session based on dialector
func Open(dialect Dialect) (db *DB, err error) {

	db = &DB{Dialect: dialect, Options: dialect.Options(), clone: 1}

	if db.NamingStrategy == nil {
		db.NamingStrategy = schema.NamingStrategy{}
	}

	if db.NowFunc == nil {
		db.NowFunc = func() time.Time { return time.Now().Local() }
	}

	if db.cacheStore == nil {
		db.cacheStore = &sync.Map{}
	}

	db.callbacks = initializeCallbacks(db)

	if db.ClauseBuilders == nil {
		db.ClauseBuilders = map[string]clause.ClauseBuilder{}
	}

	preparedStmt := &PreparedStmtDB{
		ConnPool:    db.ConnPool,
		Stmts:       map[string]Stmt{},
		Mux:         &sync.RWMutex{},
		PreparedSQL: make([]string, 0, 100),
	}
	db.cacheStore.Store("preparedStmt", preparedStmt)

	if db.PrepareStmt {
		db.ConnPool = preparedStmt
	}

	db.Statement = &Statement{
		DB:       db,
		ConnPool: db.ConnPool,
		Context:  context.Background(),
		Clauses:  map[string]clause.Clause{},
	}

	if !db.DisableAutomaticPing {
		if pinger, ok := db.ConnPool.(interface{ Ping() error }); ok {
			err = pinger.Ping()
		}
	}

	return
}

// Session session config when create session with Session() method
type Session struct {
	DryRun                   bool
	PrepareStmt              bool
	NewDB                    bool
	SkipHooks                bool
	SkipDefaultTransaction   bool
	DisableNestedTransaction bool
	AllowGlobalUpdate        bool
	FullSaveAssociation      bool
	QueryFields              bool
	Context                  context.Context
	NowFunc                  func() time.Time
	CreateBatchSize          int
}

// Session create new db session
func (db *DB) Session(config *Session) *DB {
	tx := &DB{
		Dialect:   db.Dialect,
		Options:   db.Options,
		Statement: db.Statement,
		Error:     db.Error,
		clone:     1,
	}

	if config.CreateBatchSize > 0 {
		tx.CreateBatchSize = config.CreateBatchSize
	}

	if config.SkipDefaultTransaction {
		tx.SkipDefaultTransaction = true
	}

	if config.AllowGlobalUpdate {
		tx.AllowGlobalUpdate = true
	}

	if config.FullSaveAssociation {
		tx.FullSaveAssociations = true
	}

	if config.Context != nil || config.PrepareStmt || config.SkipHooks {
		tx.Statement = tx.Statement.clone()
		tx.Statement.DB = tx
	}

	if config.Context != nil {
		tx.Statement.Context = config.Context
	}

	if config.PrepareStmt {
		if v, ok := db.cacheStore.Load("preparedStmt"); ok {
			preparedStmt := v.(*PreparedStmtDB)
			tx.Statement.ConnPool = &PreparedStmtDB{
				ConnPool: db.ConnPool,
				Mux:      preparedStmt.Mux,
				Stmts:    preparedStmt.Stmts,
			}
			db.ConnPool = tx.Statement.ConnPool
			db.PrepareStmt = true
		}
	}

	if config.SkipHooks {
		tx.Statement.SkipHooks = true
	}

	if config.DisableNestedTransaction {
		tx.DisableNestedTransaction = true
	}

	if !config.NewDB {
		tx.clone = 2
	}

	if config.DryRun {
		tx.DryRun = true
	}

	if config.QueryFields {
		tx.QueryFields = true
	}

	if config.NowFunc != nil {
		tx.NowFunc = config.NowFunc
	}

	return tx
}

// WithContext change current instance db's context to ctx
func (db *DB) WithContext(ctx context.Context) *DB {
	return db.Session(&Session{Context: ctx})
}

// Set store value with key into current db instance's context
func (db *DB) Set(key string, value interface{}) *DB {
	tx := db.getInstance()
	tx.Statement.Settings.Store(key, value)
	return tx
}

// Get get value with key from current db instance's context
func (db *DB) Get(key string) (interface{}, bool) {
	return db.Statement.Settings.Load(key)
}

// InstanceSet store value with key into current db instance's context
func (db *DB) InstanceSet(key string, value interface{}) *DB {
	tx := db.getInstance()
	tx.Statement.Settings.Store(fmt.Sprintf("%p", tx.Statement)+key, value)
	return tx
}

// InstanceGet get value with key from current db instance's context
func (db *DB) InstanceGet(key string) (interface{}, bool) {
	return db.Statement.Settings.Load(fmt.Sprintf("%p", db.Statement) + key)
}

// Callback returns callback manager
func (db *DB) Callback() *callbacks {
	return db.callbacks
}

// AddError add error to db
func (db *DB) AddError(err error) error {
	if db.Error == nil {
		db.Error = err
	} else if err != nil {
		db.Error = fmt.Errorf("%v; %w", db.Error, err)
	}
	return db.Error
}

// DB returns `*sql.DB`
func (db *DB) DB() (*sql.DB, error) {
	connPool := db.ConnPool

	if stmtDB, ok := connPool.(*PreparedStmtDB); ok {
		connPool = stmtDB.ConnPool
	}

	if sqldb, ok := connPool.(*sql.DB); ok {
		return sqldb, nil
	}

	return nil, errors.New("invalid db")
}

func (db *DB) getInstance() *DB {
	if db.clone > 0 {
		tx := &DB{Dialect: db.Dialect, Options: db.Options}

		if db.clone == 1 {
			// clone with new statment
			tx.Statement = &Statement{
				DB:       tx,
				ConnPool: db.Statement.ConnPool,
				Context:  db.Statement.Context,
				Clauses:  map[string]clause.Clause{},
				Vars:     make([]interface{}, 0, 8),
			}
		} else {
			// with clone statement
			tx.Statement = db.Statement.clone()
			tx.Statement.DB = tx
		}

		return tx
	}

	return db
}

func Expr(expr string, args ...interface{}) clause.Expr {
	return clause.Expr{SQL: expr, Vars: args}
}

func (db *DB) SetupJoinTable(model interface{}, field string, joinTable interface{}) error {
	var (
		tx                      = db.getInstance()
		stmt                    = tx.Statement
		modelSchema, joinSchema *schema.Schema
	)

	if err := stmt.Parse(model); err == nil {
		modelSchema = stmt.Schema
	} else {
		return err
	}

	if err := stmt.Parse(joinTable); err == nil {
		joinSchema = stmt.Schema
	} else {
		return err
	}

	if relation, ok := modelSchema.Relationships.Relations[field]; ok && relation.JoinTable != nil {
		for _, ref := range relation.References {
			if f := joinSchema.LookUpField(ref.ForeignKey.DBName); f != nil {
				f.DataType = ref.ForeignKey.DataType
				f.DAODataType = ref.ForeignKey.DAODataType
				if f.Size == 0 {
					f.Size = ref.ForeignKey.Size
				}
				ref.ForeignKey = f
			} else {
				return fmt.Errorf("missing field %v for join table", ref.ForeignKey.DBName)
			}
		}

		for name, rel := range relation.JoinTable.Relationships.Relations {
			if _, ok := joinSchema.Relationships.Relations[name]; !ok {
				rel.Schema = joinSchema
				joinSchema.Relationships.Relations[name] = rel
			}
		}

		relation.JoinTable = joinSchema
	} else {
		return fmt.Errorf("failed to found relation: %v", field)
	}

	return nil
}
