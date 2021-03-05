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

package schema

import "github.com/lack-io/vine/service/dao/clause"

type DataTypeInterface interface {
	DaoDataType() string
}

type CreateClausesInterface interface {
	CreateClauses(*Field) []clause.Interface
}

type QueryClausesInterface interface {
	QueryClauses(*Field) []clause.Interface
}

type UpdateClausesInterface interface {
	UpdateClauses(*Field) []clause.Interface
}

type DeleteClausesInterface interface {
	DeleteClauses(*Field) []clause.Interface
}
