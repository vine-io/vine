// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"net/url"

	"github.com/lack-io/cli"
	"github.com/pkg/errors"

	snap "github.com/lack-io/vine/client/cli/store/snapshot"
	"github.com/lack-io/vine/service/logger"
)

// snapshot in the entrypoint for vine store snapshot
func snapshot(ctx *cli.Context) error {
	s, err := makeStore(ctx)
	if err != nil {
		return errors.Wrap(err, "couldn't construct a store")
	}
	log := logger.DefaultLogger
	dest := ctx.String("destination")
	var sn snap.Snapshot

	if len(dest) == 0 {
		return errors.New("destination flag must be set")
	}
	u, err := url.Parse(dest)
	if err != nil {
		return errors.Wrap(err, "destination is invalid")
	}
	switch u.Scheme {
	case "file":
		sn = snap.NewFileSnapshot(snap.Destination(dest))
	default:
		return errors.Errorf("unsupported destination scheme: %s", u.Scheme)
	}
	err = sn.Init()
	if err != nil {
		return errors.Wrap(err, "failed to initialise the snapshotter")
	}

	log.Logf(logger.InfoLevel, "Snapshotting store %s", s.String())
	recordChan, err := sn.Start()
	if err != nil {
		return errors.Wrap(err, "couldn't start the snapshotter")
	}
	keys, err := s.List()
	if err != nil {
		return errors.Wrap(err, "couldn't List() from store "+s.String())
	}
	log.Logf(logger.DebugLevel, "Snapshotting %d keys", len(keys))

	for _, key := range keys {
		r, err := s.Read(key)
		if err != nil {
			return errors.Wrapf(err, "couldn't read key %s", key)
		}
		if len(r) != 1 {
			return errors.Errorf("reading %s from %s returned 0 records", key, s.String())
		}
		recordChan <- r[0]
	}
	close(recordChan)
	sn.Wait()
	return nil
}
