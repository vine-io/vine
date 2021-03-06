package tests

import (
	"database/sql"
	"time"
)

// User has one `Account` (has one), many `Pets` (has many) and `Toys` (has many - polymorphic)
// He works in a Company (belongs to), he has a Manager (belongs to - single-table), and also managed a Team (has many - single-table)
// He speaks many languages (many to many) and has many friends (many to many - single-table)
// His pet also has one Toy (has one - polymorphic)
type User struct {
	Name      string
	Age       uint
	Birthday  *time.Time
	Account   Account
	Pets      []*Pet
	Toys      []Toy `dao:"polymorphic:Owner"`
	CompanyID *int
	Company   Company
	ManagerID *uint
	Manager   *User
	Team      []User     `dao:"foreignkey:ManagerID"`
	Languages []Language `dao:"many2many:UserSpeak;"`
	Friends   []*User    `dao:"many2many:user_friends;"`
	Active    bool
}

type Account struct {
	UserID sql.NullInt64
	Number string
}

type Pet struct {
	UserID *uint
	Name   string
	Toy    Toy `dao:"polymorphic:Owner;"`
}

type Toy struct {
	Name      string
	OwnerID   string
	OwnerType string
}

type Company struct {
	ID   int
	Name string
}

type Language struct {
	Code string `dao:"primarykey"`
	Name string
}
