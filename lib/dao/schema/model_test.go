package schema_test

import (
	"database/sql"
	"time"

	"github.com/lack-io/vine/lib/dao"
	"github.com/lack-io/vine/lib/dao/utils/tests"
)

type User struct {
	Name      *string
	Age       *uint
	Birthday  *time.Time
	Account   *tests.Account
	Pets      []*tests.Pet
	Toys      []*tests.Toy `dao:"polymorphic:Owner"`
	CompanyID *int
	Company   *tests.Company
	ManagerID *uint
	Manager   *User
	Team      []*User           `dao:"foreignkey:ManagerID"`
	Languages []*tests.Language `dao:"many2many:UserSpeak"`
	Friends   []*User           `dao:"many2many:user_friends"`
	Active    *bool
}

type mytime time.Time
type myint int
type mybool = bool

type AdvancedDataTypeUser struct {
	ID           sql.NullInt64
	Name         *sql.NullString
	Birthday     sql.NullTime
	RegisteredAt mytime
	DeletedAt    *mytime
	Active       mybool
	Admin        *mybool
}

type BaseModel struct {
	ID        uint
	CreatedAt time.Time
	CreatedBy *int
	Created   *VersionUser `dao:"foreignKey:CreatedBy"`
	UpdatedAt time.Time
	DeletedAt dao.DeletedAt `dao:"index"`
}

type VersionModel struct {
	BaseModel
	Version int
}

type VersionUser struct {
	VersionModel
	Name     string
	Age      uint
	Birthday *time.Time
}
