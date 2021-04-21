package schema_test

import (
	"database/sql"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/lack-io/vine/lib/dao/schema"
	"github.com/lack-io/vine/lib/dao/utils/tests"
)

func TestFieldValuerAndSetter(t *testing.T) {
	var (
		userSchema, _ = schema.Parse(&tests.User{}, &sync.Map{}, schema.NamingStrategy{})
		user          = tests.User{
			Name:     "valuer_and_setter",
			Age:      18,
			Birthday: tests.Now(),
			Active:   true,
		}
		reflectValue = reflect.ValueOf(&user)
	)

	// test valuer
	values := map[string]interface{}{
		"name":     user.Name,
		"age":      user.Age,
		"birthday": user.Birthday,
		"active":   true,
	}
	checkField(t, userSchema, reflectValue, values)

	var f *bool
	// test setter
	newValues := map[string]interface{}{
		"name":       "valuer_and_setter_2",
		"id":         2,
		"created_at": time.Now(),
		"updated_at": nil,
		"deleted_at": time.Now(),
		"age":        20,
		"birthday":   time.Now(),
		"active":     f,
	}

	for k, v := range newValues {
		if err := userSchema.FieldsByDBName[k].Set(reflectValue, v); err != nil {
			t.Errorf("no error should happen when assign value to field %v, but got %v", k, err)
		}
	}
	newValues["updated_at"] = time.Time{}
	newValues["active"] = false
	checkField(t, userSchema, reflectValue, newValues)

	// test valuer and other type
	age := myint(10)
	var nilTime *time.Time
	newValues2 := map[string]interface{}{
		"name":       sql.NullString{String: "valuer_and_setter_3", Valid: true},
		"id":         &sql.NullInt64{Int64: 3, Valid: true},
		"created_at": tests.Now(),
		"updated_at": nilTime,
		"deleted_at": time.Now(),
		"age":        &age,
		"birthday":   mytime(time.Now()),
		"active":     mybool(true),
	}

	for k, v := range newValues2 {
		if err := userSchema.FieldsByDBName[k].Set(reflectValue, v); err != nil {
			t.Errorf("no error should happen when assign value to field %v, but got %v", k, err)
		}
	}
	newValues2["updated_at"] = time.Time{}
	checkField(t, userSchema, reflectValue, newValues2)
}

func TestPointerFieldValuerAndSetter(t *testing.T) {
	var (
		userSchema, _      = schema.Parse(&User{}, &sync.Map{}, schema.NamingStrategy{})
		name               = "pointer_field_valuer_and_setter"
		age           uint = 18
		active             = true
		user               = User{
			Name:     &name,
			Age:      &age,
			Birthday: tests.Now(),
			Active:   &active,
		}
		reflectValue = reflect.ValueOf(&user)
	)

	// test valuer
	values := map[string]interface{}{
		"name":     user.Name,
		"age":      user.Age,
		"birthday": user.Birthday,
		"active":   true,
	}
	checkField(t, userSchema, reflectValue, values)

	// test setter
	newValues := map[string]interface{}{
		"name":       "valuer_and_setter_2",
		"id":         2,
		"created_at": time.Now(),
		"deleted_at": time.Now(),
		"age":        20,
		"birthday":   time.Now(),
		"active":     false,
	}

	for k, v := range newValues {
		if err := userSchema.FieldsByDBName[k].Set(reflectValue, v); err != nil {
			t.Errorf("no error should happen when assign value to field %v, but got %v", k, err)
		}
	}
	checkField(t, userSchema, reflectValue, newValues)

	// test valuer and other type
	age2 := myint(10)
	newValues2 := map[string]interface{}{
		"name":       sql.NullString{String: "valuer_and_setter_3", Valid: true},
		"id":         &sql.NullInt64{Int64: 3, Valid: true},
		"created_at": tests.Now(),
		"deleted_at": time.Now(),
		"age":        &age2,
		"birthday":   mytime(time.Now()),
		"active":     mybool(true),
	}

	for k, v := range newValues2 {
		if err := userSchema.FieldsByDBName[k].Set(reflectValue, v); err != nil {
			t.Errorf("no error should happen when assign value to field %v, but got %v", k, err)
		}
	}
	checkField(t, userSchema, reflectValue, newValues2)
}

func TestAdvancedDataTypeValuerAndSetter(t *testing.T) {
	var (
		userSchema, _ = schema.Parse(&AdvancedDataTypeUser{}, &sync.Map{}, schema.NamingStrategy{})
		name          = "advanced_data_type_valuer_and_setter"
		deletedAt     = mytime(time.Now())
		isAdmin       = mybool(false)
		user          = AdvancedDataTypeUser{
			ID:           sql.NullInt64{Int64: 10, Valid: true},
			Name:         &sql.NullString{String: name, Valid: true},
			Birthday:     sql.NullTime{Time: time.Now(), Valid: true},
			RegisteredAt: mytime(time.Now()),
			DeletedAt:    &deletedAt,
			Active:       mybool(true),
			Admin:        &isAdmin,
		}
		reflectValue = reflect.ValueOf(&user)
	)

	// test valuer
	values := map[string]interface{}{
		"id":            user.ID,
		"name":          user.Name,
		"birthday":      user.Birthday,
		"registered_at": user.RegisteredAt,
		"deleted_at":    user.DeletedAt,
		"active":        user.Active,
		"admin":         user.Admin,
	}
	checkField(t, userSchema, reflectValue, values)

	// test setter
	newDeletedAt := mytime(time.Now())
	newIsAdmin := mybool(true)
	newValues := map[string]interface{}{
		"id":            sql.NullInt64{Int64: 1, Valid: true},
		"name":          &sql.NullString{String: name + "rename", Valid: true},
		"birthday":      time.Now(),
		"registered_at": mytime(time.Now()),
		"deleted_at":    &newDeletedAt,
		"active":        mybool(false),
		"admin":         &newIsAdmin,
	}

	for k, v := range newValues {
		if err := userSchema.FieldsByDBName[k].Set(reflectValue, v); err != nil {
			t.Errorf("no error should happen when assign value to field %v, but got %v", k, err)
		}
	}
	checkField(t, userSchema, reflectValue, newValues)

	newValues2 := map[string]interface{}{
		"id":            5,
		"name":          name + "rename2",
		"birthday":      time.Now(),
		"registered_at": time.Now(),
		"deleted_at":    time.Now(),
		"active":        true,
		"admin":         false,
	}

	for k, v := range newValues2 {
		if err := userSchema.FieldsByDBName[k].Set(reflectValue, v); err != nil {
			t.Errorf("no error should happen when assign value to field %v, but got %v", k, err)
		}
	}
	checkField(t, userSchema, reflectValue, newValues2)
}

type UserWithPermissionControl struct {
	ID    uint
	Name  string `dao:"-"`
	Name2 string `dao:"->"`
	Name3 string `dao:"<-"`
	Name4 string `dao:"<-:create"`
	Name5 string `dao:"<-:update"`
	Name6 string `dao:"<-:create,update"`
	Name7 string `dao:"->:false;<-:create,update"`
}

func TestParseFieldWithPermission(t *testing.T) {
	user, err := schema.Parse(&UserWithPermissionControl{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("Failed to parse user with permission, got error %v", err)
	}

	fields := []schema.Field{
		{Name: "ID", DBName: "id", BindNames: []string{"ID"}, DataType: schema.Uint, PrimaryKey: true, Size: 64, Creatable: true, Updatable: true, Readable: true, HasDefaultValue: true, AutoIncrement: true},
		{Name: "Name", DBName: "", BindNames: []string{"Name"}, DataType: "", Tag: `dao:"-"`, Creatable: false, Updatable: false, Readable: false},
		{Name: "Name2", DBName: "name2", BindNames: []string{"Name2"}, DataType: schema.String, Tag: `dao:"->"`, Creatable: false, Updatable: false, Readable: true},
		{Name: "Name3", DBName: "name3", BindNames: []string{"Name3"}, DataType: schema.String, Tag: `dao:"<-"`, Creatable: true, Updatable: true, Readable: true},
		{Name: "Name4", DBName: "name4", BindNames: []string{"Name4"}, DataType: schema.String, Tag: `dao:"<-:create"`, Creatable: true, Updatable: false, Readable: true},
		{Name: "Name5", DBName: "name5", BindNames: []string{"Name5"}, DataType: schema.String, Tag: `dao:"<-:update"`, Creatable: false, Updatable: true, Readable: true},
		{Name: "Name6", DBName: "name6", BindNames: []string{"Name6"}, DataType: schema.String, Tag: `dao:"<-:create,update"`, Creatable: true, Updatable: true, Readable: true},
		{Name: "Name7", DBName: "name7", BindNames: []string{"Name7"}, DataType: schema.String, Tag: `dao:"->:false;<-:create,update"`, Creatable: true, Updatable: true, Readable: false},
	}

	for _, f := range fields {
		checkSchemaField(t, user, &f, func(f *schema.Field) {})
	}
}