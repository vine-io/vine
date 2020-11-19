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
	"github.com/lack-io/vine/internal/runtime"
)

type Dao interface {
	// Prefix returns the information about database table
	Prefix() string

	Get(ctx context.Context, selectors []selection.Selector, into runtime.Object) error

	List(ctx context.Context, selectors[]selection.Selector, list runtime.Object) error

	Create(ctx context.Context, out runtime.Object) error

	Update(ctx context.Context, out runtime.Object) error

	Delete(ctx context.Context, grace bool, into runtime.Object) error
}
