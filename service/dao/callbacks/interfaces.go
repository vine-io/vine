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

import "github.com/lack-io/vine/service/dao"

type BeforeCreateInterface interface {
	BeforeCreate(*dao.DB) error
}

type AfterCreateInterface interface {
	AfterCreate(*dao.DB) error
}

type BeforeUpdateInterface interface {
	BeforeUpdate(*dao.DB) error
}

type AfterUpdateInterface interface {
	AfterUpdate(*dao.DB) error
}

type BeforeSaveInterface interface {
	BeforeSave(*dao.DB) error
}

type AfterSaveInterface interface {
	AfterSave(*dao.DB) error
}

type BeforeDeleteInterface interface {
	BeforeDelete(*dao.DB) error
}

type AfterDeleteInterface interface {
	AfterDelete(*dao.DB) error
}

type AfterFindInterface interface {
	AfterFind(*dao.DB) error
}
