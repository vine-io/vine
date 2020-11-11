// Copyright 2020 The kubernetes Authors
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

package namer

import "github.com/lack-io/vine/gogenerator/types"

// ImportTracker may be passed to a namer.RawNamer, to tracker the imports needed
// for the type it names.
//
// TODO: pay attention to the package name (instead of renaming every package).
type DefaultImportTracker struct {
	pathToName map[string]string
	// forbidden names are in here. (e.g. "go" is a directory in which
	// there is code, but "go" is not a legal name for a package, so we put
	// it here to prevent us from naming any package "go")
	nameToPath map[string]string
	local types.Name

	
}