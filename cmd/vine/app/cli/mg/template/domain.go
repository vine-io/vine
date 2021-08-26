package template

var (
	DomainSRV = `package domain

import (
	"context"

	"{{.Dir}}/pkg/runtime/inject"
)

type Caller interface {
	Call(ctx context.Context, name string) (string, error)
}

func init() {
	inject.ProvidePanic(new(caller))
}

type caller struct {
	// inject Repo
}

func (c *caller) Call(ctx context.Context, name string) (string, error) {
	return "reply: " + name, nil
}
`
)
