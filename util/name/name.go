// MIT License
//
// Copyright (c) 2020 The vine Authors
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
	n.pick(count, lower+upper)
	return n
}

// Digit appends digit character to name
func (n *name) Digit(count int) *name {
	n.pick(count, digit)
	return n
}

// Any appends digits or letters character to name
func (n *name) Any(count int) *name {
	n.pick(count, digit+lower+upper)
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
