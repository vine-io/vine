// MIT License
//
// Copyright (c) 2021 Lack
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

package build

import (
	"fmt"

	"github.com/lack-io/cli"
	"github.com/lack-io/vine/cmd/vine/app/cli/util/tool"
)

func runSRV(ctx *cli.Context) {
	cfg, err := tool.New("vine.toml")
	if err != nil {
		fmt.Printf("invalid vine project: %v\n", err)
		return
	}

	atype := ctx.String("type")
	name := ctx.Args().First()

	var pb *tool.Proto
	for _, p := range cfg.Proto {
		if p.Name == name && p.Type == atype {
			pb = &p
			break
		}
	}

	if pb == nil {
		fmt.Printf("file %s.proto not found\n", name)
		return
	}
}

func cmdSRV() *cli.Command {
	return &cli.Command{
		Name:  "service",
		Usage: "build vine project",
		Flags: []cli.Flag{
			//&cli.StringFlag{
			//	Name:  "type",
			//	Usage: "the type of protobuf file eg api, service.",
			//	Value: "api",
			//},
		},
		Action: func(c *cli.Context) error {
			runSRV(c)
			return nil
		},
	}
}
