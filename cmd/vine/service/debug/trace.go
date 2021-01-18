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

package debug

import (
	"fmt"
	"time"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/service"
	"github.com/lack-io/vine/service/debug"
	log "github.com/lack-io/vine/service/logger"
)

const (
	// logUsage message for logs command
	traceUsage = "Required usage: vine trace example"
)

func getTrace(ctx *cli.Context, srvOpts ...service.Option) {
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
