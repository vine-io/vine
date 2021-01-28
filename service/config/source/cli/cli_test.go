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

package cli

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine"
	"github.com/lack-io/vine/service/config"
	"github.com/lack-io/vine/service/config/cmd"
	"github.com/lack-io/vine/service/config/source"
)

func TestCliSourceDefault(t *testing.T) {
	const expVal string = "flagvalue"

	service := vine.NewService(
		vine.Flags(
			// to be able to run inside go test
			&cli.StringFlag{
				Name: "test.timeout",
			},
			&cli.BoolFlag{
				Name: "test.v",
			},
			&cli.StringFlag{
				Name: "test.run",
			},
			&cli.StringFlag{
				Name: "test.testlogfile",
			},
			&cli.StringFlag{
				Name:    "flag",
				Usage:   "It changes something",
				EnvVars: []string{"flag"},
				Value:   expVal,
			},
		),
	)
	var cliSrc source.Source
	service.Init(
		// Loads CLI configuration
		vine.Action(func(c *cli.Context) error {
			cliSrc = NewSource(
				Context(c),
			)
			return nil
		}),
	)

	config.Load(cliSrc)
	if fval := config.Get("flag").String("default"); fval != expVal {
		t.Fatalf("default flag value not loaded %v != %v", fval, expVal)
	}
}

func test(t *testing.T, withContext bool) {
	var src source.Source

	// setup app
	app := cmd.App()
	app.Name = "testapp"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "db-host",
			EnvVars: []string{"db-host"},
			Value:   "myval",
		},
	}

	// with context
	if withContext {
		// set action
		app.Action = func(c *cli.Context) error {
			src = WithContext(c)
			return nil
		}

		// run app
		app.Run([]string{"run", "-db-host", "localhost"})
		// no context
	} else {
		// set args
		os.Args = []string{"run", "-db-host", "localhost"}
		src = NewSource()
	}

	// test config
	c, err := src.Read()
	if err != nil {
		t.Error(err)
	}

	var actual map[string]interface{}
	if err := json.Unmarshal(c.Data, &actual); err != nil {
		t.Error(err)
	}

	actualDB := actual["db"].(map[string]interface{})
	if actualDB["host"] != "localhost" {
		t.Errorf("expected localhost, got %v", actualDB["name"])
	}

}

func TestCliSource(t *testing.T) {
	// without context
	test(t, false)
}

func TestCliSourceWithContext(t *testing.T) {
	// with context
	test(t, true)
}
