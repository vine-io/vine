// MIT License
//
// Copyright (c) 2021 Lack
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

package runtime

import (
	"context"
	"strings"
	"sync"

	"github.com/vine-io/vine/lib/dao"
	"github.com/vine-io/vine/lib/dao/clause"
)

// GroupVersionKind contains the information of Entity, etc Group, Version, Kind
type GroupVersionKind struct {
	Group   string
	Version string
	Kind    string
}

func (gvk *GroupVersionKind) APIGroup() string {
	if gvk.Group == "" {
		return gvk.Version
	}
	return gvk.Group + "/" + gvk.Version
}

func (gvk *GroupVersionKind) String() string {
	var s string
	if gvk.Group != "" {
		s = gvk.Group + "/"
	}
	if gvk.Version != "" {
		s = s + gvk.Version + "."
	}
	return s + gvk.Kind
}

func FromGVK(s string) *GroupVersionKind {
	gvk := &GroupVersionKind{}
	if idx := strings.Index(s, "/"); idx != -1 {
		gvk.Group = s[:idx]
		s = s[idx+1:]
	}
	if idx := strings.Index(s, "."); idx != -1 {
		gvk.Version = s[:idx]
		s = s[idx+1:]
	} else {
		gvk.Version = "v1"
	}
	gvk.Kind = s
	return gvk
}

// Entity is an interface that describes protocol message
type Entity interface {
	// GVK get the GroupVersionKind of Entity
	GVK() *GroupVersionKind
	// DeepCopyFrom deep copy the struct from another
	DeepCopyFrom(Entity)
	// DeepCopy deep copy the struct
	DeepCopy() Entity
}

var eset = NewEntitySet()

type EntitySet struct {
	sync.RWMutex

	sets map[string]Entity

	OnCreate func(in Entity) Entity
}

// NewEntity creates a new entity, trigger OnCreate function
func (os *EntitySet) NewEntity(gvk string) (Entity, bool) {
	os.RLock()
	defer os.RUnlock()
	out, ok := os.sets[gvk]
	if !ok {
		return nil, false
	}
	return os.OnCreate(out.DeepCopy()), true
}

// NewEntWithGVK creates a new entity, trigger OnCreate function
func (os *EntitySet) NewEntWithGVK(gvk *GroupVersionKind) (Entity, bool) {
	os.RLock()
	defer os.RUnlock()
	out, ok := os.sets[gvk.String()]
	if !ok {
		return nil, false
	}
	return os.OnCreate(out.DeepCopy()), true
}

func (os *EntitySet) IsExists(gvk *GroupVersionKind) bool {
	os.RLock()
	defer os.RUnlock()
	_, ok := os.sets[gvk.String()]
	return ok
}

func (os *EntitySet) GetEntity(gvk *GroupVersionKind) (Entity, bool) {
	os.RLock()
	defer os.RUnlock()
	out, ok := os.sets[gvk.String()]
	if !ok {
		return nil, false
	}
	return out.DeepCopy(), true
}

// AddEnt push entities to Set
func (os *EntitySet) AddEnt(v ...Entity) {
	os.Lock()
	for _, in := range v {
		os.sets[in.GVK().String()] = in
	}
	os.Unlock()
}

// NewEntity creates a new entity, trigger OnCreate function
func NewEntity(gvk string) (Entity, bool) {
	return eset.NewEntity(gvk)
}

// NewEntWithGVK creates a new entity, trigger OnCreate function
func NewEntWithGVK(gvk *GroupVersionKind) (Entity, bool) {
	return eset.NewEntWithGVK(gvk)
}

func AddEnt(v ...Entity) {
	eset.AddEnt(v...)
}

func NewEntitySet() *EntitySet {
	return &EntitySet{
		sets:     map[string]Entity{},
		OnCreate: func(in Entity) Entity { return in },
	}
}

type Repo interface {
	FindPage(ctx context.Context, page, size int32) ([]Entity, int64, error)
	FindAll(ctx context.Context) ([]Entity, error)
	FindPureAll(ctx context.Context) ([]Entity, error)
	Count(ctx context.Context) (total int64, err error)
	FindOne(ctx context.Context) (Entity, error)
	FindPureOne(ctx context.Context) (Entity, error)
	Cond(exprs ...clause.Expression) Repo
	Create(ctx context.Context) (Entity, error)
	BatchUpdates(ctx context.Context) error
	Updates(ctx context.Context) (Entity, error)
	BatchDelete(ctx context.Context, soft bool) error
	Delete(ctx context.Context, soft bool) error
	Tx(ctx context.Context) *dao.DB
}

var pset = RepoSet{sets: map[string]func(Entity) Repo{}}

type RepoSet struct {
	sync.RWMutex
	sets map[string]func(Entity) Repo
}

func (s *RepoSet) RegisterRepo(g *GroupVersionKind, fn func(Entity) Repo) {
	s.Lock()
	s.sets[g.String()] = fn
	s.Unlock()
}

func (s *RepoSet) NewRepo(in Entity) (Repo, bool) {
	s.RLock()
	defer s.RUnlock()
	fn, ok := s.sets[in.GVK().String()]
	if !ok {
		return nil, false
	}
	return fn(in), true
}

func NewRepoSet() *RepoSet {
	return &RepoSet{sets: map[string]func(Entity) Repo{}}
}

func RegisterRepo(g *GroupVersionKind, fn func(Entity) Repo) {
	pset.RegisterRepo(g, fn)
}

func NewRepo(in Entity) (Repo, bool) {
	return pset.NewRepo(in)
}
