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

// APIGroup contains the information of Object, etc Group, Version, Name
type APIGroup struct {
	Group   string
	Version string
	Name    string
}

func (g *APIGroup) String() string {
	var s string
	if g.Group != "" {
		s = g.Group + "/"
	}
	if g.Version != "" {
		s = g.Version + "."
	}
	return s + g.Name
}

// Object is an interface that describes protocol message
type Object interface {
	// GetAPIGroup get the APIGroup of Object
	GetAPIGroup() *APIGroup
	// DeepCopy deep copy the struct
	DeepCopy() Object
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

var schemaSet = SchemaSet{sets: map[*APIGroup]func(Object) Schema{}}

type SchemaSet struct {
	sync.RWMutex
	sets map[*APIGroup]func(Object) Schema
}

func (s *SchemaSet) RegistrySchema(g *APIGroup, fn func(Object) Schema) {
	s.Lock()
	s.sets[g] = fn
	s.Unlock()
}

func (s *SchemaSet) NewSchema(in Object) (Schema, bool) {
	s.RLock()
	defer s.RUnlock()
	fn, ok := s.sets[in.GetAPIGroup()]
	if !ok {
		return nil, false
	}
	return fn(in), true
}

func NewSchemaSet() *SchemaSet {
	return &SchemaSet{sets: map[*APIGroup]func(Object) Schema{}}
}

func RegistrySchema(g *APIGroup, fn func(Object) Schema) {
	schemaSet.RegistrySchema(g, fn)
}

func NewSchema(in Object) (Schema, bool) {
	return schemaSet.NewSchema(in)
}
