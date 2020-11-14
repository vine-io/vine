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

package deepcopy_gen

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/lack-io/vine/gogenerator/args"
)

// NewDefaults returns arguments for the generator.
func NewDefaults() (*args.GeneratorArgs, *CustomArgs) {
	genericArgs := args.Default().WithoutDefaultFlagParsing()
	customArgs := &CustomArgs{}
	genericArgs.CustomArgs = (*CustomArgs)(customArgs) // convert to upstream type to make type-casts work there
	genericArgs.OutputFileBaseName = "deepcopy_generated"
	return genericArgs, customArgs
}

// AddFlags add the generator flags to the flag set.
func (ca *CustomArgs) AddFlags(fs *pflag.FlagSet) {
	pflag.CommandLine.StringSliceVar(&ca.BoundingDirs, "bounding-dirs", ca.BoundingDirs,
		"Comma-separated list of import path which bound the types for which deep-copies will be generated.")
}

// Validate checks the given arguments.
func Validate(genericArgs *args.GeneratorArgs) error {
	_ = genericArgs.CustomArgs.(*CustomArgs)

	if len(genericArgs.OutputFileBaseName) == 0 {
		return fmt.Errorf("output file base name cannot be empty")
	}

	return nil
}