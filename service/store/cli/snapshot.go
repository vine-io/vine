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

	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/store/cli/snapshot"
)

// Snapshot in the entrypoint for vine store snapshot
func Snapshot(ctx *cli.Context) error {
	s, err := makeStore(ctx)
	if err != nil {
		return fmt.Errorf("%w couldn't construct a store", err)
	}
	dest := ctx.String("destination")
	var sn snapshot.Snapshot

	if len(dest) == 0 {
		return errors.New("destination flag must be set")
	}
	u, err := url.Parse(dest)
	if err != nil {
		return fmt.Errorf("%w destination is invalid", err)
	}
	switch u.Scheme {
	case "file":
		sn = snapshot.NewFileSnapshot(snapshot.Destination(dest))
	default:
		return fmt.Errorf("unsupported destination scheme: %s", u.Scheme)
	}
	err = sn.Init()
	if err != nil {
		return fmt.Errorf("%w failed to initialise the snapshotter", err)
	}

	log.Debugf("Snapshotting store %s", s.String())
	recordChan, err := sn.Start()
	if err != nil {
		return fmt.Errorf("%w couldn't start the snapshotter", err)
	}
	keys, err := s.List()
	if err != nil {
		return fmt.Errorf("%w couldn't List() from store "+s.String(), err)
	}
	log.Debugf("Snapshotting %d keys", len(keys))

	for _, key := range keys {
		r, err := s.Read(key)
		if err != nil {
			return fmt.Errorf("%w couldn't read key %s", err, key)
		}
		if len(r) != 1 {
			return fmt.Errorf("reading %s from %s returned 0 records", key, s.String())
		}
		recordChan <- r[0]
	}
	close(recordChan)
	sn.Wait()
	return nil
}
