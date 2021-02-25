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

package field

import "github.com/lack-io/vine/service/ent/schema"

// Annotation is a builtin schema annotation for
// configuration the schema fields in codegen.
type Annotation struct {
	// The StructTag option allows overriding the struct-tag
	// of the fields in the generated entity. For example:
	//
	//	field.Annotation{
	//		StructTag: map[string]string{
	//			"id": `json:"id,omitempty" yaml:"-"`
	//		}
	//	}
	//
	StructTag map[string]string
}

// Name describes the annotation name.
func (Annotation) Name() string {
	return "Fields"
}

// Merge implements the schema.Merger interface.
func (a Annotation) Merge(other schema.Annotation) schema.Annotation {
	var ant Annotation
	switch other := other.(type) {
	case Annotation:
		ant = other
	case *Annotation:
		ant = *other
	default:
		return a
	}
	for k, v := range ant.StructTag {
		if a.StructTag == nil {
			a.StructTag = make(map[string]string)
		}
		a.StructTag[k] = v
	}
	return a
}

var (
	_ schema.Annotation = (*Annotation)(nil)
	_ schema.Merger     = (*Annotation)(nil)
)
