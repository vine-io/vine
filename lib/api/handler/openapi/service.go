package openapi

import (
	"context"
	gosync "sync"

	"github.com/vine-io/vine/core/server"
	pb "github.com/vine-io/vine/lib/api/handler/openapi/proto"
)

var _ pb.OpenAPIServiceHandler = (*openapiService)(nil)

var (
	globalOpenAPI *openapiService
	openAPIOnce   gosync.Once
)

func init() {
	openAPIOnce.Do(func() {
		globalOpenAPI = &openapiService{}
	})
}

type openapiService struct {
	gosync.RWMutex

	apis []*pb.OpenAPI
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

func RegisterOpenAPIDoc(api *pb.OpenAPI) {
	globalOpenAPI.AppendAPIDoc(api)
}

func RegisterOpenAPIHandler(s server.Server, opts ...server.HandlerOption) error {
	return pb.RegisterOpenAPIServiceHandler(s, globalOpenAPI, opts...)
}

func GetAllOpenAPIDoc() []*pb.OpenAPI {
	return globalOpenAPI.GetAPIDoc()
}
