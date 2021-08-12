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

package memory

import (
	"time"

	"github.com/vine-io/vine/lib/config/reader"
)

type value struct{}

func newValue() reader.Value {
	return new(value)
}

func (v *value) Bool(def bool) bool {
	return false
}

func (v *value) Int(def int64) int64 {
	return 0
}

func (v *value) String(def string) string {
	return ""
}

func (v *value) Float64(def float64) float64 {
	return 0.0
}

func (v *value) Duration(def time.Duration) time.Duration {
	return time.Duration(0)
}

func (v *value) StringSlice(def []string) []string {
	return []string{}
}

func (v *value) StringMap(def map[string]string) map[string]string {
	return map[string]string{}
}

func (v *value) Scan(val interface{}) error {
	return nil
}

func (v *value) Bytes() []byte {
	return nil
}
