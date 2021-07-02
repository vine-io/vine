package template

var (
	ServiceSRV = `package service

import (
	"fmt"

	"github.com/lack-io/foo/pkg/runtime/inject"
	"github.com/lack-io/vine"
)

func init() {
	if err := inject.Provide(new({{.Name}})); err != nil {
		panic(err)
	}
}

type {{title .Name}} interface {
	Init() error
	Call()
	Stream()
	PingPong()
}

var _ {{title .Name}} = (*{{.Name}})(nil)

type {{.Name}} struct {
	vine.Service +"` + "`inject:\"\"`" + `
}

func (s *{{.Name}}) Call() {
	// FIXME: modify method
	panic("implement me")
}

func (s *{{.Name}}) Stream() {
	panic("implement me")
}

func (s *{{.Name}}) PingPong() {
	panic("implement me")
}
`
)
