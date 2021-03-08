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

func BeginTransaction(db *dao.DB) {
	if !db.Options.SkipDefaultTransaction {
		if tx := db.Begin(); tx.Error == nil {
			db.Statement.ConnPool = tx.Statement.ConnPool
			db.InstanceSet("dao:started_transaction", true)
		} else if tx.Error == dao.ErrInvalidTransaction {
			tx.Error = nil
		}
	}
}

func CommitOrRollbackTransaction(db *dao.DB) {
	if !db.Options.SkipDefaultTransaction {
		if _, ok := db.InstanceGet("dao:started_transaction"); ok {
			if db.Error == nil {
				db.Commit()
			} else {
				db.Rollback()
			}
			db.Statement.ConnPool = db.ConnPool
		}
	}
}
