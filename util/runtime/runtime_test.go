package runtime

import (
	"context"
	"reflect"
	"testing"

	"github.com/lack-io/vine/lib/dao"
	"github.com/lack-io/vine/lib/dao/clause"
)

func TestFromGVK(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want *GroupVersionKind
	}{
		{"FrmGVK-test1", args{"v1.App"}, &GroupVersionKind{Version: "v1", Kind: "App"}},
		{"FrmGVK-test2", args{"auth/v1.User"}, &GroupVersionKind{Group: "auth", Version: "v1", Kind: "User"}},
		{"FrmGVK-test3", args{"core/A"}, &GroupVersionKind{Group: "core", Version: "v1", Kind: "A"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromGVK(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromGVK() = %v, want %v", got, tt.want)
			}
		})
	}
}

type TestData struct {

}

func (t *TestData) GVK() *GroupVersionKind {
	return &GroupVersionKind{Group: "", Version: "v1", Kind: "TestData"}
}

func (t *TestData) DeepCopy() Object {
	return &TestData{}
}

type TestDataSchema struct {

}

func (t TestDataSchema) FindPage(ctx context.Context, page, size int) ([]Object, int64, error) {
	panic("implement me")
}

func (t TestDataSchema) FindAll(ctx context.Context) ([]Object, error) {
	panic("implement me")
}

func (t TestDataSchema) FindPureAll(ctx context.Context) ([]Object, error) {
	panic("implement me")
}

func (t TestDataSchema) Count(ctx context.Context) (total int64, err error) {
	panic("implement me")
}

func (t TestDataSchema) FindOne(ctx context.Context) (Object, error) {
	panic("implement me")
}

func (t TestDataSchema) Cond(exprs ...clause.Expression) Schema {
	panic("implement me")
}

func (t TestDataSchema) Create(ctx context.Context) (Object, error) {
	panic("implement me")
}

func (t TestDataSchema) BatchUpdates(ctx context.Context) error {
	panic("implement me")
}

func (t TestDataSchema) Updates(ctx context.Context) (Object, error) {
	panic("implement me")
}

func (t TestDataSchema) BatchDelete(ctx context.Context, soft bool) error {
	panic("implement me")
}

func (t TestDataSchema) Delete(ctx context.Context, soft bool) error {
	panic("implement me")
}

func (t TestDataSchema) Tx(ctx context.Context) *dao.DB {
	panic("implement me")
}

var _ Schema = (*TestDataSchema)(nil)

func TestSchemaSet_RegistrySchema(t *testing.T) {
	set := NewSchemaSet()

	set.RegistrySchema(new(TestData).GVK(), func(object Object) Schema {
		return TestDataSchema{}
	})

	t.Log(set)

	ss, ok := set.NewSchema(new(TestData))
	if !ok {
		t.Fatal("unknown schema")
	}
	t.Log(ss)
}