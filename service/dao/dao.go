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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/lack-io/vine/service/dao/clause"
	"github.com/lack-io/vine/service/dao/schema"
	"github.com/lack-io/vine/service/logger"
)

//
//type Dao interface {
//	Init(...Option) error
//	Options() Options
//	Migrate(...Model) error
//	DB() DB
//	String() string
//}
//
//type Model interface {
//	GetSchema() Schema
//	Create(ctx context.Context) error
//	FindOne(ctx context.Context, dest interface{}, ps ...P) error
//	FindAll(ctx context.Context, dest interface{}, ps ...P) error
//	Update(ctx context.Context, ps ...P) error
//	BatchUpdate(ctx context.Context, ps ...P) error
//	Delete(ctx context.Context, ps ...P) error
//	BatchDelete(ctx context.Context, ps ...P) error
//}
//
//type Schema interface {
//	TableName() string
//	Fields() []string
//	PrimaryKey() string
//	From(Model) Schema
//	To() Model
//}
//
//type DB interface {
//	WithContext(ctx context.Context) DB
//	Table(name string) DB
//	Distinct(args ...interface{}) DB
//	Select(columns ...string) DB
//	Where(query interface{}, args ...interface{}) DB
//	Or(query interface{}, args ...interface{}) DB
//	First(dest interface{}, conds ...interface{}) DB
//	Last(dest interface{}, conds ...interface{}) DB
//	Find(dest interface{}, conds ...interface{}) DB
//	Take(dest interface{}, args ...interface{}) DB
//	Limit(limit int32) DB
//	Offset(offset int32) DB
//	Group(name string) DB
//	Having(query interface{}, args ...interface{}) DB
//	Joins(query interface{}, args ...interface{}) DB
//	Omit(columns ...string) DB
//	Order(value interface{}) DB
//	Not(query interface{}, args ...interface{}) DB
//	Count(count *int64) DB
//	Create(Schema) DB
//	Updates(Schema) DB
//	Delete(value Schema, conds ...interface{}) DB
//	Exec(sql string, values ...interface{}) DB
//	Scan(dest interface{}) DB
//	Row() *sql.Row
//	Rows() (*sql.Rows, error)
//	Begin(opts ...*sql.TxOptions) DB
//	Rollback() DB
//	Commit() DB
//	JSONQuery(predicate *Predicate) DB
//	Err() error
//}
//
//var (
//	DefaultDao             Dao
//	DefaultMaxIdleConns    int32         = 10
//	DefaultMaxOpenConns    int32         = 100
//	DefaultConnMaxLifetime time.Duration = time.Hour
//)

// Config DAO config
type Config struct {
	// You can disable it by setting `SkipDefaultTransaction` to true
	SkipDefaultTransaction bool
	// NamingStrategy tables, columns naming strategy
	NamingStrategy schema.Namer
	// FullSaveAssociations full save associations
	FullSaveAssociations bool
	// Logger
	Logger logger.Logger
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
	ConnPool ConnPool
	// Dialect database dialect
	Dialector

	callbacks  *callbacks
	cacheStore *sync.Map
}

// DB DAO DB definition
type DB struct {
	*Config
	Error        error
	RowsAffected int64
	Statement    *Statement
	clone        int
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

// Open initialize db session based on dialector
func Open(dialector Dialector, config *Config) (db *DB, err error) {
	if config == nil {
		config = &Config{}
	}

	if config.NamingStrategy == nil {
		config.NamingStrategy = schema.NamingStrategy{}
	}

	if config.NowFunc == nil {
		config.NowFunc = func() time.Time {return time.Now().Local()}
	}

	if dialector != nil {
		config.Dialector = dialector
	}

	if config.cacheStore == nil {
		config.cacheStore = &sync.Map{}
	}

	db = &DB{Config: config, clone: 1}

	db.callbacks = initializeCallbacks(db)

	if config.ClauseBuilders == nil {
		config.ClauseBuilders = map[string]clause.ClauseBuilder{}
	}

	if config.Dialector != nil {
		err = config.Dialector.Initialize(db)
	}

	prepareStmt := &PreparedStmtDB{
		ConnPool: db.ConnPool,
		Stmts: map[string]Stmt{},
		Mux: &sync.RWMutex{},
		PreparedSQL: make([]string, 0, 100),
	}
	db.cacheStore.Store("preparedStmt", prepareStmt)

	if config.PrepareStmt {
		db.ConnPool = prepareStmt
	}

	db.Statement = &Statement{
		DB: db,
		ConnPool: db.ConnPool,
		Context: context.Background(),
		Clauses: map[string]clause.Clause{},
	}

	if err == nil && !config.DisableAutomaticPing {
		if pinger, ok := db.ConnPool.(interface{Ping() error}); ok {
			err = pinger.Ping()
		}
	}

	if err != nil {
		logger.Error("failed to initialize database, got error %v", err)
	}

	return
}

// Session create new db session
func (db *DB) Session(config *Session) *DB {
	var (
		txConfig = *db.Config
		tx = &DB{
			Config: &txConfig,
			Statement: db.Statement,
			Error: db.Error,
			clone: 1,
		}
	)
	if config.CreateBatchSize > 0 {
		tx.Config.CreateBatchSize = config.CreateBatchSize
	}

	if config.SkipDefaultTransaction {
		tx.Config.SkipDefaultTransaction = true
	}

	if config.AllowGlobalUpdate {
		txConfig.AllowGlobalUpdate = true
	}

	if config.FullSaveAssociation {
		txConfig.FullSaveAssociations = true
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
				ConnPool: db.Config.ConnPool,
				Mux: preparedStmt.Mux,
				Stmts: preparedStmt.Stmts,
			}
			txConfig.ConnPool = tx.Statement.ConnPool
			txConfig.PrepareStmt = true
		}
	}

	if config.SkipHooks {
		tx.Statement.SkipHooks = true
	}

	if config.DisableNestedTransaction {
		txConfig.DisableNestedTransaction = true
	}

	if !config.NewDB {
		tx.clone = 2
	}

	if config.DryRun {
		tx.Config.DryRun = true
	}

	if config.QueryFields {
		tx.Config.QueryFields = true
	}

	if config.NowFunc != nil {
		tx.Config.NowFunc = config.NowFunc
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
	return db.Statement.Settings.Load(fmt.Sprintf("%p", db.Statement)+key)
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
		tx := &DB{Config: db.Config}

		if db.clone == 1 {
			// clone with new statment
			tx.Statement = &Statement{
				DB:tx,
				ConnPool: db.Statement.ConnPool,
				Context: db.Statement.Context,
				Clauses: map[string]clause.Clause{},
				Vars: make([]interface{}, 0, 8),
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
		tx = db.getInstance()
		stmt = tx.Statement
		modelSchema, joinSchema  *schema.Schema
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
