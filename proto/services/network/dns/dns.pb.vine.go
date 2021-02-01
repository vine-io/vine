// Code generated by proto-gen-vine. DO NOT EDIT.
// source: github.com/lack-io/vine/proto/services/network/dns/dns.proto

package dns

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
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

// API Endpoints for Dns service
func NewDnsEndpoints() []*apipb.Endpoint {
	return []*apipb.Endpoint{}
}

// Client API for Dns service
type DnsService interface {
	Advertise(ctx context.Context, in *AdvertiseRequest, opts ...client.CallOption) (*AdvertiseResponse, error)
	Remove(ctx context.Context, in *RemoveRequest, opts ...client.CallOption) (*RemoveResponse, error)
	Resolve(ctx context.Context, in *ResolveRequest, opts ...client.CallOption) (*ResolveResponse, error)
}

type dnsService struct {
	c    client.Client
	name string
}

func NewDnsService(name string, c client.Client) DnsService {
	return &dnsService{
		c:    c,
		name: name,
	}
}

func (c *dnsService) Advertise(ctx context.Context, in *AdvertiseRequest, opts ...client.CallOption) (*AdvertiseResponse, error) {
	req := c.c.NewRequest(c.name, "Dns.Advertise", in)
	out := new(AdvertiseResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dnsService) Remove(ctx context.Context, in *RemoveRequest, opts ...client.CallOption) (*RemoveResponse, error) {
	req := c.c.NewRequest(c.name, "Dns.Remove", in)
	out := new(RemoveResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dnsService) Resolve(ctx context.Context, in *ResolveRequest, opts ...client.CallOption) (*ResolveResponse, error) {
	req := c.c.NewRequest(c.name, "Dns.Resolve", in)
	out := new(ResolveResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Dns service
type DnsHandler interface {
	Advertise(context.Context, *AdvertiseRequest, *AdvertiseResponse) error
	Remove(context.Context, *RemoveRequest, *RemoveResponse) error
	Resolve(context.Context, *ResolveRequest, *ResolveResponse) error
}

func RegisterDnsHandler(s server.Server, hdlr DnsHandler, opts ...server.HandlerOption) error {
	type dnsImpl interface {
		Advertise(ctx context.Context, in *AdvertiseRequest, out *AdvertiseResponse) error
		Remove(ctx context.Context, in *RemoveRequest, out *RemoveResponse) error
		Resolve(ctx context.Context, in *ResolveRequest, out *ResolveResponse) error
	}
	type Dns struct {
		dnsImpl
	}
	h := &dnsHandler{hdlr}
	return s.Handle(s.NewHandler(&Dns{h}, opts...))
}

type dnsHandler struct {
	DnsHandler
}

func (h *dnsHandler) Advertise(ctx context.Context, in *AdvertiseRequest, out *AdvertiseResponse) error {
	return h.DnsHandler.Advertise(ctx, in, out)
}

func (h *dnsHandler) Remove(ctx context.Context, in *RemoveRequest, out *RemoveResponse) error {
	return h.DnsHandler.Remove(ctx, in, out)
}

func (h *dnsHandler) Resolve(ctx context.Context, in *ResolveRequest, out *ResolveResponse) error {
	return h.DnsHandler.Resolve(ctx, in, out)
}
