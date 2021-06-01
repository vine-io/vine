package runtime

import (
	"reflect"
	"testing"
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
