// Copyright 2020 lack
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

// Package main define the proto-gen-gogo binary we use to generate out proto go file,
package main

import (
	"github.com/gogo/protobuf/vanity/command"

	_ "github.com/gogo/protobuf/gogoproto"
	_ "github.com/gogo/protobuf/proto"
)

func main() {
	command.Write(command.Generate(command.Read()))
}
