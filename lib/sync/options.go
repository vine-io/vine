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
	TTL       int64
	Namespace string
	Id        string
}

type LeaderOption func(o *LeaderOptions)

// LeaderTTL sets the leader ttl
func LeaderTTL(t int64) LeaderOption {
	return func(o *LeaderOptions) {
		o.TTL = t
	}
}

// LeaderNS sets the leader namespace
func LeaderNS(ns string) LeaderOption {
	return func(o *LeaderOptions) {
		o.Namespace = ns
	}
}

// LeaderId sets the leader id
func LeaderId(id string) LeaderOption {
	return func(o *LeaderOptions) {
		o.Id = id
	}
}

type ListMembersOptions struct {
	Namespace string
}

type ListMembersOption func(o *ListMembersOptions)

// MemberNS sets the list member namespace
func MemberNS(ns string) ListMembersOption {
	return func(o *ListMembersOptions) {
		o.Namespace = ns
	}
}

type WatchElectOptions struct {
	Namespace string
	Id        string
}

type WatchElectOption func(o *WatchElectOptions)

// WatchNS sets the watch elector namespace
func WatchNS(ns string) WatchElectOption {
	return func(o *WatchElectOptions) {
		o.Namespace = ns
	}
}

// WatchId sets the watch elector id
func WatchId(id string) WatchElectOption {
	return func(o *WatchElectOptions) {
		o.Id = id
	}
}

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
