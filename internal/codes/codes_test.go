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

package codes

import (
	"testing"

	"github.com/stretchr/testify/assert"

)

func TestCodes_Call(t *testing.T) {
	c := New(Unknown).Call()
	assert.NotEqual(t, c.s.Pos, "")

	t.Log(c)
}

func TestCodes_Stack(t *testing.T) {
	c := New(Unknown).Stack("stack")
	assert.NotEqual(t, len(c.s.Details), 0)
	assert.Equal(t, c.s.Details[0].Desc, "stack")

	t.Log(c)
}