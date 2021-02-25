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

package edge

import "github.com/lack-io/vine/service/ent/schema"

// Annotation is a builtin schema annotation for
// configuring the edges behavior in codegen.
type Annotation struct {
	// The StructTag option allows overriding the struct-tag
	// of the `Edges` field in the generated entity. For example:
	//
	// edge.Annotation{
	//		StructTag: `json: "pet_edges"
	// }
	//
	StructTag string
}

// Name describes the annotation name.
func (Annotation) Name() string {
	return "Edges"
}

// Merge implements the schema.Merge interface.
func (a Annotation) Merge(other schema.Annotation) schema.Annotation {
	var ant Annotation
	switch other := other.(type) {
	case Annotation:
		ant = other
	case *Annotation:
		if other != nil {
			ant = *other
		}
	default:
		return a
	}
	if tag := ant.StructTag; tag != "" {
		a.StructTag = tag
	}
	return a
}

var (
	_ schema.Annotation = (*Annotation)(nil)
	_ schema.Merger     = (*Annotation)(nil)
)
