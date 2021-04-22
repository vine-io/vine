package template

var (
	ServiceSRV = `package service

import "github.com/lack-io/vine"

type {{title .Alias}} interface {
	Call()
	Stream()
	PingPong()
}

var _ {{title .Alias}} = (*{{.Alias}})(nil)

type {{.Alias}} struct {
	vine.Service
}

func (s *{{.Alias}}) Call() {
	// FIXME: modify method
	panic("implement me")
}

func (s *{{.Alias}}) Stream() {
	panic("implement me")
}

func (s *{{.Alias}}) PingPong() {
	panic("implement me")
}

func New(s vine.Service) *{{.Alias}} {
	return &{{.Alias}}{Service: s}
}`
)
