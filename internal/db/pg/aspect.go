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

package pg

import (
	"context"
	"time"

	"github.com/lack-io/vine/internal/db"
	"github.com/lack-io/vine/internal/db/watch"
	metav1 "github.com/lack-io/vine/internal/meta/v1"
	"github.com/lack-io/vine/internal/runtime"
	utilname "github.com/lack-io/vine/util/name"
)

// APIObjectAspect implements interface db.Aspect
type APIObjectAspect struct{}

// BeforeRequest implements db.Aspect
func (a *APIObjectAspect) BeforeRequest(ctx context.Context, obj runtime.Object) {
	// TODO: do nothing
}

// PrepareObjectForDB implements db.Aspect
func (a *APIObjectAspect) PrepareObjectForDB(_ context.Context, obj runtime.Object) error {
	accessor, err := metav1.Accessor(obj)
	if err != nil {
		return err
	}
	if accessor.GetCreationTimestamp() == 0 {
		accessor.SetCreationTimestamp(time.Now().Unix())
	}
	if accessor.GetName() == "" {
		accessor.SetName(utilname.New().Lower(6).String())
	}

	return nil
}

// UpdateObject implements db.Aspect
func (a *APIObjectAspect) UpdateObject(_ context.Context, obj runtime.Object, meta *db.ResponseMeta) error {
	//accessor, err := metav1.ListAccessor()
	return nil
}

// UpdateList implements db.Aspect
func (a *APIObjectAspect) UpdateList(_ context.Context, obj runtime.Object, meta *db.ResponseMeta) error {
	listAccessor, err := metav1.ListAccessor(obj)
	if err != nil {
		return err
	}
	if meta != nil {
		if meta.Count != 0 {
			listAccessor.SetCount(meta.Count)
		}
	}

	return nil
}

// GetObjectFromDB implements db.Aspect
func (a *APIObjectAspect) GetObjectFromDB(ctx context.Context, key string, obj runtime.Object, eventType watch.EventType) error {
	// TODO: do nothing...
	return nil
}

// AfterRequest implements db.Aspect
func (a *APIObjectAspect) AfterRequest(ctx context.Context, obj runtime.Object) {
	// TODO: do nothing...
}
