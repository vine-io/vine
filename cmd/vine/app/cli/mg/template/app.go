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

package template

var (
	SingleApp = `package app

import (
	"context"

	"{{.Dir}}/pkg/runtime/inject"
	"github.com/vine-io/vine"

	"{{.Dir}}/pkg/biz"
)

func init() {
	inject.ProvidePanic(new({{.Name}}))
}

type {{title .Name}} interface {
	Init() error
	Call(ctx context.Context, name string) (string, error)
	Stream()
	PingPong()
}

var _ {{title .Name}} = (*{{.Name}})(nil)

type {{.Name}} struct {
	vine.Service ` + "`inject:\"\"`" + `

	Caller biz.Caller ` + "`inject:\"\"`" + `
}

func (s *{{.Name}}) Init() error {
	return nil
}

func (s *{{.Name}}) Call(ctx context.Context, name string) (string, error) {
	return s.Caller.Call(ctx, name)
}

func (s *{{.Name}}) Stream() {
	panic("implement me")
}

func (s *{{.Name}}) PingPong() {
	panic("implement me")
}
`

	ClusterApp = `package app

import (
	"context"

	"{{.Dir}}/pkg/runtime/inject"
	"github.com/vine-io/vine"

	"{{.Dir}}/pkg/{{.Name}}/biz"
)

func init() {
	inject.ProvidePanic(new({{.Name}}))
}

type {{title .Name}} interface {
	Init() error
	Call(ctx context.Context, name string) (string, error)
	Stream()
	PingPong()
}

var _ {{title .Name}} = (*{{.Name}})(nil)

type {{.Name}} struct {
	vine.Service ` + "`inject:\"\"`" + `

	Caller biz.Caller ` + "`inject:\"\"`" + `
}

func (s *{{.Name}}) Init() error {
	return nil
}

func (s *{{.Name}}) Call(ctx context.Context, name string) (string, error) {
	return s.Caller.Call(ctx, name)
}

func (s *{{.Name}}) Stream() {
	panic("implement me")
}

func (s *{{.Name}}) PingPong() {
	panic("implement me")
}
`
)