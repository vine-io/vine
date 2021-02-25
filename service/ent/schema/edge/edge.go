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

// A Descriptor for edge configuration.
type Descriptor struct {
	Tag         string              // struct tag.
	Type        string              // edge type.
	Name        string              // edge name.
	RefName     string              // ref name; inverse only.
	Ref         *Descriptor         // edge reference; to/from of the same type.
	Unique      bool                // unique edge.
	Inverse     bool                // inverse edge.
	Required    bool                // required on creation.
	StorageKey  *StorageKey         // optional storage-key configuration.
	Annotations []schema.Annotation // edge annotation.
}

// StorageKey holds the configuration for edge storage-key
type StorageKey struct {
	Table   string   // Table or label.
	Columns []string // Foreign-key columns.
}

// StorageOption allows for setting the storage configuration using functional options.
type StorageOption func(*StorageKey)

// Table sets the table name option for M2M edges.
func Table(name string) StorageOption {
	return func(key *StorageKey) {
		key.Table = name
	}
}

// Column sets the foreign-key column name option for O2O, O2M, and M2O edges.
// Note that, for M2M edges (2 column), use the edge.Columns option.
func Column(name string) StorageOption {
	return func(key *StorageKey) {
		key.Columns = []string{name}
	}
}

// Columns sets the foreign-key column names option for M2M edges.
// The 1st column defines the name of the "To" edge, and the 2nd defines
// the name of the "From" edge (inverse edge).
// Note that, for O2O, O2M, M2O edges, use the edge.Column option.
func Columns(to, from string) StorageOption {
	return func(key *StorageKey) {
		key.Columns = []string{to, from}
	}
}
