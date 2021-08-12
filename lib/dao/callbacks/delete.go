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

package callbacks

import (
	"reflect"
	"strings"

	"github.com/vine-io/vine/lib/dao"
	"github.com/vine-io/vine/lib/dao/clause"
	"github.com/vine-io/vine/lib/dao/schema"
)

func BeforeDelete(db *dao.DB) {
	if db.Error == nil && db.Statement.Schema != nil && !db.Statement.SkipHooks && db.Statement.Schema.BeforeDelete {
		callMethod(db, func(value interface{}, tx *dao.DB) bool {
			if i, ok := value.(BeforeDeleteInterface); ok {
				db.AddError(i.BeforeDelete(tx))
				return true
			}

			return false
		})
	}
}

func DeleteBeforeAssociations(db *dao.DB) {
	if db.Error == nil && db.Statement.Schema != nil {
		selectColumns, restricted := db.Statement.SelectAndOmitColumns(true, false)

		if restricted {
			for column, v := range selectColumns {
				if v {
					if rel, ok := db.Statement.Schema.Relationships.Relations[column]; ok {
						switch rel.Type {
						case schema.HasOne, schema.HasMany:
							queryConds := rel.ToQueryConditions(db.Statement.ReflectValue)
							modelValue := reflect.New(rel.FieldSchema.ModelType).Interface()
							tx := db.Session(&dao.Session{NewDB: true}).Model(modelValue)
							withoutConditions := false
							if db.Statement.Unscoped {
								tx = tx.Unscoped()
							}

							if len(db.Statement.Selects) > 0 {
								var selects []string
								for _, s := range db.Statement.Selects {
									if s == clause.Associations {
										selects = append(selects, s)
									} else if strings.HasPrefix(s, column+".") {
										selects = append(selects, strings.TrimPrefix(s, column+"."))
									}
								}

								if len(selects) > 0 {
									tx = tx.Select(selects)
								}
							}

							for _, cond := range queryConds {
								if c, ok := cond.(clause.IN); ok && len(c.Values) == 0 {
									withoutConditions = true
									break
								}
							}

							if !withoutConditions {
								if db.AddError(tx.Clauses(clause.Where{Exprs: queryConds}).Delete(modelValue).Error) != nil {
									return
								}
							}
						case schema.Many2Many:
							var (
								queryConds     []clause.Expression
								foreignFields  []*schema.Field
								relForeignKeys []string
								modelValue     = reflect.New(rel.JoinTable.ModelType).Interface()
								table          = rel.JoinTable.Table
								tx             = db.Session(&dao.Session{NewDB: true}).Model(modelValue).Table(table)
							)

							for _, ref := range rel.References {
								if ref.OwnPrimaryKey {
									foreignFields = append(foreignFields, ref.PrimaryKey)
									relForeignKeys = append(relForeignKeys, ref.ForeignKey.DBName)
								} else if ref.PrimaryValue != "" {
									queryConds = append(queryConds, clause.Eq{
										Column: clause.Column{Table: rel.JoinTable.Table, Name: ref.ForeignKey.DBName},
										Value:  ref.PrimaryValue,
									})
								}
							}

							_, foreignValues := schema.GetIdentityFieldValuesMap(db.Statement.ReflectValue, foreignFields)
							column, values := schema.ToQueryValues(table, relForeignKeys, foreignValues)
							queryConds = append(queryConds, clause.IN{Column: column, Values: values})

							if db.AddError(tx.Clauses(clause.Where{Exprs: queryConds}).Delete(modelValue).Error) != nil {
								return
							}
						}
					}
				}
			}
		}
	}
}

func Delete(db *dao.DB) {
	if db.Error == nil {
		if db.Statement.Schema != nil && !db.Statement.Unscoped {
			for _, c := range db.Statement.Schema.DeleteClauses {
				db.Statement.AddClause(c)
			}
		}

		if db.Statement.SQL.String() == "" {
			db.Statement.SQL.Grow(100)
			db.Statement.AddClauseIfNotExists(clause.Delete{})

			if db.Statement.Schema != nil {
				_, queryValues := schema.GetIdentityFieldValuesMap(db.Statement.ReflectValue, db.Statement.Schema.PrimaryFields)
				column, values := schema.ToQueryValues(db.Statement.Table, db.Statement.Schema.PrimaryFieldDBNames, queryValues)

				if len(values) > 0 {
					db.Statement.AddClause(clause.Where{Exprs: []clause.Expression{clause.IN{Column: column, Values: values}}})
				}

				if db.Statement.ReflectValue.CanAddr() && db.Statement.Dest != db.Statement.Model && db.Statement.Model != nil {
					_, queryValues = schema.GetIdentityFieldValuesMap(reflect.ValueOf(db.Statement.Model), db.Statement.Schema.PrimaryFields)
					column, values = schema.ToQueryValues(db.Statement.Table, db.Statement.Schema.PrimaryFieldDBNames, queryValues)

					if len(values) > 0 {
						db.Statement.AddClause(clause.Where{Exprs: []clause.Expression{clause.IN{Column: column, Values: values}}})
					}
				}
			}

			db.Statement.AddClauseIfNotExists(clause.From{})
			db.Statement.Build("DELETE", "FROM", "WHERE")
		}

		if _, ok := db.Statement.Clauses["WHERE"]; !db.AllowGlobalUpdate && !ok && db.Error == nil {
			db.AddError(dao.ErrMissingWhereClause)
			return
		}

		if !db.DryRun && db.Error == nil {
			result, err := db.Statement.ConnPool.ExecContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)

			if err == nil {
				db.RowsAffected, _ = result.RowsAffected()
			} else {
				db.AddError(err)
			}
		}
	}
}

func AfterDelete(db *dao.DB) {
	if db.Error == nil && db.Statement.Schema != nil && !db.Statement.SkipHooks && db.Statement.Schema.AfterDelete {
		callMethod(db, func(value interface{}, tx *dao.DB) bool {
			if i, ok := value.(AfterDeleteInterface); ok {
				db.AddError(i.AfterDelete(tx))
				return true
			}
			return false
		})
	}
}
