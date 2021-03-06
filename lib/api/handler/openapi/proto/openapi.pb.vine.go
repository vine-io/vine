// Code generated by proto-gen-vine. DO NOT EDIT.
// source: github.com/vine-io/vine/lib/api/handler/openapi/proto/openapi.proto

package openapi

import (
	context "context"
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	vine "github.com/vine-io/vine"
	client "github.com/vine-io/vine/core/client"
	server "github.com/vine-io/vine/core/server"
	api "github.com/vine-io/vine/lib/api"
	math "math"
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

// API Endpoints for OpenAPIService service
func NewOpenAPIServiceEndpoints() []*api.Endpoint {
	return []*api.Endpoint{}
}

// Client API for OpenAPIService service
type OpenAPIService interface {
	GetOpenAPIDoc(ctx context.Context, in *GetOpenAPIDocRequest, opts ...client.CallOption) (*GetOpenAPIDocResponse, error)
}

type openAPIService struct {
	c    client.Client
	name string
}

func NewOpenAPIService(name string, c client.Client) OpenAPIService {
	return &openAPIService{
		c:    c,
		name: name,
	}
}

func (c *openAPIService) GetOpenAPIDoc(ctx context.Context, in *GetOpenAPIDocRequest, opts ...client.CallOption) (*GetOpenAPIDocResponse, error) {
	req := c.c.NewRequest(c.name, "OpenAPIService.GetOpenAPIDoc", in)
	out := new(GetOpenAPIDocResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for OpenAPIService service
type OpenAPIServiceHandler interface {
	GetOpenAPIDoc(*vine.Context, *GetOpenAPIDocRequest, *GetOpenAPIDocResponse) error
}

func RegisterOpenAPIServiceHandler(s server.Server, hdlr OpenAPIServiceHandler, opts ...server.HandlerOption) error {
	type openAPIServiceImpl interface {
		GetOpenAPIDoc(ctx context.Context, in *GetOpenAPIDocRequest, out *GetOpenAPIDocResponse) error
	}
	type OpenAPIService struct {
		openAPIServiceImpl
	}
	h := &openAPIServiceHandler{hdlr}
	return s.Handle(s.NewHandler(&OpenAPIService{h}, opts...))
}

type openAPIServiceHandler struct {
	OpenAPIServiceHandler
}

func (h *openAPIServiceHandler) GetOpenAPIDoc(ctx context.Context, in *GetOpenAPIDocRequest, out *GetOpenAPIDocResponse) error {
	return h.OpenAPIServiceHandler.GetOpenAPIDoc(vine.InitContext(ctx), in, out)
}
