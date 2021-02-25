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

package schema

// Annotation is used to attach arbitrary metadata to the schema objects in codegen.
// The object must be serializable to JSON raw value (e.g struct, map or slice).
//
// Template extensions can retrieve this metadata and use it inside their templates.
type Annotation interface {
	// Name defines the name of the annotation to be retrieved by the codegen.
	Name() string
}

// Merger wraps the single Merge function allows custom annotation to provide
// an implementation for merging 2 or more annotations from the same type.
//
// A common use case is where the same Annotation type is defined both in
// mixin.Schema and ent.Schema
type Merger interface {
	Merge(Annotation) Annotation
}