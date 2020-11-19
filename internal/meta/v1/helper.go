// Copyright 2020 The vine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/lack-io/vine/internal/runtime"
)

var (
	errExpectFieldItems = errors.New("no Items field in this object")
	//errExpectFieldStatus = errors.New("no Status field in this object")
	errExpectSliceItems = errors.New("items field must be a slice of objects")
	// errNotList is returned when an object implements the Object style interfaces but not the
	// List style interfaces
	errNotList = fmt.Errorf("object does not implement the List interfaces")
	// errNotObject is returned when an object implements the List style interfaces but not the
	// Object style interfaces
	errNotObject = errors.New("object does not implement the object interfaces")
)

// GetItemPtr returns a pointer to the list object's Items member.
// If 'list' doesn't have an Items member, it's not really a list type
// and an error will be returned.
// This function will either return a pointer to a slice, or an error, but not both.
func GetItemPtr(list runtime.Object) (interface{}, error) {
	obj, err := getItemsPtr(list)
	if err != nil {
		return nil, fmt.Errorf("%T is not a list: %v", list, err)
	}
	return obj, nil
}

// getItemsPtr returns a pointer to the list object's Items member or an error.
func getItemsPtr(list runtime.Object) (interface{}, error) {
	v, err := EnforcePtr(list)
	if err != nil {
		return nil, err
	}

	items := v.FieldByName("Items")
	if !items.IsValid() {
		return nil, errExpectFieldItems
	}
	switch items.Kind() {
	case reflect.Interface, reflect.Ptr:
		target := reflect.TypeOf(items.Interface()).Elem()
		if target.Kind() != reflect.Slice {
			return nil, errExpectSliceItems
		}
		return items.Interface(), nil
	case reflect.Slice:
		return items.Addr().Interface(), nil
	default:
		return nil, errExpectSliceItems
	}
}

// EnforcePtr ensures that object is a pointer of some sort. Returns a reflect.Value
// of the dereferenced pointer, ensuring that it is settable/addressable.
// Returns an error if this is not possible.
func EnforcePtr(obj interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		if v.Kind() == reflect.Invalid {
			return reflect.Value{}, fmt.Errorf("expected pointer, but got invalid kind")
		}
		return reflect.Value{}, fmt.Errorf("expected pointer, but got %v type", v.Type())
	}
	if v.IsNil() {
		return reflect.Value{}, fmt.Errorf("expected pointer, but got nil")
	}
	return v.Elem(), nil
}

//
func ListAccessor(obj interface{}) (ListInterface, error) {
	switch t := obj.(type) {
	case ListInterface:
		return t, nil
	default:
		return nil, errNotList
	}
}

func Accessor(obj interface{}) (Object, error) {
	switch t := obj.(type) {
	case Object:
		return t, nil
	default:
		return nil, errNotObject
	}
}
