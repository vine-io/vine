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
	"fmt"
	"strings"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/service/config/cmd"
	"github.com/lack-io/vine/service/store"
)

// makeStore is a helper function that creates a store for snapshot and restore
func makeStore(ctx *cli.Context) (store.Store, error) {
	builtinStore, err := getStore(ctx.String("store"))
	if err != nil {
		return nil, fmt.Errorf("makeStore: %w", err)
	}
	s := builtinStore(
		store.Nodes(strings.Split(ctx.String("nodes"), ",")...),
		store.Database(ctx.String("database")),
		store.Table(ctx.String("table")),
	)
	if err := s.Init(); err != nil {
		return nil, fmt.Errorf("Couldn't init %s store: %w", ctx.String("store"), err)
	}
	return s, nil
}

// makeStores is a helper function that sets up 2 stores for sync
func makeStores(ctx *cli.Context) (store.Store, store.Store, error) {
	fromBuilder, err := getStore(ctx.String("from-backend"))
	if err != nil {
		return nil, nil, fmt.Errorf("from store: %w", err)
	}
	toBuilder, err := getStore(ctx.String("to-backend"))
	if err != nil {
		return nil, nil, fmt.Errorf("to store: %w", err)
	}
	from := fromBuilder(
		store.Nodes(strings.Split(ctx.String("from-nodes"), ",")...),
		store.Database(ctx.String("from-database")),
		store.Table(ctx.String("from-table")),
	)
	if err := from.Init(); err != nil {
		return nil, nil, fmt.Errorf("from: couldn't init %s: %w", ctx.String("from-backend"), err)
	}
	to := toBuilder(
		store.Nodes(strings.Split(ctx.String("to-nodes"), ",")...),
		store.Database(ctx.String("to-database")),
		store.Table(ctx.String("to-table")),
	)
	if err := to.Init(); err != nil {
		return nil, nil, fmt.Errorf("to: couldn't init %s: %w", ctx.String("to-backend"), err)
	}
	return from, to, nil
}

func getStore(s string) (func(...store.Option) store.Store, error) {
	builtinStore, exists := cmd.DefaultStores[s]
	if !exists {
		return nil, fmt.Errorf("store %s is not an implemented store - check your plugins", s)
	}
	return builtinStore, nil
}
