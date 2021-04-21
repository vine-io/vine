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
	"fmt"
	"strings"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/lib/cmd"
	"github.com/lack-io/vine/lib/store"
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
