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
	"sync"

	"github.com/lack-io/vine/lib/dao"
	"github.com/lack-io/vine/lib/dao/clause"
)

// GroupVersionKind contains the information of Object, etc Group, Version, Name
type GroupVersionKind struct {
	Group   string
	Version string
	Kind    string
}

func (gvk *GroupVersionKind) String() string {
	var s string
	if gvk.Group != "" {
		s = gvk.Group + "/"
	}
	if gvk.Version != "" {
		s = gvk.Version + "."
	}
	return s + gvk.Kind
}

// Object is an interface that describes protocol message
type Object interface {
	// APIGroup get the GroupVersionKind of Object
	APIGroup() *GroupVersionKind
	// DeepCopy deep copy the struct
	DeepCopy() Object
}

var oset = NewObjectSet()

type ObjectSet struct {
	sync.RWMutex

	sets    map[string]Object
	gvkSets map[*GroupVersionKind]Object

	OnCreate func(in Object) Object
}

// NewObj creates a new object
func (os *ObjectSet) NewObj(name string) (Object, bool) {
	os.RLock()
	defer os.RUnlock()
	out, ok := os.sets[name]
	if !ok {
		return nil, false
	}
	return os.OnCreate(out.DeepCopy()), true
}

func (os *ObjectSet) NewObjWithGVK(gvk *GroupVersionKind) (Object, bool) {
	os.RLock()
	defer os.RUnlock()
	out, ok := os.gvkSets[gvk]
	if !ok {
		return nil, false
	}
	return os.OnCreate(out.DeepCopy()), true
}

func (os *ObjectSet) AddObj(v ...Object) {
	os.Lock()
	for _, in := range v {
		os.gvkSets[in.APIGroup()] = in
		os.sets[in.APIGroup().String()] = in
	}
	os.Unlock()
}

// NewObj creates a new object
func NewObj(name string) (Object, bool) {
	return oset.NewObj(name)
}

func NewObjWithGVK(gvk *GroupVersionKind) (Object, bool) {
	return oset.NewObjWithGVK(gvk)
}

func AddObj(v ...Object) {
	oset.AddObj(v...)
}

func NewObjectSet() *ObjectSet {
	return &ObjectSet{
		sets:     map[string]Object{},
		gvkSets:  map[*GroupVersionKind]Object{},
		OnCreate: func(in Object) Object { return in },
	}
}

type Schema interface {
	FindPage(ctx context.Context, page, size int) ([]Object, int64, error)
	FindAll(ctx context.Context) ([]Object, error)
	FindPureAll(ctx context.Context) ([]Object, error)
	Count(ctx context.Context) (total int64, err error)
	FindOne(ctx context.Context) (Object, error)
	Cond(exprs ...clause.Expression) Schema
	Create(ctx context.Context) (Object, error)
	BatchUpdates(ctx context.Context) error
	Updates(ctx context.Context) (Object, error)
	BatchDelete(ctx context.Context, soft bool) error
	Delete(ctx context.Context, soft bool) error
	Tx(ctx context.Context) *dao.DB
}

var sset = SchemaSet{sets: map[*GroupVersionKind]func(Object) Schema{}}

type SchemaSet struct {
	sync.RWMutex
	sets map[*GroupVersionKind]func(Object) Schema
}

func (s *SchemaSet) RegistrySchema(g *GroupVersionKind, fn func(Object) Schema) {
	s.Lock()
	s.sets[g] = fn
	s.Unlock()
}

func (s *SchemaSet) NewSchema(in Object) (Schema, bool) {
	s.RLock()
	defer s.RUnlock()
	fn, ok := s.sets[in.APIGroup()]
	if !ok {
		return nil, false
	}
	return fn(in), true
}

func NewSchemaSet() *SchemaSet {
	return &SchemaSet{sets: map[*GroupVersionKind]func(Object) Schema{}}
}

func RegistrySchema(g *GroupVersionKind, fn func(Object) Schema) {
	sset.RegistrySchema(g, fn)
}

func NewSchema(in Object) (Schema, bool) {
	return sset.NewSchema(in)
}
