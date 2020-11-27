package mock

import (
	"github.com/lack-io/vine/registry"
	"github.com/lack-io/vine/server"
)

type MockHandler struct {
	Id   string
	Opts server.HandlerOptions
	Hdlr interface{}
}

func (m *MockHandler) Name() string {
	return m.Id
}

func (m *MockHandler) Handler() interface{} {
	return m.Hdlr
}

func (m *MockHandler) Endpoints() []*registry.Endpoint {
	return []*registry.Endpoint{}
}

func (m *MockHandler) Options() server.HandlerOptions {
	return m.Opts
}
