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

package cli

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/lack-io/cli"

	"github.com/vine-io/vine"
	"github.com/vine-io/vine/lib/config"
	"github.com/vine-io/vine/lib/cmd"
	"github.com/vine-io/vine/lib/config/source"
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
