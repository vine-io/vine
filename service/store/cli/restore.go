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

package cli

import (
	"net/url"

	"github.com/lack-io/cli"
	"github.com/pkg/errors"

	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/store/cli/snapshot"
)

// Restore is the entrypoint for vine store restore
func Restore(ctx *cli.Context) error {
	s, err := makeStore(ctx)
	if err != nil {
		return errors.Wrap(err, "couldn't construct a store")
	}
	var rs snapshot.Restore
	source := ctx.String("source")

	if len(source) == 0 {
		return errors.New("source flag must be set")
	}
	u, err := url.Parse(source)
	if err != nil {
		return errors.Wrap(err, "source is invalid")
	}
	switch u.Scheme {
	case "file":
		rs = snapshot.NewFileRestore(snapshot.Source(source))
	default:
		return errors.Errorf("unsupported source scheme: %s", u.Scheme)
	}

	err = rs.Init()
	if err != nil {
		return errors.Wrap(err, "failed to initialise the restorer")
	}

	recordChan, err := rs.Start()
	if err != nil {
		return errors.Wrap(err, "couldn't start the restorer")
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
