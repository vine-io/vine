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

package db

import (
	"context"

	"github.com/lack-io/vine/internal/db/selection"
	"github.com/lack-io/vine/internal/db/watch"
	"github.com/lack-io/vine/internal/runtime"
)

// Aspect adopts the AOP design, like the JAVA spring. It abstracts settings and retrieving metadata fields from
// database response into the object or list. And you have extra choice to change object value, not to rewrite
// original code.
type Aspect interface {
	// BeforeRequest returns runtime.Object in the beginning.
	BeforeRequest(ctx context.Context, obj runtime.Object)

	// PrepareObjectForDB should set metadata fields to the empty value. Should
	// return an error if the specified object cannot be updated.
	PrepareObjectForDB(ctx context.Context, obj runtime.Object) error

	// UpdateObject sets database metadata into API object. Returns an error if the object
	// cannot be updated correctly. May return nil if the requested object does not need
	// metadata database.
	UpdateObject(ctx context.Context, obj runtime.Object, meta *ResponseMeta) error

	// UpdateList sets the metadata into API list object. Returns an error if the object
	// cannot be updated correctly. May return nil if the requested object does not need
	// metadata from database.
	UpdateList(ctx context.Context, obj runtime.Object, meta *ResponseMeta) error

	// GetObjectFromDB returning currently runtime.Object when client change database.
	GetObjectFromDB(ctx context.Context, key string, obj runtime.Object, eventType watch.EventType) error

	// AfterRequest returns runtime.Object at the end.
	AfterRequest(ctx context.Context, obj runtime.Object)
}

type ResponseMeta struct {
	Count  int64
}

// DB offers a common interface for object marshaling/unmarshalling operations and
// hides all the database-related operations behind it.
type DB interface {
	// Returns Aspect associated with this interface.
	Aspect() Aspect

	// Get unmarshalls json found at key into obj.
	Get(ctx context.Context, key string, pre selection.Predicate, obj runtime.Object, ignoreNotFound bool) error

	// List unmarshalls json found at key into listObj.
	List(ctx context.Context, key string, pre selection.Predicate, listObj runtime.Object) error

	// Create adds a new object at key unless it already exists. If no error is returned and
	// out is not nil, out will be set to the read value from database.
	Create(ctx context.Context, key string, obj, out runtime.Object) error

	// Update modifies the specified key and returns new value from database.
	Update(ctx context.Context, key string, pre selection.Predicate, out runtime.Object, ignoreNotFound bool) error

	// Delete removes the specified key and returns the value that existed at that spot
	// If key didn't exist, it will return NotFound db error.
	Delete(ctx context.Context, key string, pre selection.Predicate, out runtime.Object) error

	// Watch begins watching the specified key. Events are decode into objects.
	// and any items selected key are sent down to returned watch.Watch.
	Watch(ctx context.Context, key string, pre selection.Predicate, into runtime.Object) (watch.Watch, error)
}
