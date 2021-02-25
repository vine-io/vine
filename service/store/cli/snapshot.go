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
