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

// Package build builds a vine runtime package
package build

import (
	"github.com/lack-io/vine/runtime/local/source"
)

// Builder builds binaries
type Builder interface {
	// Build builds a package
	Build(*Source) (*Package, error)
	// Clean deletes the package
	Clean(*Package) error
}

// Source is the source of a build
type Source struct {
	// Language is the language of code
	Language string
	// Location of the source
	Repository *source.Repository
}

// Package is vine service package
type Package struct {
	// Name of the binary
	Name string
	// Location of the binary
	Path string
	// Type of binary
	Type string
	// Source of the binary
	Source *Source
}
