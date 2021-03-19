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

package sqlite

import (
	"database/sql"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/lack-io/vine/service/dao"
	"github.com/lack-io/vine/service/dao/callbacks"
	"github.com/lack-io/vine/service/dao/clause"
	"github.com/lack-io/vine/service/dao/logger"
	"github.com/lack-io/vine/service/dao/migrator"
	"github.com/lack-io/vine/service/dao/schema"
)

// DefaultDriverName is the default driver name for SQLite.
const DefaultDriverName = "sqlite3"

type Dialect struct {
	DB         *dao.DB
	Opts       dao.Options
	DriverName string
	Conn       dao.ConnPool
}

func newSQLiteDialect(opts ...dao.Option) dao.Dialect {
	options := dao.NewOptions(opts...)

	for _, opt := range opts {
		opt(&options)
	}

	dialect := &Dialect{
		Opts: options,
		Conn: options.ConnPool,
	}

	if name, ok := options.Context.Value(driverNameKey{}).(string); ok {
		dialect.DriverName = name
	} else {
		dialect.DriverName = DefaultDriverName
	}

	return dialect
}

func (d *Dialect) Init(opts ...dao.Option) (err error) {
	for _, opt := range opts {
		opt(&d.Opts)
	}

	if name, ok := d.Opts.Context.Value(driverNameKey{}).(string); ok {
		d.DriverName = name
	} else {
		d.DriverName = DefaultDriverName
	}

	if d.DB == nil {
		d.DB, err = dao.Open(d)
		if err != nil {
			return err
		}
	}

	callbacks.RegisterDefaultCallbacks(d.DB, &callbacks.Options{
		LastInsertIDReversed: true,
	})

	if d.Conn != nil {
		d.DB.ConnPool = d.Conn
	} else {
		d.DB.ConnPool, err = sql.Open(d.DriverName, d.Opts.DSN)
		if err != nil {
			return err
		}
	}

	d.DB.Statement.ConnPool = d.DB.ConnPool

	for k, v := range d.ClauseBuilders() {
		d.DB.ClauseBuilders[k] = v
	}
	return nil
}

func (d *Dialect) Options() dao.Options {
	return d.Opts
}

func (d *Dialect) NewTx() *dao.DB {
	return d.DB.Session(&dao.Session{})
}

func (d *Dialect) Migrator() dao.Migrator {
	return Migrator{migrator.Migrator{
		Options: migrator.Options{
			DB:                          d.DB,
			Dialect:                     d,
			CreateIndexAfterCreateTable: true,
		},
	}}
}

func (d *Dialect) DataTypeOf(field *schema.Field) string {
	switch field.DataType {
	case schema.Bool:
		return "numeric"
	case schema.Int, schema.Uint:
		if field.AutoIncrement && !field.PrimaryKey {
			return "integer PRIMARY KEY AUTOINCREMENT"
		} else {
			return "integer"
		}
	case schema.Float:
		return "real"
	case schema.String:
		return "text"
	case schema.Time:
		return "datetime"
	case schema.Bytes:
		return "blob"
	}
	return string(field.DataType)
}

func (d *Dialect) DefaultValueOf(field *schema.Field) clause.Expression {
	if field.AutoIncrement {
		return clause.Expr{SQL: "NULL"}
	}

	// doesn't work, will raise error
	return clause.Expr{SQL: "DEFAULT"}
}

func (d *Dialect) BindVarTo(writer clause.Writer, stmt *dao.Statement, v interface{}) {
	writer.WriteByte('?')
}

func (d *Dialect) QuoteTo(writer clause.Writer, str string) {
	writer.WriteByte('`')
	if strings.Contains(str, ".") {
		for idx, str := range strings.Split(str, ".") {
			if idx > 0 {
				writer.WriteString(".`")
			}
			writer.WriteString(str)
			writer.WriteByte('`')
		}
	} else {
		writer.WriteString(str)
		writer.WriteByte('`')
	}
}

func (d *Dialect) Explain(sql string, vars ...interface{}) string {
	return logger.ExplainSQL(sql, nil, `"`, vars...)
}

func (d *Dialect) SavePoint(tx *dao.DB, name string) error {
	tx.Exec("SAVEPOINT " + name)
	return nil
}

func (d *Dialect) RollbackTo(tx *dao.DB, name string) error {
	tx.Exec("ROLLBACK TO SAVEPOINT " + name)
	return nil
}

func (d *Dialect) ClauseBuilders() map[string]clause.ClauseBuilder {
	return map[string]clause.ClauseBuilder{
		"INSERT": func(c clause.Clause, builder clause.Builder) {
			if insert, ok := c.Expression.(clause.Insert); ok {
				if stmt, ok := builder.(*dao.Statement); ok {
					stmt.WriteString("INSERT ")
					if insert.Modifier != "" {
						stmt.WriteString(insert.Modifier)
						stmt.WriteByte(' ')
					}

					stmt.WriteString("INTO ")
					if insert.Table.Name == "" {
						stmt.WriteQuoted(stmt.Table)
					} else {
						stmt.WriteQuoted(insert.Table)
					}
					return
				}
			}

			c.Build(builder)
		},
		"LIMIT": func(c clause.Clause, builder clause.Builder) {
			if limit, ok := c.Expression.(clause.Limit); ok {
				if limit.Limit > 0 {
					builder.WriteString("LIMIT ")
					builder.WriteString(strconv.Itoa(limit.Limit))
				}
				if limit.Offset > 0 {
					if limit.Limit > 0 {
						builder.WriteString(" ")
					}
					builder.WriteString("OFFSET ")
					builder.WriteString(strconv.Itoa(limit.Offset))
				}
			}
		},
		"FOR": func(c clause.Clause, builder clause.Builder) {
			if _, ok := c.Expression.(clause.Locking); ok {
				// SQLite3 does not support row-level locking.
				return
			}
			c.Build(builder)
		},
	}
}

func (d *Dialect) JSONDataType() string {
	return "JSON"
}

func (d *Dialect) JSONBuild(tx *dao.DB, column string) dao.JSONQuery {
	return JSONQuery(tx, column)
}

func (d *Dialect) String() string {
	return "sqlite3"
}

func NewDialect(opts ...dao.Option) dao.Dialect {
	return newSQLiteDialect(opts...)
}
