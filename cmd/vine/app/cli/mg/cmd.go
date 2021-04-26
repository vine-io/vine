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

// Package mg new generates vine service templates
package mg

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/lack-io/cli"
	"github.com/lack-io/vine/util/helper"
)

func Commands() []*cli.Command {
	return []*cli.Command{
		{
			Name:        "new",
			Usage:       "Create vine resource template",
			Subcommands: []*cli.Command{cmdSRV(), cmdGateway(), cmdWeb(), cmdProto()},
			Action: func(c *cli.Context) error {
				if c.Args().Len() > 0 {
					command := c.Args().First()

					v, err := exec.LookPath(command)
					if err != nil {
						fmt.Println(helper.UnexpectedSubcommand(c))
						os.Exit(1)
					}

					// execute the command
					ce := exec.Command(v, c.Args().Slice()[1:]...)
					ce.Stdout = os.Stdout
					ce.Stderr = os.Stderr
					return ce.Run()
				}
				fmt.Println("No command provided to vine. Please refer to 'vine new help'")
				os.Exit(1)
				return nil
			},
		},
		{
			Name:  "init",
			Usage: "Initialize a vine project",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "namespace",
					Usage: "Namespace for the project e.g com.example",
					Value: "go.vine",
				},
				&cli.BoolFlag{
					Name:  "cluster",
					Usage: "create cluster package.",
				},
			},
			Action: func(c *cli.Context) error {
				runInit(c)
				return nil
			},
		},
	}
}
