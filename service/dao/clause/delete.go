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

package clause

type Delete struct {
	Modifier string
}

func (delete Delete) Name() string {
	return "DELETE"
}

func (delete Delete) Build(builder Builder) {
	builder.WriteString("DELETE")

	if delete.Modifier != "" {
		builder.WriteByte(' ')
		builder.WriteString(delete.Modifier)
	}
}

func (delete Delete) MergeClause(clause *Clause) {
	clause.Name = ""
	clause.Expression = delete
}
