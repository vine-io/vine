// Code generated by proto-gen-vine. DO NOT EDIT.
// source: github.com/lack-io/vine/proto/services/registry/registry.proto

package registry

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	registry "github.com/lack-io/vine/proto/apis/registry"
	math "math"
)

import (
	context "context"
	apipb "github.com/lack-io/vine/proto/apis/api"
	openapi "github.com/lack-io/vine/proto/apis/openapi"
	api "github.com/lack-io/vine/service/api"
	client "github.com/lack-io/vine/service/client"
	server "github.com/lack-io/vine/service/server"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ apipb.Endpoint
var _ api.Option
var _ openapi.OpenAPI
var _ context.Context
var _ client.Option
var _ server.Option

// API Endpoints for Registry service
func NewRegistryEndpoints() []*apipb.Endpoint {
	return []*apipb.Endpoint{}
}

// Client API for Registry service
type RegistryService interface {
	GetService(ctx context.Context, in *GetRequest, opts ...client.CallOption) (*GetResponse, error)
	Register(ctx context.Context, in *registry.Service, opts ...client.CallOption) (*EmptyResponse, error)
	Deregister(ctx context.Context, in *registry.Service, opts ...client.CallOption) (*EmptyResponse, error)
	ListServices(ctx context.Context, in *ListRequest, opts ...client.CallOption) (*ListResponse, error)
	Watch(ctx context.Context, in *WatchRequest, opts ...client.CallOption) (Registry_WatchService, error)
}

type registryService struct {
	c    client.Client
	name string
}

func NewRegistryService(name string, c client.Client) RegistryService {
	return &registryService{
		c:    c,
		name: name,
	}
}

func (c *registryService) GetService(ctx context.Context, in *GetRequest, opts ...client.CallOption) (*GetResponse, error) {
	req := c.c.NewRequest(c.name, "Registry.GetService", in)
	out := new(GetResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *registryService) Register(ctx context.Context, in *registry.Service, opts ...client.CallOption) (*EmptyResponse, error) {
	req := c.c.NewRequest(c.name, "Registry.Register", in)
	out := new(EmptyResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *registryService) Deregister(ctx context.Context, in *registry.Service, opts ...client.CallOption) (*EmptyResponse, error) {
	req := c.c.NewRequest(c.name, "Registry.Deregister", in)
	out := new(EmptyResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *registryService) ListServices(ctx context.Context, in *ListRequest, opts ...client.CallOption) (*ListResponse, error) {
	req := c.c.NewRequest(c.name, "Registry.ListServices", in)
	out := new(ListResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *registryService) Watch(ctx context.Context, in *WatchRequest, opts ...client.CallOption) (Registry_WatchService, error) {
	req := c.c.NewRequest(c.name, "Registry.Watch", &WatchRequest{})
	stream, err := c.c.Stream(ctx, req, opts...)
	if err != nil {
		return nil, err
	}
	if err := stream.Send(in); err != nil {
		return nil, err
	}
	return &registryServiceWatch{stream}, nil
}

type Registry_WatchService interface {
	Context() context.Context
	SendMsg(interface{}) error
	RecvMsg(interface{}) error
	Close() error
	Recv() (*registry.Result, error)
}

type registryServiceWatch struct {
	stream client.Stream
}

func (x *registryServiceWatch) Close() error {
	return x.stream.Close()
}

func (x *registryServiceWatch) Context() context.Context {
	return x.stream.Context()
}

func (x *registryServiceWatch) SendMsg(m interface{}) error {
	return x.stream.Send(m)
}

func (x *registryServiceWatch) RecvMsg(m interface{}) error {
	return x.stream.Recv(m)
}

func (x *registryServiceWatch) Recv() (*registry.Result, error) {
	m := new(registry.Result)
	err := x.stream.Recv(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Registry service
type RegistryHandler interface {
	GetService(context.Context, *GetRequest, *GetResponse) error
	Register(context.Context, *registry.Service, *EmptyResponse) error
	Deregister(context.Context, *registry.Service, *EmptyResponse) error
	ListServices(context.Context, *ListRequest, *ListResponse) error
	Watch(context.Context, *WatchRequest, Registry_WatchStream) error
}

func RegisterRegistryHandler(s server.Server, hdlr RegistryHandler, opts ...server.HandlerOption) error {
	type registryImpl interface {
		GetService(ctx context.Context, in *GetRequest, out *GetResponse) error
		Register(ctx context.Context, in *registry.Service, out *EmptyResponse) error
		Deregister(ctx context.Context, in *registry.Service, out *EmptyResponse) error
		ListServices(ctx context.Context, in *ListRequest, out *ListResponse) error
		Watch(ctx context.Context, stream server.Stream) error
	}
	type Registry struct {
		registryImpl
	}
	h := &registryHandler{hdlr}
	return s.Handle(s.NewHandler(&Registry{h}, opts...))
}

type registryHandler struct {
	RegistryHandler
}

func (h *registryHandler) GetService(ctx context.Context, in *GetRequest, out *GetResponse) error {
	return h.RegistryHandler.GetService(ctx, in, out)
}

func (h *registryHandler) Register(ctx context.Context, in *registry.Service, out *EmptyResponse) error {
	return h.RegistryHandler.Register(ctx, in, out)
}

func (h *registryHandler) Deregister(ctx context.Context, in *registry.Service, out *EmptyResponse) error {
	return h.RegistryHandler.Deregister(ctx, in, out)
}

func (h *registryHandler) ListServices(ctx context.Context, in *ListRequest, out *ListResponse) error {
	return h.RegistryHandler.ListServices(ctx, in, out)
}

func (h *registryHandler) Watch(ctx context.Context, stream server.Stream) error {
	m := new(WatchRequest)
	if err := stream.Recv(m); err != nil {
		return err
	}
	return h.RegistryHandler.Watch(ctx, m, &registryWatchStream{stream})
}

type Registry_WatchStream interface {
	Context() context.Context
	SendMsg(interface{}) error
	RecvMsg(interface{}) error
	Close() error
	Send(*registry.Result) error
}

type registryWatchStream struct {
	stream server.Stream
}

func (x *registryWatchStream) Close() error {
	return x.stream.Close()
}

func (x *registryWatchStream) Context() context.Context {
	return x.stream.Context()
}

func (x *registryWatchStream) SendMsg(m interface{}) error {
	return x.stream.Send(m)
}

func (x *registryWatchStream) RecvMsg(m interface{}) error {
	return x.stream.Recv(m)
}

func (x *registryWatchStream) Send(m *registry.Result) error {
	return x.stream.Send(m)
}
