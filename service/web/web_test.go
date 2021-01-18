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

package web_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/service"
	"github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/web"
)

func TestWeb(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println("Test nr", i)
		testFunc()
	}
}

func testFunc() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*250)
	defer cancel()

	s := service.NewService(
		service.Name("test"),
		service.Context(ctx),
		service.HandleSignal(false),
		service.Flags(
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
		),
	)
	w := web.NewService(
		web.VineService(s),
		web.Context(ctx),
		web.HandleSignal(false),
	)
	//s.Init()
	//w.Init()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := s.Run()
		if err != nil {
			logger.Errorf("vine run error: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		err := w.Run()
		if err != nil {
			logger.Errorf("web run error: %v", err)
		}
	}()

	wg.Wait()
}
