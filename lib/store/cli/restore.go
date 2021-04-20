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
	"errors"
	"fmt"
	"net/url"

	"github.com/lack-io/cli"

	log "github.com/lack-io/vine/lib/logger"
	"github.com/lack-io/vine/lib/store/cli/snapshot"
)

// Restore is the entrypoint for vine store restore
func Restore(ctx *cli.Context) error {
	s, err := makeStore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't construct a store: %w", err)
	}
	var rs snapshot.Restore
	source := ctx.String("source")

	if len(source) == 0 {
		return errors.New("source flag must be set")
	}
	u, err := url.Parse(source)
	if err != nil {
		return fmt.Errorf("source is invalid: %w", err)
	}
	switch u.Scheme {
	case "file":
		rs = snapshot.NewFileRestore(snapshot.Source(source))
	default:
		return fmt.Errorf("unsupported source scheme: %s", u.Scheme)
	}

	err = rs.Init()
	if err != nil {
		return fmt.Errorf("failed to initialise the restorer: %w", err)
	}

	recordChan, err := rs.Start()
	if err != nil {
		return fmt.Errorf("couldn't start the restorer: %w", err)
	}
	counter := uint64(0)
	for r := range recordChan {
		err := s.Write(r)
		if err != nil {
			log.Debugf("couldn't write key %s to store %s", r.Key, s.String())
		} else {
			counter++
		}
	}
	log.Debugf("Restored %d records", counter)
	return nil
}
