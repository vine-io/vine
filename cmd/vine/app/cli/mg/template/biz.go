package template

var (
	DomainSRV = `package biz

import (
	"context"

	"{{.Dir}}/pkg/runtime/inject"
)

type Caller interface {
	CallS(ctx context.Context, name string) (string, error)
}

func init() {
	inject.ProvidePanic(new(caller))
}

type caller struct {
	// inject Repo
}

func (c *caller) CallS(ctx context.Context, name string) (string, error) {
	return "reply: " + name, nil
}
`
)
