package openapi

import (
	"context"
	"strings"
	"sync"

	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/core/registry"
	pb "github.com/vine-io/vine/lib/api/handler/openapi/proto"
	log "github.com/vine-io/vine/lib/logger"
	maddr "github.com/vine-io/vine/util/addr"
)

var once = sync.Once{}
var doc = &APIDoc{}

func init() {
	once.Do(func() {
		doc = &APIDoc{
			Doc: &pb.OpenAPI{
				Openapi: "3.0.1",
				Info: &pb.OpenAPIInfo{
					Title:       "Vine Document",
					Description: "OpenAPI3.0",
					Version:     "latest",
				},
				Servers: []*pb.OpenAPIServer{},
				Tags:    []*pb.OpenAPITag{},
				Paths:   map[string]*pb.OpenAPIPath{},
				Components: &pb.OpenAPIComponents{
					SecuritySchemes: &pb.SecuritySchemes{},
					Schemas:         map[string]*pb.Model{},
				},
			},
		}
	})
}

type APIDoc struct {
	sync.RWMutex

	Doc *pb.OpenAPI
}

func (ad *APIDoc) Init(services ...*registry.Service) {
	ctx := context.TODO()
	for _, item := range services {
		list, err := registry.GetService(ctx, item.Name)
		if err != nil {
			continue
		}
		if item.Name == "go.vine.api" {
			for _, node := range item.Nodes {
				if url, ok := node.Metadata["api-address"]; ok {
					if strings.HasPrefix(url, ":") {
						for _, ip := range maddr.IPv4s() {
							if ip == "localhost" || ip == "127.0.0.1" {
								continue
							}
							url = ip + url
						}
					}
					if !strings.HasPrefix(url, "http://") || !strings.HasPrefix(url, "https://") {
						url = "http://" + url
					}
					ad.AddServer(url, item.Name)
				}
			}
		}

		c := client.DefaultClient
		for _, i := range list {
			rsp, e := pb.NewOpenAPIService(i.Name, c).GetOpenAPIDoc(ctx, &pb.GetOpenAPIDocRequest{})
			if e != nil {
				log.Warnf("get %s openapi: %v", i.Name, e)
			}
			if e != nil || len(rsp.Apis) == 0 {
				continue
			}
			for _, api := range rsp.Apis {
				if api == nil || api.Components.SecuritySchemes == nil {
					continue
				}
				ad.AddEndpoint(api)
			}
		}
	}
}

func (ad *APIDoc) AddServer(url, desc string) {
	ad.Lock()
	defer ad.Unlock()

	ad.Doc.Servers = append(ad.Doc.Servers, &pb.OpenAPIServer{
		Url:         url,
		Description: desc,
	})
}

func (ad *APIDoc) AddEndpoint(api *pb.OpenAPI) {
	tags := map[string]*pb.OpenAPITag{}
	ad.RLock()
	for _, tag := range ad.Doc.Tags {
		tags[tag.Name] = tag
	}
	ad.RUnlock()

	ad.Lock()
	defer ad.Unlock()

	for _, tag := range api.Tags {
		if _, ok := tags[tag.Name]; !ok {
			tags[tag.Name] = tag
		}
	}
	ad.Doc.Tags = []*pb.OpenAPITag{}
	for _, tag := range tags {
		ad.Doc.Tags = append(ad.Doc.Tags, tag)
	}
	for name, path := range api.Paths {
		ad.Doc.Paths[name] = path
	}
	for name, schema := range api.Components.Schemas {
		ad.Doc.Components.Schemas[name] = schema
	}
	if api.Components.SecuritySchemes.Basic != nil {
		ad.Doc.Components.SecuritySchemes.Basic = api.Components.SecuritySchemes.Basic
	}
	if api.Components.SecuritySchemes.Bearer != nil {
		ad.Doc.Components.SecuritySchemes.Bearer = api.Components.SecuritySchemes.Bearer
	}
	if api.Components.SecuritySchemes.ApiKeys != nil {
		ad.Doc.Components.SecuritySchemes.ApiKeys = api.Components.SecuritySchemes.ApiKeys
	}
}

func (ad *APIDoc) Out() *pb.OpenAPI {
	out := &pb.OpenAPI{
		Openapi: "3.0.1",
		Info: &pb.OpenAPIInfo{
			Title:       "Vine Document",
			Description: "OpenAPI3.0",
			Version:     "latest",
		},
		Servers: []*pb.OpenAPIServer{},
		Tags:    []*pb.OpenAPITag{},
		Paths:   map[string]*pb.OpenAPIPath{},
		Components: &pb.OpenAPIComponents{
			SecuritySchemes: &pb.SecuritySchemes{},
			Schemas:         map[string]*pb.Model{},
		},
	}
	if ad.Doc == nil {
		return out
	}

	for _, server := range ad.Doc.Servers {
		out.Servers = append(out.Servers, server)
	}
	for _, tag := range ad.Doc.Tags {
		out.Tags = append(out.Tags, tag)
	}
	for name, path := range ad.Doc.Paths {
		out.Paths[name] = path
	}
	if ad.Doc.Components.SecuritySchemes.Basic != nil {
		out.Components.SecuritySchemes.Basic = new(pb.BasicSecurity)
		*out.Components.SecuritySchemes.Basic = *ad.Doc.Components.SecuritySchemes.Basic
	}
	if ad.Doc.Components.SecuritySchemes.ApiKeys != nil {
		out.Components.SecuritySchemes.ApiKeys = new(pb.APIKeysSecurity)
		*out.Components.SecuritySchemes.ApiKeys = *ad.Doc.Components.SecuritySchemes.ApiKeys
	}
	if ad.Doc.Components.SecuritySchemes.Bearer != nil {
		out.Components.SecuritySchemes.Bearer = new(pb.BearerSecurity)
		*out.Components.SecuritySchemes.Bearer = *ad.Doc.Components.SecuritySchemes.Bearer
	}
	for name, model := range ad.Doc.Components.Schemas {
		out.Components.Schemas[name] = model
	}

	return out
}
