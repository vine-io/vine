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

package metadata

import (
	"context"
	"reflect"
	"testing"
)

func TestMetadataSet(t *testing.T) {
	ctx := Set(context.TODO(), "Key", "val")

	val, ok := Get(ctx, "Key")
	if !ok {
		t.Fatal("key Key not found")
	}
	if val != "val" {
		t.Errorf("key Key with value val != %v", val)
	}
}

func TestMetadataDelete(t *testing.T) {
	md := Metadata{
		"Foo": "bar",
		"Baz": "empty",
	}

	ctx := NewContext(context.TODO(), md)
	ctx = Delete(ctx, "Baz")

	emd, ok := FromContext(ctx)
	if !ok {
		t.Fatal("key Key not found")
	}

	_, ok = emd["Baz"]
	if ok {
		t.Fatal("key Baz not deleted")
	}

}

func TestMetadataCopy(t *testing.T) {
	md := Metadata{
		"Foo": "bar",
		"bar": "baz",
	}

	cp := Copy(md)

	for k, v := range md {
		if cv := cp[k]; cv != v {
			t.Fatalf("Got %s:%s for %s:%s", k, cv, k, v)
		}
	}
}

func TestMetadataContext(t *testing.T) {
	md := Metadata{
		"foo": "bar",
	}

	ctx := NewContext(context.TODO(), md)

	emd, ok := FromContext(ctx)
	if !ok {
		t.Errorf("Unexpected error retrieving metadata, got %t", ok)
	}

	if emd["foo"] != md["foo"] {
		t.Errorf("Expected key: %s val: %s, got key: %s val: %s", "foo", md["foo"], "foo", emd["foo"])
	}

	if i := len(emd); i != 1 {
		t.Errorf("Expected metadata length 1 got %d", i)
	}
}

func TestMergeContext(t *testing.T) {
	type args struct {
		existing  Metadata
		append    Metadata
		overwrite bool
	}
	tests := []struct {
		name string
		args args
		want Metadata
	}{
		{
			name: "matching key, overwrite false",
			args: args{
				existing:  Metadata{"Foo": "bar", "Sumo": "demo"},
				append:    Metadata{"Sumo": "demo2"},
				overwrite: false,
			},
			want: Metadata{"foo": "bar", "sumo": "demo"},
		},
		{
			name: "matching key, overwrite true",
			args: args{
				existing:  Metadata{"Foo": "bar", "Sumo": "demo"},
				append:    Metadata{"Sumo": "demo2"},
				overwrite: true,
			},
			want: Metadata{"foo": "bar", "sumo": "demo2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := FromContext(MergeContext(NewContext(context.TODO(), tt.args.existing), tt.args.append, tt.args.overwrite)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
