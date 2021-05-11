package template

var (
	ServiceSRV = `package service

import "github.com/lack-io/vine"

type {{title .Name}} interface {
	Call()
	Stream()
	PingPong()
}

var _ {{title .Name}} = (*{{.Name}})(nil)

type {{.Name}} struct {
	vine.Service
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

func new{{title .Name}}(s vine.Service) *{{.Name}} {
	return &{{.Name}}{Service: s}
}`

	Wire = `// +build wireinject

package service

import (
	"github.com/google/wire"

	"github.com/lack-io/vine"
)

func New(s vine.Service) *{{.Name}} {
	wire.Build(new{{title .Name}})
	return nil
}
`
)
