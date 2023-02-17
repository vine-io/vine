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
	"context"
	"errors"

	"github.com/spf13/pflag"
)

var (
	ErrLockTimeout = errors.New("lock timeout")

	Flag = pflag.NewFlagSet("sync", pflag.ExitOnError)
)

func init() {
	Flag.String("sync-default", "", "Sync for vine")
}

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
	Leader(ctx context.Context, name string, opts ...LeaderOption) (Leader, error)
	// ListMembers get all election member
	ListMembers(ctx context.Context, opts ...ListMembersOption) ([]*Member, error)
	// WatchElect watch leader event
	WatchElect(ctx context.Context, opts ...WatchElectOption) (ElectWatcher, error)
	// Lock acquires a lock
	Lock(ctx context.Context, id string, opts ...LockOption) error
	// Unlock releases a lock
	Unlock(ctx context.Context, id string) error
	// String Sync implementation
	String() string
}

type Member struct {
	Leader    string `json:"leader"`
	Id        string `json:"id"`
	Namespace string `json:"namespace"`
	Role      Role   `json:"role"`
}

type ObserveResult struct {
	Namespace string `json:"namespace"`
	Id        string `json:"id"`
}

// Leader provides leadership election
type Leader interface {
	// Id leader node
	Id() string
	// Resign resigns leadership
	Resign() error
	// Observe watch leadership event
	Observe() chan ObserveResult
	// Primary get the info of primary role
	Primary() (*Member, error)
	// Status returns when leadership is lost
	Status() chan bool
}

// ElectWatcher watch election event
type ElectWatcher interface {
	// Next is a blocking call
	Next() (*Member, error)
	Close()
}
