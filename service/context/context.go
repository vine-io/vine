// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package context provides a context for accessing services
package context

import (
	"context"

	"github.com/lack-io/vine/internal/namespace"
	"github.com/lack-io/vine/service/context/metadata"
)

var (
	// DefaultContext is a context which can be used to access vine services
	DefaultContext = WithNamespace("vine")
)

// WithNamespace creates a new context with the given namespace
func WithNamespace(ns string) context.Context {
	return SetNamespace(context.TODO(), ns)
}

// SetNamespace sets the namespace for a context
func SetNamespace(ctx context.Context, ns string) context.Context {
	return namespace.ContextWithNamespace(ctx, ns)
}

// SetMetadata sets the metadata within the context
func SetMetadata(ctx context.Context, k, v string) context.Context {
	return metadata.Set(ctx, k, v)
}

// GetMetadata returns metadata from the context
func GetMetadata(ctx context.Context, k string) (string, bool) {
	return metadata.Get(ctx, k)
}
