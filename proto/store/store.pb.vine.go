// Code generated by proto-gen-vine. DO NOT EDIT.
// source: github.com/lack-io/vine/proto/store/store.proto

package store

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	math "math"
)

import (
	context "context"
	registry "github.com/lack-io/vine/proto/registry"
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
var _ api.Endpoint
var _ context.Context
var _ client.Option
var _ server.Option
var _ registry.OpenAPI

// API Endpoints for Store service
func NewStoreEndpoints() []*api.Endpoint {
	return []*api.Endpoint{}
}

// Swagger OpenAPI 3.0 for Store service
func NewStoreOpenAPI() *registry.OpenAPI {
	return &registry.OpenAPI{}
}

// Client API for Store service
type StoreService interface {
	Read(ctx context.Context, in *ReadRequest, opts ...client.CallOption) (*ReadResponse, error)
	Write(ctx context.Context, in *WriteRequest, opts ...client.CallOption) (*WriteResponse, error)
	Delete(ctx context.Context, in *DeleteRequest, opts ...client.CallOption) (*DeleteResponse, error)
	List(ctx context.Context, in *ListRequest, opts ...client.CallOption) (Store_ListService, error)
	Databases(ctx context.Context, in *DatabasesRequest, opts ...client.CallOption) (*DatabasesResponse, error)
	Tables(ctx context.Context, in *TablesRequest, opts ...client.CallOption) (*TablesResponse, error)
}

type storeService struct {
	c    client.Client
	name string
}

func NewStoreService(name string, c client.Client) StoreService {
	return &storeService{
		c:    c,
		name: name,
	}
}

func (c *storeService) Read(ctx context.Context, in *ReadRequest, opts ...client.CallOption) (*ReadResponse, error) {
	req := c.c.NewRequest(c.name, "Store.Read", in)
	out := new(ReadResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storeService) Write(ctx context.Context, in *WriteRequest, opts ...client.CallOption) (*WriteResponse, error) {
	req := c.c.NewRequest(c.name, "Store.Write", in)
	out := new(WriteResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storeService) Delete(ctx context.Context, in *DeleteRequest, opts ...client.CallOption) (*DeleteResponse, error) {
	req := c.c.NewRequest(c.name, "Store.Delete", in)
	out := new(DeleteResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storeService) List(ctx context.Context, in *ListRequest, opts ...client.CallOption) (Store_ListService, error) {
	req := c.c.NewRequest(c.name, "Store.List", &ListRequest{})
	stream, err := c.c.Stream(ctx, req, opts...)
	if err != nil {
		return nil, err
	}
	if err := stream.Send(in); err != nil {
		return nil, err
	}
	return &storeServiceList{stream}, nil
}

type Store_ListService interface {
	Context() context.Context
	SendMsg(interface{}) error
	RecvMsg(interface{}) error
	Close() error
	Recv() (*ListResponse, error)
}

type storeServiceList struct {
	stream client.Stream
}

func (x *storeServiceList) Close() error {
	return x.stream.Close()
}

func (x *storeServiceList) Context() context.Context {
	return x.stream.Context()
}

func (x *storeServiceList) SendMsg(m interface{}) error {
	return x.stream.Send(m)
}

func (x *storeServiceList) RecvMsg(m interface{}) error {
	return x.stream.Recv(m)
}

func (x *storeServiceList) Recv() (*ListResponse, error) {
	m := new(ListResponse)
	err := x.stream.Recv(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (c *storeService) Databases(ctx context.Context, in *DatabasesRequest, opts ...client.CallOption) (*DatabasesResponse, error) {
	req := c.c.NewRequest(c.name, "Store.Databases", in)
	out := new(DatabasesResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storeService) Tables(ctx context.Context, in *TablesRequest, opts ...client.CallOption) (*TablesResponse, error) {
	req := c.c.NewRequest(c.name, "Store.Tables", in)
	out := new(TablesResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Store service
type StoreHandler interface {
	Read(context.Context, *ReadRequest, *ReadResponse) error
	Write(context.Context, *WriteRequest, *WriteResponse) error
	Delete(context.Context, *DeleteRequest, *DeleteResponse) error
	List(context.Context, *ListRequest, Store_ListStream) error
	Databases(context.Context, *DatabasesRequest, *DatabasesResponse) error
	Tables(context.Context, *TablesRequest, *TablesResponse) error
}

func RegisterStoreHandler(s server.Server, hdlr StoreHandler, opts ...server.HandlerOption) error {
	type storeImpl interface {
		Read(ctx context.Context, in *ReadRequest, out *ReadResponse) error
		Write(ctx context.Context, in *WriteRequest, out *WriteResponse) error
		Delete(ctx context.Context, in *DeleteRequest, out *DeleteResponse) error
		List(ctx context.Context, stream server.Stream) error
		Databases(ctx context.Context, in *DatabasesRequest, out *DatabasesResponse) error
		Tables(ctx context.Context, in *TablesRequest, out *TablesResponse) error
	}
	type Store struct {
		storeImpl
	}
	h := &storeHandler{hdlr}
	opts = append(opts, server.OpenAPIHandler(NewStoreOpenAPI()))
	return s.Handle(s.NewHandler(&Store{h}, opts...))
}

type storeHandler struct {
	StoreHandler
}

func (h *storeHandler) Read(ctx context.Context, in *ReadRequest, out *ReadResponse) error {
	return h.StoreHandler.Read(ctx, in, out)
}

func (h *storeHandler) Write(ctx context.Context, in *WriteRequest, out *WriteResponse) error {
	return h.StoreHandler.Write(ctx, in, out)
}

func (h *storeHandler) Delete(ctx context.Context, in *DeleteRequest, out *DeleteResponse) error {
	return h.StoreHandler.Delete(ctx, in, out)
}

func (h *storeHandler) List(ctx context.Context, stream server.Stream) error {
	m := new(ListRequest)
	if err := stream.Recv(m); err != nil {
		return err
	}
	return h.StoreHandler.List(ctx, m, &storeListStream{stream})
}

type Store_ListStream interface {
	Context() context.Context
	SendMsg(interface{}) error
	RecvMsg(interface{}) error
	Close() error
	Send(*ListResponse) error
}

type storeListStream struct {
	stream server.Stream
}

func (x *storeListStream) Close() error {
	return x.stream.Close()
}

func (x *storeListStream) Context() context.Context {
	return x.stream.Context()
}

func (x *storeListStream) SendMsg(m interface{}) error {
	return x.stream.Send(m)
}

func (x *storeListStream) RecvMsg(m interface{}) error {
	return x.stream.Recv(m)
}

func (x *storeListStream) Send(m *ListResponse) error {
	return x.stream.Send(m)
}

func (h *storeHandler) Databases(ctx context.Context, in *DatabasesRequest, out *DatabasesResponse) error {
	return h.StoreHandler.Databases(ctx, in, out)
}

func (h *storeHandler) Tables(ctx context.Context, in *TablesRequest, out *TablesResponse) error {
	return h.StoreHandler.Tables(ctx, in, out)
}
