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

// Package sync is an interface for distributed synchronization
package sync

import (
	"errors"
)

var (
	ErrLockTimeout = errors.New("lock timeout")
)

type Role string

const (
	Primary Role = "primary"
	Follow  Role = "follow"
)

// Sync is an interface for distributed synchronization
type Sync interface {
	// Init Initialise options
	Init(...Option) error
	// Options Return the options
	Options() Options
	// Leader Elect a leader
	Leader(id string, opts ...LeaderOption) (Leader, error)
	// ListMembers get all election member
	ListMembers(opts ...ListMembersOption) ([]*Member, error)
	// Lock acquires a lock
	Lock(id string, opts ...LockOption) error
	// Unlock releases a lock
	Unlock(id string) error
	// String Sync implementation
	String() string
}

type Member struct {
	Id        string `json:"id"`
	Namespace string `json:"namespace"`
	Role      Role   `json:"role"`
}

// Leader provides leadership election
type Leader interface {
	// Resign resigns leadership
	Resign() error
	// Status returns when leadership is lost
	Status() chan bool
}
