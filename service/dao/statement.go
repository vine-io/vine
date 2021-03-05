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
	"database/sql/driver"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/lack-io/vine/service/dao/clause"
	"github.com/lack-io/vine/service/dao/schema"
)

// Statement statement
type Statement struct {
	*DB
	TableExpr            *clause.Expr
	Table                string
	Model                interface{}
	Unscoped             bool
	Dest                 interface{}
	ReflectValue         reflect.Value
	Clauses              map[string]clause.Clause
	Distinct             map[string]clause.Clause
	Selects              []string
	Omits                []string
	Joins                []join
	Preloads             map[string][]interface{}
	Settings             sync.Map
	ConnPool             ConnPool
	Schema               *schema.Schema
	Context              context.Context
	RaiseErrorOnNotFound bool
	SkipHooks            bool
	SQL                  strings.Builder
	Vars                 []interface{}
	CurDestIndex         int
	attrs                []interface{}
	assigns              []interface{}
	scopes               []func(*DB) *DB
}

type join struct {
	Name  string
	Conds []interface{}
}

// StatementModifier statement modifier interface
type StatementModifier interface {
	ModifyStatement(*Statement)
}

// Write write string
func (stmt *Statement) WriteString(str string) (int, error) {
	return stmt.SQL.WriteString(str)
}

// Write write string
func (stmt *Statement) WriteByte(c byte) error {
	return stmt.SQL.WriteByte(c)
}

// WriteQuoted write quoted value
func (stmt *Statement) WriteQuoted(value interface{}) {
	stmt.QuoteTo(&stmt.SQL, value)
}

// QuoteTo write quoted value to writer
func (stmt *Statement) QuoteTo(writer clause.Writer, field interface{}) {
	switch v := field.(type) {
	case clause.Table:
		if v.Name == clause.CurrentTable {
			if stmt.TableExpr != nil {
				stmt.TableExpr.Build(stmt)
			} else {
				stmt.DB.Dialector.QuoteTo(writer, stmt.Table)
			}
		} else if v.Raw {
			writer.WriteString(v.Name)
		} else {
			stmt.DB.Dialector.QuoteTo(writer, v.Name)
		}

		if v.Alias != "" {
			writer.WriteByte(' ')
			stmt.DB.Dialector.QuoteTo(writer, v.Alias)
		}
	case clause.Column:
		if v.Table != "" {
			if v.Table == clause.CurrentTable {
				stmt.DB.Dialector.QuoteTo(writer, stmt.Table)
			} else {
				stmt.DB.Dialector.QuoteTo(writer, v.Table)
			}
			writer.WriteByte('.')
		}

		if v.Name == clause.PrimaryKey {
			if stmt.Schema == nil {
				stmt.DB.AddError(ErrModelValueRequired)
			} else if stmt.Schema.PrioritizedPrimaryField != nil {
				stmt.DB.Dialector.QuoteTo(writer, stmt.Schema.PrioritizedPrimaryField.DBName)
			} else if len(stmt.Schema.DBNames) > 0 {
				stmt.DB.Dialector.QuoteTo(writer, stmt.Schema.DBNames[0])
			}
		} else if v.Raw {
			writer.WriteString(v.Name)
		} else {
			stmt.DB.Dialector.QuoteTo(writer, v.Name)
		}

		if v.Alias != "" {
			writer.WriteString(" AS ")
			stmt.DB.Dialector.QuoteTo(writer, v.Alias)
		}
	case []clause.Column:
		writer.WriteByte('(')
		for idx, d := range v {
			if idx > 0 {
				writer.WriteString(",")
			}
			stmt.QuoteTo(writer, d)
		}
		writer.WriteByte(')')
	case string:
		stmt.DB.Dialector.QuoteTo(writer, v)
	case []string:
		writer.WriteByte('(')
		for idx, d := range v {
			if idx > 0 {
				writer.WriteString(",")
			}
			stmt.DB.Dialector.QuoteTo(writer, d)
		}
		writer.WriteByte(')')
	default:
		stmt.DB.Dialector.QuoteTo(writer, fmt.Sprint(field))
	}
}

// Quote returns quoted value
func (stmt *Statement) Quote(field interface{}) string {
	var builder strings.Builder
	stmt.QuoteTo(&builder, field)
	return builder.String()
}

// Write write string
func (stmt *Statement) AddVar(writer clause.Writer, vars ...interface{}) {
	for idx, v := range vars {
		if idx > 0 {
			writer.WriteByte(',')
		}

		switch v := v.(type) {
		case sql.NamedArg:
			stmt.Vars = append(stmt.Vars, v.Value)
		case clause.Column, clause.Table:
			stmt.QuoteTo(writer, v)
		case Valuer:
			stmt.AddVar(writer, v.DaoValue(stmt.Context, stmt.DB))
		case clause.Expr:
			v.Build(stmt)
		case driver.Valuer:
			stmt.Vars = append(stmt.Vars, v)
			stmt.DB.Dialector.BindVarTo(writer, stmt, v)
		case []byte:
			stmt.Vars = append(stmt.Vars, v)
			stmt.DB.Dialector.BindVarTo(writer, stmt, v)
		case []interface{}:
			if len(v) > 0 {
				writer.WriteByte('(')
				stmt.AddVar(writer, v...)
				stmt.WriteByte(')')
			} else {
				writer.WriteString("(NULL)")
			}
		case *DB:
			subdb := v.Session(&Session{DryRun: true}).getInstance()
			if v.Statement.SQL.Len() > 0 {
				var (
					vars = subdb.Statement.Vars
					sql  = v.Statement.SQL.String()
				)

				subdb.Statement.Vars = make([]interface{}, 0, len(vars))
				for _, vv := range vars {
					subdb.Statement.Vars = append(subdb.Statement.Vars, vv)
					bindvar := strings.Builder{}
					v.Dialector.BindVarTo(&bindvar, subdb.Statement, vv)
					sql = strings.Replace(sql, bindvar.String(), "?", 1)
				}

				subdb.Statement.SQL.Reset()
				subdb.Statement.Vars = stmt.Vars
				if strings.Contains(sql, "@") {
					clause.NamedExpr{SQL: sql, Vars: vars}.Build(subdb.Statement)
				} else {
					clause.Expr{SQL: sql, Vars: vars}.Build(subdb.Statement)
				}
			} else {
				subdb.Statement.Vars = append(stmt.Vars, subdb.Statement.Vars...)
				subdb.callbacks.Query().Execute(subdb)
			}

			writer.WriteString(subdb.Statement.SQL.String())
			stmt.Vars = subdb.Statement.Vars
		default:
			switch rv := reflect.ValueOf(v); rv.Kind() {
			case reflect.Slice, reflect.Array:
				if rv.Len() == 0 {
					writer.WriteString("(NULL)")
				} else {
					writer.WriteByte('(')
					for i := 0; i < rv.Len(); i++ {
						if i > 0 {
							writer.WriteByte(',')
						}
						stmt.AddVar(writer, rv.Index(i).Interface())
					}
					writer.WriteByte(')')
				}
			default:
				stmt.Vars = append(stmt.Vars, v)
				stmt.DB.Dialector.BindVarTo(writer, stmt, v)
			}
		}
	}
}

// AddClause add clause
func (stmt *Statement) AddClause(v clause.Interface) {
	if optimizer, ok := v.(StatementModifier); ok {
		optimizer.ModifyStatement(stmt)
	} else {
		name := v.Name()
		c := stmt.Clauses[name]
		c.Name = name
		v.MergeClause(&c)
		stmt.Clauses[name] = c
	}
}

// AddClauseIfNotExists add clause if not exists
func (stmt *Statement) AddClauseIfNotExists(v clause.Interface) {
	if c, ok := stmt.Clauses[v.Name()]; !ok || c.Expression == nil {
		stmt.AddClause(v)
	}
}

// BuildCondition build condition
func (stmt *Statement) BuildCondition(query interface{}, args ...interface{}) []clause.Expression {
	if s, ok := query.(string); ok {
		// if it is a number, then treats it as primary key
		if _, err := strconv.Atoi(s); err != nil {
			if s == "" && len(args) == 0 {
				return nil
			} else if len(args) == 0 || len(args) > 0 && strings.Contains(s, "?") {
				// looks like a where condition
				return []clause.Expression{clause.Expr{SQL: s, Vars: args}}
			} else if len(args) > 0 && strings.Contains(s, "@") {
				// looks like a named query
				return []clause.Expression{clause.NamedExpr{SQL: s, Vars: args}}
			} else if len(args) == 1 {
				return []clause.Expression{clause.Eq{Column: s, Value: args[0]}}
			}
		}
	}

	conds := make([]clause.Expression, 0, 4)
	args = append([]interface{}{query}, args...)
	for idx, arg := range args {
		if valuer, ok := arg.(driver.Valuer); ok {
			arg, _ = valuer.Value()
		}

		switch v := arg.(type) {
		case clause.Expression:
			conds = append(conds, v)
		case *DB:
			if cs, ok := v.Statement.Clauses["WHERE"]; ok {
				if where, ok := cs.Expression.(clause.Where); ok {
					if len(where.Exprs) == 1 {
						if orConds, ok := where.Exprs[0].(clause.OrConditions); ok {
							where.Exprs[0] = clause.AndConditions{Exprs: orConds.Exprs}
						}
					}
					conds = append(conds, clause.And(where.Exprs...))
				} else if cs.Expression != nil {
					conds = append(conds, cs.Expression)
				}
			}
		case map[interface{}]interface{}:
			for i, j := range v {
				conds = append(conds, clause.Eq{Column: i, Value: j})
			}
		case map[string]string:
			var keys = make([]string, 0, len(v))
			for i := range v {
				keys = append(keys, i)
			}
			sort.Strings(keys)

			for _, key := range keys {
				conds = append(conds, clause.Eq{Column: key, Value: v[key]})
			}
		case map[string]interface{}:
			var keys = make([]string, 0, len(v))
			for i := range v {
				keys = append(keys, i)
			}
			sort.Strings(keys)

			for _, key := range keys {
				reflectValue := reflect.Indirect(reflect.ValueOf(v[key]))
				switch reflectValue.Kind() {
				case reflect.Slice, reflect.Array:
					if _, ok := v[key].(driver.Valuer); ok {
						conds = append(conds, clause.Eq{Column: key, Value: v[key]})
					} else if _, ok := v[key].(Valuer); ok {
						conds = append(conds, clause.Eq{Column: key, Value: v[key]})
					} else {
						values := make([]interface{}, reflectValue.Len())
						for i := 0; i < reflectValue.Len(); i++ {
							values[i] = reflectValue.Index(i).Interface()
						}

						conds = append(conds, clause.IN{Column: key, Values: values})
					}
				default:
					conds = append(conds, clause.Eq{Column: key, Value: v[key]})
				}
			}
		default:
			reflectValue := reflect.Indirect(reflect.ValueOf(arg))
			if s, err := schema.Parse(arg, stmt.DB.cacheStore, stmt.DB.NamingStrategy); err == nil {
				selectedColumns := map[string]bool{}
				if idx == 0 {
					for _, v := range args[1:] {
						if vs, ok := v.(string); ok {
							selectedColumns[vs] = true
						}
					}
				}
				restricted := len(selectedColumns) != 0

				switch reflectValue.Kind() {
				case reflect.Struct:
					for _, field := range s.Fields {
						selected := selectedColumns[field.DBName] || selectedColumns[field.Name]
						if selected || (!restricted && field.Readable) {
							if v, isZero := field.ValueOf(reflectValue); !isZero || selected {
								if field.DBName != "" {
									conds = append(conds, clause.Eq{Column: clause.Column{Table: clause.CurrentTable, Name: field.DBName}, Value: v})
								} else if field.DataType != "" {
									conds = append(conds, clause.Eq{Column: clause.Column{Table: clause.CurrentTable, Name: field.Name}, Value: v})
								}
							}
						}
					}
				case reflect.Slice, reflect.Array:
					for i := 0; i < reflectValue.Len(); i++ {
						for _, field := range s.Fields {
							selected := selectedColumns[field.DBName] || selectedColumns[field.Name]
							if selected || (!restricted && field.Readable) {
								if v, isZero := field.ValueOf(reflectValue.Index(i)); !isZero || selected {
									conds = append(conds, clause.Eq{Column: clause.Column{Table: clause.CurrentTable, Name: field.DBName}, Value: v})
								} else if field.DataType != "" {
									conds = append(conds, clause.Eq{Column: clause.Column{Table: clause.CurrentTable, Name: field.Name}, Value: v})
								}
							}
						}
					}
				}

				if restricted {
					break
				}
			} else if !reflectValue.IsValid() {
				stmt.AddError(ErrInvalidData)
			} else if len(conds) == 0 {
				if len(args) == 1 {
					switch reflectValue.Kind() {
					case reflect.Slice, reflect.Array:
						values := make([]interface{}, reflectValue.Len())
						for i := 0; i < reflectValue.Len(); i++ {
							values[i] = reflectValue.Index(i).Interface()
						}

						if len(values) > 0 {
							conds = append(conds, clause.IN{Column: clause.PrimaryColumn, Values: values})
						}
						return conds
					}
				}

				conds = append(conds, clause.IN{Column: clause.PrimaryColumn, Values: args})
			}
		}
	}

	return conds
}

// Build build sql with clauses names
func (stmt *Statement) Build(clauses ...string) {
	var firstClauseWritten bool

	for _, name := range clauses {
		if c, ok := stmt.Clauses[name]; ok {
			if firstClauseWritten {
				stmt.WriteByte(' ')
			}

			firstClauseWritten = true
			if b, ok := stmt.DB.ClauseBuilders[name]; ok {
				b(c, stmt)
			} else {
				c.Build(stmt)
			}
		}
	}
}

func (stmt *Statement) Parse(value interface{}) (err error) {
	if stmt.Schema, err = schema.Parse(value, stmt.DB.cacheStore, stmt.DB.NamingStrategy); err == nil && stmt.Table == "" {
		if tables := strings.Split(stmt.Schema.Table, "."); len(tables) == 2 {
			stmt.TableExpr = &clause.Expr{SQL: stmt.Quote(stmt.Schema.Table)}
			stmt.Table = tables[1]
			return
		}

		stmt.Table = stmt.Schema.Table
	}
	return err
}

func (stmt *Statement) clone() *Statement {
	newStmt := &Statement{

	}

	return newStmt
}
