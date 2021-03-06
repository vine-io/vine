package schema_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/vine-io/vine/lib/dao"
	"github.com/vine-io/vine/lib/dao/schema"
)

type UserWithCallback struct {
}

func (UserWithCallback) BeforeSave(*dao.DB) error {
	return nil
}

func (UserWithCallback) AfterCreate(*dao.DB) error {
	return nil
}

func TestCallback(t *testing.T) {
	user, err := schema.Parse(&UserWithCallback{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("failed to parse user with callback, got error %v", err)
	}

	for _, str := range []string{"BeforeSave", "AfterCreate"} {
		if !reflect.Indirect(reflect.ValueOf(user)).FieldByName(str).Interface().(bool) {
			t.Errorf("%v should be true", str)
		}
	}

	for _, str := range []string{"BeforeCreate", "BeforeUpdate", "AfterUpdate", "AfterSave", "BeforeDelete", "AfterDelete", "AfterFind"} {
		if reflect.Indirect(reflect.ValueOf(user)).FieldByName(str).Interface().(bool) {
			t.Errorf("%v should be false", str)
		}
	}
}
