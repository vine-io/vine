package grpc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	membroker "github.com/vine-io/vine/core/broker/memory"
	regMemory "github.com/vine-io/vine/core/registry/memory"
	"github.com/vine-io/vine/core/server"
	grpcServer "github.com/vine-io/vine/core/server/grpc"
)

func TestNewServer(t *testing.T) {
	options := make([]server.Option, 0)
	a := 0

	reg := regMemory.NewRegistry()
	if err := reg.Init(); err != nil {
		t.Fatal(err)
	}

	b := membroker.NewBroker()
	b.Init()

	options = append(options,
		server.Registry(reg),
		server.Broker(b),
		grpcServer.WrapGRPCServer(func(s *grpc.Server) error {
			a += 1
			return nil
		}),
	)

	s := grpcServer.NewServer()

	if err := s.Init(options...); err != nil {
		t.Fatal(err)
	}

	_ = s.Start()

	assert.Equal(t, a, 1)
}
