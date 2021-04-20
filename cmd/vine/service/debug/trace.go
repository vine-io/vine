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

package debug

import (
	"fmt"
	"time"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine"
	"github.com/lack-io/vine/lib/debug"
	log "github.com/lack-io/vine/lib/logger"
)

const (
	// logUsage message for logs command
	traceUsage = "Required usage: vine trace example"
)

func getTrace(ctx *cli.Context, svcOpts ...vine.Option) {
	log.Trace("debug")

	// TODO look for trace id

	if ctx.Args().Len() == 0 {
		fmt.Println("Require service name")
		return
	}

	name := ctx.Args().Get(0)

	// must specify service name
	if len(name) == 0 {
		fmt.Println(traceUsage)
		return
	}

	// initialise a new service log
	// TODO: allow "--source" e.g. kubernetes
	client := debug.NewClient(name)

	spans, err := client.Trace()
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(spans) == 0 {
		return
	}

	fmt.Println("Id\tName\tTime\tDuration\tStatus")

	for _, span := range spans {
		fmt.Printf("%s\t%s\t%s\t%v\t%s\n",
			span.Trace,
			span.Name,
			time.Unix(0, int64(span.Started)).String(),
			time.Duration(span.Duration),
			"",
		)
	}
}
