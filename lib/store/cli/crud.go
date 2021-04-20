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
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"
	"unicode/utf8"

	"github.com/dustin/go-humanize"
	"github.com/lack-io/cli"

	"github.com/lack-io/vine/lib/config/cmd"
	"github.com/lack-io/vine/lib/store"
)

// Read gets something from the store
func Read(ctx *cli.Context) error {
	if err := initStore(ctx); err != nil {
		return err
	}
	if ctx.Args().Len() < 1 {
		return errors.New("Key arg is required")
	}
	opts := []store.ReadOption{}
	if ctx.Bool("prefix") {
		opts = append(opts, store.ReadPrefix())
	}

	store := *cmd.DefaultOptions().Store
	records, err := store.Read(ctx.Args().First(), opts...)
	if err != nil {
		if err.Error() == "not found" {
			return err
		}
		return fmt.Errorf("%w Couldn't read %s from store", err, ctx.Args().First())
	}
	switch ctx.String("output") {
	case "json":
		jsonRecords, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			return fmt.Errorf("%w failed marshalling JSON", err)
		}
		fmt.Printf("%s\n", string(jsonRecords))
	default:
		if ctx.Bool("verbose") {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			fmt.Fprintf(w, "%v \t %v \t %v\n", "KEY", "VALUE", "EXPIRY")
			for _, r := range records {
				var key, value, expiry string
				key = r.Key
				if isPrintable(r.Value) {
					value = string(r.Value)
					if len(value) > 50 {
						runes := []rune(value)
						value = string(runes[:50]) + "..."
					}
				} else {
					value = fmt.Sprintf("%#x", r.Value[:20])
				}
				if r.Expiry == 0 {
					expiry = "None"
				} else {
					expiry = humanize.Time(time.Now().Add(r.Expiry))
				}
				fmt.Fprintf(w, "%v \t %v \t %v\n", key, value, expiry)
			}
			w.Flush()
			return nil
		}
		for _, r := range records {
			fmt.Println(string(r.Value))
		}
	}
	return nil
}

// Write puts something in the store.
func Write(ctx *cli.Context) error {
	if err := initStore(ctx); err != nil {
		return err
	}
	if ctx.Args().Len() < 2 {
		return errors.New("Key and Value args are required")
	}
	record := &store.Record{
		Key:   ctx.Args().First(),
		Value: []byte(strings.Join(ctx.Args().Tail(), " ")),
	}
	if len(ctx.String("expiry")) > 0 {
		d, err := time.ParseDuration(ctx.String("expiry"))
		if err != nil {
			return fmt.Errorf("expiry flag is invalid: %w", err)
		}
		record.Expiry = d
	}

	store := *cmd.DefaultOptions().Store
	if err := store.Write(record); err != nil {
		return fmt.Errorf("couldn't write: %w", err)
	}
	return nil
}

// List retrieves keys
func List(ctx *cli.Context) error {
	if err := initStore(ctx); err != nil {
		return err
	}
	var opts []store.ListOption
	if ctx.Bool("prefix") {
		opts = append(opts, store.ListPrefix(ctx.Args().First()))
	}
	if ctx.Uint("limit") != 0 {
		opts = append(opts, store.ListLimit(ctx.Uint("limit")))
	}
	if ctx.Uint("offset") != 0 {
		opts = append(opts, store.ListLimit(ctx.Uint("offset")))
	}
	store := *cmd.DefaultOptions().Store
	keys, err := store.List(opts...)
	if err != nil {
		return fmt.Errorf("couldn't list: %w", err)
	}
	switch ctx.String("output") {
	case "json":
		jsonRecords, err := json.MarshalIndent(keys, "", "  ")
		if err != nil {
			return fmt.Errorf("failed marshalling JSON: %w", err)
		}
		fmt.Printf("%s\n", string(jsonRecords))
	default:
		for _, key := range keys {
			fmt.Println(key)
		}
	}
	return nil
}

// Delete deletes keys
func Delete(ctx *cli.Context) error {
	if err := initStore(ctx); err != nil {
		return err
	}
	if len(ctx.Args().Slice()) == 0 {
		return errors.New("key is required")
	}
	store := *cmd.DefaultOptions().Store
	if err := store.Delete(ctx.Args().First()); err != nil {
		return fmt.Errorf("couldn't delete key %s: %w", ctx.Args().First(), err)
	}
	return nil
}

func initStore(ctx *cli.Context) error {
	opts := []store.Option{}
	if len(ctx.String("database")) > 0 {
		opts = append(opts, store.Database(ctx.String("database")))
	}
	if len(ctx.String("table")) > 0 {
		opts = append(opts, store.Table(ctx.String("table")))
	}
	if len(opts) > 0 {
		store := *cmd.DefaultOptions().Store
		if err := store.Init(opts...); err != nil {
			return fmt.Errorf("couldn't reinitialise store with options: %w", err)
		}
	}
	return nil
}

func isPrintable(b []byte) bool {
	s := string(b)
	for _, r := range []rune(s) {
		if r == utf8.RuneError {
			return false
		}
	}
	return true
}
