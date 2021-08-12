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

package callbacks

import (
	"github.com/vine-io/vine/lib/dao"
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
