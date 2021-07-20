package template

var (
	ServiceSRV = `package service

import (
	"{{.Dir}}/pkg/runtime/inject"
	"github.com/lack-io/vine"
)

func init() {
	inject.ProvidePanic(new({{.Name}}))
}

type {{title .Name}} interface {
	Init() error
	Call()
	Stream()
	PingPong()
}

var _ {{title .Name}} = (*{{.Name}})(nil)

type {{.Name}} struct {
	vine.Service ` + "`inject:\"\"`" + `
}

func (s *{{.Name}}) Init() error {
	return nil
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
