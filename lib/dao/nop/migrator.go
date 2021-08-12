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
