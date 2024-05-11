package grpc_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

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

func TestWithKeepalive(t *testing.T) {
	var kaep = keepalive.EnforcementPolicy{
		MinTime:             15 * time.Second, // If a client pings more than once every 15 seconds, terminate the connection
		PermitWithoutStream: true,             // Allow pings even when there are no active streams
	}

	var kasp = keepalive.ServerParameters{
		MaxConnectionIdle:     30 * time.Second, // If a client is idle for 30 seconds, send a GOAWAY
		MaxConnectionAge:      45 * time.Second, // If any connection is alive for more than 45 seconds, send a GOAWAY
		MaxConnectionAgeGrace: 15 * time.Second, // Allow 15 seconds for pending RPCs to complete before forcibly closing connections
		Time:                  15 * time.Second, // Ping the client if it is idle for 15 seconds to ensure the connection is still active
		Timeout:               5 * time.Second,  // Wait 5 second for the ping ack before assuming the connection is dead
	}

	opts := []grpc.ServerOption{grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp)}
	options := make([]server.Option, 0)
	reg := regMemory.NewRegistry()
	if err := reg.Init(); err != nil {
		t.Fatal(err)
	}

	b := membroker.NewBroker()
	b.Init()

	options = append(options,
		server.Registry(reg),
		server.Broker(b),
		grpcServer.Options(opts...),
	)

	s := grpcServer.NewServer()

	if err := s.Init(options...); err != nil {
		t.Fatal(err)
	}

	_ = s.Start()
}
