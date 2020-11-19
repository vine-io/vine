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

package selection


type PK struct {
	Field string
	Value interface{}
}

// Predicate is used to represent the way to change the interaction with the database
type Predicate struct {
	// primary key information
	PK *PK

	// Fields returns specified field.
	Fields []string

	// Selectors selects specified data and contains query conditions.
	Selectors []Selector

	// Clauses contains database clauses when client request database.
	Clauses []Clause

	// delete object gracefully
	DeletionGrace bool
}
