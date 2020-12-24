// Copyright 2020 The vine Authors
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

// Package plugins includes the plugins we want to load
package plugins

import (
	"github.com/lack-io/vine/service/config/cmd"
	// we only use CF internally for certs
	cfStore "github.com/lack-io/vine/internal/plugins/store/cloudflare"

	// import specific plugins
	fileStore "github.com/lack-io/vine/service/store/bolt"
	memStore "github.com/lack-io/vine/service/store/memory"
	pg "github.com/lack-io/vine/service/store/postgres"
)

func init() {
	// TODO: make it so we only have to import them
	cmd.DefaultStores["cloudflare"] = cfStore.NewStore
	cmd.DefaultStores["postgres"] = pg.NewStore
	cmd.DefaultStores["file"] = fileStore.NewStore
	cmd.DefaultStores["memory"] = memStore.NewStore
}
