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

package sqlite

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lack-io/vine/lib/dao"
	"github.com/lack-io/vine/lib/dao/clause"
	"github.com/lack-io/vine/lib/dao/migrator"
	"github.com/lack-io/vine/lib/dao/schema"
)

type Migrator struct {
	migrator.Migrator
}

func (m *Migrator) RunWithoutForeignKey(fc func() error) error {
	var enabled int
	m.DB.Raw("PRAGMA foreign_keys").Scan(&enabled)
	if enabled == 1 {
		m.DB.Exec("PRAGMA foreign_keys = OFF")
		defer m.DB.Exec("PRAGMA foreign_keys = ON")
	}

	return fc()
}

func (m Migrator) HasTable(value interface{}) bool {
	var count int
	m.Migrator.RunWithValue(value, func(stmt *dao.Statement) error {
		return m.DB.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", stmt.Table).Row().Scan(&count)
	})
	return count > 0
}

func (m Migrator) DropTable(values ...interface{}) error {
	return m.RunWithoutForeignKey(func() error {
		values = m.ReorderModels(values, false)
		tx := m.DB.Session(&dao.Session{})

		for i := len(values) - 1; i >= 0; i-- {
			if err := m.RunWithValue(values[i], func(stmt *dao.Statement) error {
				return tx.Exec("DROP TABLE IF EXISTS ?", clause.Table{Name: stmt.Table}).Error
			}); err != nil {
				return err
			}
		}

		return nil
	})

	return nil
}

func (m Migrator) HasColumn(value interface{}, name string) bool {
	var count int
	m.Migrator.RunWithValue(value, func(stmt *dao.Statement) error {
		if field := stmt.Schema.LookUpField(name); field != nil {
			name = field.DBName
		}

		if name != "" {
			m.DB.Raw(
				"SELECT count(*) FROM sqlite_master WHERE type = ? AND tbl_name = ? AND (sql LIKE ? OR sql LIKE ? OR sql LIKE ?)",
				"table", stmt.Table, `%"`+name+`" %`, `%`+name+` %`, "%`"+name+"`%",
			).Row().Scan(&count)
		}
		return nil
	})
	return count > 0
}

func (m Migrator) AlterColumn(value interface{}, name string) error {
	return m.RunWithoutForeignKey(func() error {
		return m.RunWithValue(value, func(stmt *dao.Statement) error {
			if field := stmt.Schema.LookUpField(name); field != nil {
				var (
					createSQL    string
					newTableName = stmt.Table + "__temp"
				)

				m.DB.Raw("SELECT sql FROM sqlite_master WHERE type = ? AND tbl_name = ? AND name = ?", "table", stmt.Table, stmt.Table).Row().Scan(&createSQL)

				if reg, err := regexp.Compile("(`|'|\"| )" + field.DBName + "(`|'|\"| ) .*?,"); err == nil {
					tableReg, err := regexp.Compile(" ('|`|\"| )" + stmt.Table + "('|`|\"| ) ")
					if err != nil {
						return err
					}

					createSQL = tableReg.ReplaceAllString(createSQL, fmt.Sprintf(" `%v` ", newTableName))
					createSQL = reg.ReplaceAllString(createSQL, fmt.Sprintf("`%v` ?,", field.DBName))

					var columns []string
					columnTypes, _ := m.DB.Migrator().ColumnTypes(value)
					for _, columnType := range columnTypes {
						columns = append(columns, fmt.Sprintf("`%v`", columnType.Name()))
					}

					return m.DB.Transaction(func(tx *dao.DB) error {
						queries := []string{
							createSQL,
							fmt.Sprintf("INSERT INTO `%v`(%v) SELECT %v FROM `%v`", newTableName, strings.Join(columns, ","), strings.Join(columns, ","), stmt.Table),
							fmt.Sprintf("DROP TABLE `%v`", stmt.Table),
							fmt.Sprintf("ALTER TABLE `%v` RENAME TO `%v`", newTableName, stmt.Table),
						}
						for _, query := range queries {
							if err := tx.Exec(query, m.FullDataTypeOf(field)).Error; err != nil {
								return err
							}
						}
						return nil
					})
				} else {
					return err
				}
			} else {
				return fmt.Errorf("failed to alter field with name %v", name)
			}
		})
	})
}

func (m Migrator) DropColumn(value interface{}, name string) error {
	return m.RunWithValue(value, func(stmt *dao.Statement) error {
		if field := stmt.Schema.LookUpField(name); field != nil {
			name = field.DBName
		}

		var (
			createSQL    string
			newTableName = stmt.Table + "__temp"
		)

		m.DB.Raw("SELECT sql FROM sqlite_master WHERE type = ? AND tbl_name = ? AND name = ?", "table", stmt.Table, stmt.Table).Row().Scan(&createSQL)

		if reg, err := regexp.Compile("(`|'|\"| )" + name + "(`|'|\"| ) .*?,"); err == nil {
			tableReg, err := regexp.Compile(" ('|`|\"| )" + stmt.Table + "('|`|\"| ) ")
			if err != nil {
				return err
			}

			createSQL = tableReg.ReplaceAllString(createSQL, fmt.Sprintf(" `%v` ", newTableName))
			createSQL = reg.ReplaceAllString(createSQL, "")

			var columns []string
			columnTypes, _ := m.DB.Migrator().ColumnTypes(value)
			for _, columnType := range columnTypes {
				if columnType.Name() != name {
					columns = append(columns, fmt.Sprintf("`%v`", columnType.Name()))
				}
			}

			return m.DB.Transaction(func(tx *dao.DB) error {
				queries := []string{
					createSQL,
					fmt.Sprintf("INSERT INTO `%v`(%v) SELECT %v FROM `%v`", newTableName, strings.Join(columns, ","), strings.Join(columns, ","), stmt.Table),
					fmt.Sprintf("DROP TABLE `%v`", stmt.Table),
					fmt.Sprintf("ALTER TABLE `%v` RENAME TO `%v`", newTableName, stmt.Table),
				}
				for _, query := range queries {
					if err := tx.Exec(query).Error; err != nil {
						return err
					}
				}
				return nil
			})
		} else {
			return err
		}
	})
}

func (m Migrator) CreateConstraint(interface{}, string) error {
	return ErrConstraintsNotImplemented
}

func (m Migrator) DropConstraint(interface{}, string) error {
	return ErrConstraintsNotImplemented
}

func (m Migrator) HasConstraint(value interface{}, name string) bool {
	var count int64
	m.RunWithValue(value, func(stmt *dao.Statement) error {
		m.DB.Raw(
			"SELECT count(*) FROM sqlite_master WHERE type = ? AND tbl_name = ? AND (sql LIKE ? OR sql LIKE ? OR sql LIKE ?)",
			"table", stmt.Table, `%CONSTRAINT "`+name+`" %`, `%CONSTRAINT `+name+` %`, "%CONSTRAINT `"+name+"`%",
		).Row().Scan(&count)

		return nil
	})

	return count > 0
}

func (m Migrator) CurrentDatabase() (name string) {
	var null interface{}
	m.DB.Raw("PRAGMA database_list").Row().Scan(&null, &name, &null)
	return
}

func (m Migrator) BuildIndexOptions(opts []schema.IndexOption, stmt *dao.Statement) (results []interface{}) {
	for _, opt := range opts {
		str := stmt.Quote(opt.DBName)
		if opt.Expression != "" {
			str = opt.Expression
		}

		if opt.Collate != "" {
			str += " COLLATE " + opt.Collate
		}

		if opt.Sort != "" {
			str += " " + opt.Sort
		}
		results = append(results, clause.Expr{SQL: str})
	}
	return
}

func (m Migrator) CreateIndex(value interface{}, name string) error {
	return m.RunWithValue(value, func(stmt *dao.Statement) error {
		if idx := stmt.Schema.LookIndex(name); idx != nil {
			opts := m.BuildIndexOptions(idx.Fields, stmt)
			values := []interface{}{clause.Column{Name: idx.Name}, clause.Table{Name: stmt.Table}, opts}

			createIndexSQL := "CREATE "
			if idx.Class != "" {
				createIndexSQL += idx.Class + " "
			}
			createIndexSQL += "INDEX ?"

			if idx.Type != "" {
				createIndexSQL += " USING " + idx.Type
			}
			createIndexSQL += " ON ??"

			if idx.Where != "" {
				createIndexSQL += " WHERE " + idx.Where
			}

			return m.DB.Exec(createIndexSQL, values...).Error
		}

		return fmt.Errorf("failed to create index with name %v", name)
	})
}

func (m Migrator) HasIndex(value interface{}, name string) bool {
	var count int
	m.RunWithValue(value, func(stmt *dao.Statement) error {
		if idx := stmt.Schema.LookIndex(name); idx != nil {
			name = idx.Name
		}

		if name != "" {
			m.DB.Raw(
				"SELECT count(*) FROM sqlite_master WHERE type = ? AND tbl_name = ? AND name = ?", "index", stmt.Table, name,
			).Row().Scan(&count)
		}
		return nil
	})
	return count > 0
}

func (m Migrator) RenameIndex(value interface{}, oldName, newName string) error {
	return m.RunWithValue(value, func(stmt *dao.Statement) error {
		var sql string
		m.DB.Raw("SELECT sql FROM sqlite_master WHERE type = ? AND tbl_name = ? AND name = ?", "index", stmt.Table, oldName).Row().Scan(&sql)
		if sql != "" {
			return m.DB.Exec(strings.Replace(sql, oldName, newName, 1)).Error
		}
		return fmt.Errorf("failed to find index with name %v", oldName)
	})
}

func (m Migrator) DropIndex(value interface{}, name string) error {
	return m.RunWithValue(value, func(stmt *dao.Statement) error {
		if idx := stmt.Schema.LookIndex(name); idx != nil {
			name = idx.Name
		}

		return m.DB.Exec("DROP INDEX ?", clause.Column{Name: name}).Error
	})
}
