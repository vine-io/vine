package openapi

import (
	"context"
	"fmt"
	"strings"
	gosync "sync"

	"github.com/vine-io/vine/core/server"
	"github.com/vine-io/vine/lib/api"
	pb "github.com/vine-io/vine/lib/api/handler/openapi/proto"
)

var _ pb.OpenAPIServiceHandler = (*openapiService)(nil)

var (
	globalOpenAPI *openapiService
	openAPIOnce   gosync.Once
)

func init() {
	openAPIOnce.Do(func() {
		globalOpenAPI = &openapiService{endpoints: map[string]*pb.Endpoint{}}
	})
}

type openapiService struct {
	gosync.RWMutex

	apis      []*pb.OpenAPI
	endpoints map[string]*pb.Endpoint
}

func (s *openapiService) AppendAPIDoc(api *pb.OpenAPI) {
	s.Lock()
	defer s.Unlock()

	s.apis = append(s.apis, api)
}

func (s *openapiService) GetAPIDoc() []*pb.OpenAPI {
	s.RLock()
	defer s.RUnlock()

	return s.apis
}

func (s *openapiService) GetOpenAPIDoc(ctx context.Context, req *pb.GetOpenAPIDocRequest, rsp *pb.GetOpenAPIDocResponse) error {
	s.RLock()
	defer s.RUnlock()

	rsp.Apis = s.apis
	return nil
}

func (s *openapiService) GetEndpoint(ctx context.Context, req *pb.GetEndpointRequest, rsp *pb.GetEndpointResponse) error {
	if req.Name == "" {
		rsp.Endpoints = s.endpoints
		return nil
	}

	ep, ok := s.endpoints[req.Name]
	if !ok {
		return fmt.Errorf("not found")
	}
	rsp.Endpoints = map[string]*pb.Endpoint{ep.Name: ep}
	return nil
}

func RegisterOpenAPIDoc(api *pb.OpenAPI) {
	globalOpenAPI.AppendAPIDoc(api)
}

func InjectEndpoints(endpoints ...api.Endpoint) error {
	for _, ep := range endpoints {
		if v, ok := globalOpenAPI.endpoints[ep.Name]; ok {
			return fmt.Errorf("injects a new endpoint=%s, conflicts with Endpoint{Name:%s, Method:%s}", ep.Name, v.Name, strings.Join(v.Method, ","))
		}
		globalOpenAPI.endpoints[ep.Name] = endpointToPb(&ep)
	}

	return nil
}

func RegisterOpenAPIHandler(s server.Server, opts ...server.HandlerOption) error {
	return pb.RegisterOpenAPIServiceHandler(s, globalOpenAPI, opts...)
}

func GetAllOpenAPIDoc() []*pb.OpenAPI {
	return globalOpenAPI.GetAPIDoc()
}

func GetEndpoint(name string) (map[string]*pb.Endpoint, error) {
	if name == "" {
		return globalOpenAPI.endpoints, nil
	}

	ep, ok := globalOpenAPI.endpoints[name]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return map[string]*pb.Endpoint{ep.Name: ep}, nil
}

func endpointToPb(ep *api.Endpoint) *pb.Endpoint {
	pbEp := &pb.Endpoint{
		Name:        ep.Name,
		Description: ep.Description,
		Handler:     ep.Handler,
		Host:        ep.Host,
		Method:      ep.Method,
		Path:        ep.Path,
		Entity:      ep.Entity,
		Security:    ep.Security,
		Body:        ep.Body,
		Stream:      string(ep.Stream),
	}

	return pbEp
}
