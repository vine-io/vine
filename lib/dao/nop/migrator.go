// MIT License
//
// Copyright (c) 2021 Lack
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

package nop

import (
	"github.com/vine-io/vine/lib/dao"
	"github.com/vine-io/vine/lib/dao/clause"
	"github.com/vine-io/vine/lib/dao/schema"
)

type Migrator struct {
}

func (m Migrator) AutoMigrate(dst ...interface{}) error {
	return nil
}

func (m Migrator) CurrentDatabase() string {
	return ""
}

func (m Migrator) FullDataTypeOf(field *schema.Field) clause.Expr {
	return clause.Expr{}
}

func (m Migrator) CreateTable(dst ...interface{}) error {
	return nil
}

func (m Migrator) DropTable(dst ...interface{}) error {
	return nil
}

func (m Migrator) HasTable(dst interface{}) bool {
	return false
}

func (m Migrator) RenameTable(oldName, newName interface{}) error {
	return nil
}

func (m Migrator) AddColumn(dst interface{}, field string) error {
	return nil
}

func (m Migrator) DropColumn(dst interface{}, field string) error {
	return nil
}

func (m Migrator) AlterColumn(dst interface{}, field string) error {
	return nil
}

func (m Migrator) MigrateColumn(dst interface{}, field *schema.Field, columnType dao.ColumnType) error {
	return nil
}

func (m Migrator) HasColumn(dst interface{}, field string) bool {
	return false
}

func (m Migrator) RenameColumn(dst interface{}, oldName, field string) error {
	return nil
}

func (m Migrator) ColumnTypes(dst interface{}) ([]dao.ColumnType, error) {
	return []dao.ColumnType{}, nil
}

func (m Migrator) CreateView(name string, option dao.ViewOption) error {
	return nil
}

func (m Migrator) DropView(name string) error {
	return nil
}

func (m Migrator) CreateConstraint(dst interface{}, name string) error {
	return nil
}

func (m Migrator) DropConstraint(dst interface{}, name string) error {
	return nil
}

func (m Migrator) HasConstraint(dst interface{}, name string) bool {
	return false
}

func (m Migrator) CreateIndex(dst interface{}, name string) error {
	return nil
}

func (m Migrator) DropIndex(dst interface{}, name string) error {
	return nil
}

func (m Migrator) HasIndex(dst interface{}, name string) bool {
	return false
}

func (m Migrator) RenameIndex(dst interface{}, oldName, newName string) error {
	return nil
}
