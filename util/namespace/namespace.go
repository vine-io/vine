// Copyright 2020 lack
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

package namespace

import (
	"context"

	"github.com/lack-io/vine/util/context/metadata"
)

const (
	DefaultNamespace = "go.vine"
	// NamespaceKey is used to set/get the namespace from the context
	NamespaceKey = "Vine-Namespace"
)

// FromContext gets the namespace from the context
func FromContext(ctx context.Context) string {
	// get the namespace which is set at ingress by vine web / api / proxy etc. The go-vine auth
	// wrapper will ensure the account making the request has the necessary issuer.
	ns, _ := metadata.Get(ctx, NamespaceKey)
	return ns
}

// ContextWithNamespace sets the namespace in the context
func ContextWithNamespace(ctx context.Context, ns string) context.Context {
	return metadata.Set(ctx, NamespaceKey, ns)
}
