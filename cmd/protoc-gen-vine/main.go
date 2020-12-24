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

// protoc-gen-vine is a plugin for the Google protocol buffer compiler to generate
// Go code. Run it bu building this program and putting it in your path with
// the name
//		protoc-gen-vine
// That word 'vine' at the end becomes part of the option string set for the
// protocol compiler, so once the protocol compiler (protoc) is installed
// you can run
//		protoc --vine_out=out_directory --go_out=output_directory input_directory/file.proto
// to generate vine code for the protocol defined by file.proto.
// With input that, the output will be written to
//		output_directory/file.vine.go
//
// The generated code is documented in the package comment for
// the library.
//
// See the README and documentation is protocol buffers to learn more:
//		https://developers.google.com/protocol-buffers/
package main

import (
	"io/ioutil"
	"os"

	"github.com/gogo/protobuf/proto"

	"github.com/lack-io/vine/cmd/protoc-gen-vine/generator"
	_ "github.com/lack-io/vine/cmd/protoc-gen-vine/plugin/vine"
)

func main() {
	// Begin by allocating a generator. The request and response structures are stored there
	// so we can do error handling easily - the response structure contains the field to
	// report failure.
	g := generator.New()

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		g.Error(err, "reading input")
	}

	if err := proto.Unmarshal(data, g.Request); err != nil {
		g.Error(err, "parsing input proto")
	}

	if len(g.Request.FileToGenerate) == 0 {
		g.Fail("no files to generate")
	}

	g.CommandLineParameters(g.Request.GetParameter())

	// Create a wrapped version of the Descriptors EnumDescriptors that
	// point to the file that defines them.
	g.WrapTypes()

	g.SetPackageNames()
	g.BuildTypeNameMap()

	g.GenerateAllFiles()

	// Send back the results.
	data, err = proto.Marshal(g.Response)
	if err != nil {
		g.Error(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		g.Error(err, "failed to write output proto")
	}
}
