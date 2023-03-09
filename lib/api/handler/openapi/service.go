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
		globalOpenAPI = &openapiService{endpoints: map[string]*api.Endpoint{}}
	})
}

type openapiService struct {
	gosync.RWMutex

	apis      []*pb.OpenAPI
	endpoints map[string]*api.Endpoint
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
	rsp.Endpoints = map[string]*api.Endpoint{ep.Name: ep}
	return nil
}

func RegisterOpenAPIDoc(api *pb.OpenAPI) {
	globalOpenAPI.AppendAPIDoc(api)
}

func InjectEndpoints(endpoints ...api.Endpoint) error {
	for i := range endpoints {
		ep := endpoints[i]
		if v, ok := globalOpenAPI.endpoints[ep.Name]; ok {
			return fmt.Errorf("injects a new endpoint=%s, conflicts with Endpoint{Name:%s, Method:%s}", ep.Name, v.Name, strings.Join(v.Method, ","))
		}
		globalOpenAPI.endpoints[ep.Name] = &ep
	}

	return nil
}

func RegisterOpenAPIHandler(s server.Server, opts ...server.HandlerOption) error {
	return pb.RegisterOpenAPIServiceHandler(s, globalOpenAPI, opts...)
}

func GetAllOpenAPIDoc() []*pb.OpenAPI {
	return globalOpenAPI.GetAPIDoc()
}

func GetEndpoint(name string) (map[string]*api.Endpoint, error) {
	if name == "" {
		return globalOpenAPI.endpoints, nil
	}

	ep, ok := globalOpenAPI.endpoints[name]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return map[string]*api.Endpoint{ep.Name: ep}, nil
}
