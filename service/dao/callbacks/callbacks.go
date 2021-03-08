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

package callbacks

import (
	"github.com/lack-io/vine/service/dao"
)

type Options struct {
	LastInsertIDReversed bool
	WithReturning        bool
}

func RegisterDefaultCallbacks(db *dao.DB, config *Options) {
	enableTransaction := func(db *dao.DB) bool {
		return !db.SkipDefaultTransaction
	}

	createCallback := db.Callback().Create()
	createCallback.Match(enableTransaction).Register("dao:begin_transaction", BeginTransaction)
	createCallback.Register("dao:before_create", BeforeCreate)
	createCallback.Register("dao:save_before_associations", SaveBeforeAssociations(true))
	createCallback.Register("dao:create", Create(config))
	createCallback.Register("dao:save_after_associations", SaveAfterAssociations(true))
	createCallback.Register("dao:after_create", AfterCreate)
	createCallback.Match(enableTransaction).Register("dao:commit_or_rollback_transaction", CommitOrRollbackTransaction)

	queryCallback := db.Callback().Query()
	queryCallback.Register("dao:query", Query)
	queryCallback.Register("dao:preload", Preload)
	queryCallback.Register("dao:after_query", AfterQuery)

	deleteCallback := db.Callback().Delete()
	deleteCallback.Match(enableTransaction).Register("dao:begin_transaction", BeginTransaction)
	deleteCallback.Register("dao:before_delete", BeforeDelete)
	deleteCallback.Register("dao:delete_before_associations", DeleteBeforeAssociations)
	deleteCallback.Register("dao:delete", Delete)
	deleteCallback.Register("dao:after_delete", AfterDelete)
	deleteCallback.Match(enableTransaction).Register("dao:commit_or_rollback_transaction", CommitOrRollbackTransaction)

	updateCallback := db.Callback().Update()
	updateCallback.Match(enableTransaction).Register("dao:begin_transaction", BeginTransaction)
	updateCallback.Register("dao:setup_reflect_value", SetupUpdateReflectValue)
	updateCallback.Register("dao:before_update", BeforeUpdate)
	updateCallback.Register("dao:save_before_associations", SaveBeforeAssociations(false))
	updateCallback.Register("dao:update", Update)
	updateCallback.Register("dao:save_after_associations", SaveAfterAssociations(false))
	updateCallback.Register("dao:after_update", AfterUpdate)
	updateCallback.Match(enableTransaction).Register("dao:commit_or_rollback_transaction", CommitOrRollbackTransaction)

	db.Callback().Row().Register("dao:row", RowQuery)
	db.Callback().Raw().Register("dao:raw", RawExec)
}
