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

package sync

import (
	"time"
)

type Options struct {
	Nodes  []string
	Prefix string
}

type Option func(o *Options)

// Nodes sets the addresses to use
func Nodes(a ...string) Option {
	return func(o *Options) {
		o.Nodes = a
	}
}

// Prefix sets a prefix to any lock ids used
func Prefix(p string) Option {
	return func(o *Options) {
		o.Prefix = p
	}
}

type LeaderOptions struct {
	TTL int64
}

type LeaderOption func(o *LeaderOptions)

type ListMembersOptions struct{}

type ListMembersOption func(o *ListMembersOptions)

type LockOptions struct {
	TTL  time.Duration
	Wait time.Duration
}

type LockOption func(o *LockOptions)

// LockTTL sets the lock ttl
func LockTTL(t time.Duration) LockOption {
	return func(o *LockOptions) {
		o.TTL = t
	}
}

// LockWait sets the wait time
func LockWait(t time.Duration) LockOption {
	return func(o *LockOptions) {
		o.Wait = t
	}
}
