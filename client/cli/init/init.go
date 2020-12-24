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

// Package init provides the vine init command for initialising plugins and imports
package init

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/cmd"
)

var (
	// The import path we use for imports
	Import = "github.com/lack-io/vine/profile"
	// Vesion of vine
	Version = "v1"
)

func Run(ctx *cli.Context) error {
	var imports []string

	for _, val := range ctx.StringSlice("profile") {
		for _, profile := range strings.Split(val, ",") {
			p := strings.TrimSpace(profile)
			if len(p) == 0 {
				continue
			}
			path := path.Join(Import, p, Version)
			imports = append(imports, fmt.Sprintf("\t_ \"%s\"\n", path))
		}
	}

	if len(ctx.String("package")) > 0 {
		imports = append(imports, fmt.Sprintf("\t_ \"%s\"\n", ctx.String("package")))
	}

	if len(imports) == 0 {
		return nil
	}

	f := os.Stdout

	if v := ctx.String("output"); v != "stdout" {
		var err error
		f, err = os.Create(v)
		if err != nil {
			return err
		}
	}

	fmt.Fprint(f, "package main\n\n")
	fmt.Fprint(f, "import (\n")

	// write the imports
	for _, i := range imports {
		fmt.Fprint(f, i)
	}

	fmt.Fprint(f, ")\n")
	return nil
}

func init() {
	cmd.Register(&cli.Command{
		Name:        "init",
		Usage:       "Generate a profile for vine plugins",
		Description: `'vine init' generates a profile.go file defining plugins and imports`,
		Action:      Run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "package",
				Usage: "The package to load, e.g.",
			},
			&cli.StringSliceFlag{
				Name:  "profile",
				Usage: "A comma separated list of imports to load",
				Value: cli.NewStringSlice(),
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "Where to output the file, by default stdout",
				Value: "stdout",
			},
		},
	})
}
