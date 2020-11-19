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

package name

import (
	"bytes"
	"math/rand"
	"time"
)

const (
	lower = "abcdefghijklmnopqrstuvwxyz"
	upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digit = "0123456789"
)

type name struct {
	b *bytes.Buffer
}

// New creates a name
func New() *name {
	return &name{b: bytes.NewBuffer([]byte(""))}
}

// Lower appends lowercase character to name
func (n *name) Lower(count int) *name {
	n.pick(count, lower)
	return n
}

// Upper appends uppercase character to name
func (n *name) Upper(count int) *name {
	n.pick(count, upper)
	return n
}

// Letter appends letter character to name
func (n *name) Letter(count int) *name {
	n.pick(count, lower + upper)
	return n
}

// Digit appends digit character to name
func (n *name) Digit(count int) *name {
	n.pick(count, digit)
	return n
}

// Any appends digits or letters character to name
func (n *name) Any(count int) *name {
	n.pick(count, digit + lower + upper)
	return n
}

// String implements interface Stringer
func (n *name) String() string {
	return n.b.String()
}

// pick get specified number characters from string
func (n *name) pick(c int, from string) {
	l := len(from)
	for i := 0; i < c; i++ {
		rd := randInt63n(int64(l))
		n.b.WriteByte(from[rd])
	}
}

// randInt63n returns, as an int64, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0. Uses time.Now().UnixNano() for seed.
func randInt63n(n int64) int64 {
	return rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(n)
}
