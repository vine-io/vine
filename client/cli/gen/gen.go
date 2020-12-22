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

// Package gen provides the vine gen command which simply runs go generate
package init

import (
	"fmt"
	"os/exec"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/cmd"
)

var (
	Command = "go generate"
)

func Run(ctx *cli.Context) error {
	cmd := exec.Command("go", "generate")
	b, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Print(string(b))
	return nil
}

func init() {
	cmd.Register(&cli.Command{
		Name:        "gen",
		Usage:       "Generate a vine related dependencies e.g protobuf",
		Description: `'vine gen' will generate any vine related dependencies such as proto files`,
		Action:      Run,
		Flags:       []cli.Flag{},
	})
}
