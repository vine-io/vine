package openapi

import (
	gosync "sync"

	"github.com/vine-io/vine"
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
	apis []*pb.OpenAPI
}

func (s *openapiService) GetOpenAPIDoc(ctx *vine.Context, req *pb.GetOpenAPIDocRequest, rsp *pb.GetOpenAPIDocResponse) error {
	rsp.Apis = s.apis
	return nil
}

func RegisterOpenAPIDoc(api *pb.OpenAPI) {
	globalOpenAPI.apis = append(globalOpenAPI.apis, api)
}

func RegisterOpenAPIHandler(s server.Server, opts ...server.HandlerOption) error {
	return pb.RegisterOpenAPIServiceHandler(s, globalOpenAPI, opts...)
}
