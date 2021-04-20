// MIT License
//
// Copyright (c) 2020 Lack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package build builds a vine runtime package
package build

import (
	"github.com/lack-io/vine/lib/runtime/local/source"
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
