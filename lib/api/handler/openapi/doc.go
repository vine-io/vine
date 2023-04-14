package openapi

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/core/registry"
	pb "github.com/vine-io/vine/lib/api/handler/openapi/proto"
	log "github.com/vine-io/vine/lib/logger"
	maddr "github.com/vine-io/vine/util/addr"
)

var once = sync.Once{}
var doc = &docStore{}

func init() {

	oa := &pb.OpenAPI{
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

	once.Do(func() {
		doc = newDocStore(oa)
	})
}

type docStore struct {
	sync.RWMutex

	Doc *pb.OpenAPI
}

func newDocStore(api *pb.OpenAPI) *docStore {
	return &docStore{Doc: api}
}

func (s *docStore) discovery(name string, co client.Client, reg registry.Registry, services []*registry.Service) {
	ctx := context.TODO()
	for _, item := range services {
		list, err := reg.GetService(ctx, item.Name)
		if err != nil {
			continue
		}

		for _, i := range list {
			if i.Name == name {
				for _, node := range i.Nodes {
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
						s.AddServer(url, i.Name)
					}
				}
			}

			rsp, e := pb.NewOpenAPIService(i.Name, co).GetOpenAPIDoc(ctx, &pb.GetOpenAPIDocRequest{})
			if e != nil && i.Name != "go.vine.api" {
				log.Warnf("get %s openapi: %v", i.Name, e)
			}
			if e != nil || len(rsp.Apis) == 0 {
				continue
			}
			for _, api := range rsp.Apis {
				if api == nil || api.Components.SecuritySchemes == nil {
					continue
				}
				s.AddEndpoint(api)
			}
		}
	}
}

func (s *docStore) AddServer(url, desc string) {
	s.Lock()
	defer s.Unlock()

	for _, item := range s.Doc.Servers {
		if item.Url == url {
			return
		}
	}

	s.Doc.Servers = append(s.Doc.Servers, &pb.OpenAPIServer{
		Url:         url,
		Description: desc,
	})
}

func (s *docStore) AddEndpoint(api *pb.OpenAPI) {
	tags := map[string]*pb.OpenAPITag{}
	s.RLock()
	for _, tag := range s.Doc.Tags {
		tags[tag.Name] = tag
	}
	s.RUnlock()

	s.Lock()
	defer s.Unlock()

	for _, tag := range api.Tags {
		if _, ok := tags[tag.Name]; !ok {
			tags[tag.Name] = tag
		}
	}
	s.Doc.Tags = []*pb.OpenAPITag{}
	for _, tag := range tags {
		s.Doc.Tags = append(s.Doc.Tags, tag)
	}
	for name, path := range api.Paths {
		s.Doc.Paths[name] = path
	}
	for name, schema := range api.Components.Schemas {
		s.Doc.Components.Schemas[name] = schema
	}
	if api.Components.SecuritySchemes.Basic != nil {
		s.Doc.Components.SecuritySchemes.Basic = api.Components.SecuritySchemes.Basic
	}
	if api.Components.SecuritySchemes.Bearer != nil {
		s.Doc.Components.SecuritySchemes.Bearer = api.Components.SecuritySchemes.Bearer
	}
	if api.Components.SecuritySchemes.ApiKeys != nil {
		s.Doc.Components.SecuritySchemes.ApiKeys = api.Components.SecuritySchemes.ApiKeys
	}
}

func (s *docStore) output() *pb.OpenAPI {
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
	if s.Doc == nil {
		return out
	}

	for _, server := range s.Doc.Servers {
		out.Servers = append(out.Servers, server)
	}
	for _, tag := range s.Doc.Tags {
		out.Tags = append(out.Tags, tag)
	}
	sort.Slice(out.Tags, func(i, j int) bool {
		return out.Tags[i].Name < out.Tags[j].Name
	})
	for name, path := range s.Doc.Paths {
		out.Paths[name] = path
	}
	if s.Doc.Components.SecuritySchemes.Basic != nil {
		out.Components.SecuritySchemes.Basic = new(pb.BasicSecurity)
		*out.Components.SecuritySchemes.Basic = *s.Doc.Components.SecuritySchemes.Basic
	}
	if s.Doc.Components.SecuritySchemes.ApiKeys != nil {
		out.Components.SecuritySchemes.ApiKeys = new(pb.APIKeysSecurity)
		*out.Components.SecuritySchemes.ApiKeys = *s.Doc.Components.SecuritySchemes.ApiKeys
	}
	if s.Doc.Components.SecuritySchemes.Bearer != nil {
		out.Components.SecuritySchemes.Bearer = new(pb.BearerSecurity)
		*out.Components.SecuritySchemes.Bearer = *s.Doc.Components.SecuritySchemes.Bearer
	}
	for name, model := range s.Doc.Components.Schemas {
		out.Components.Schemas[name] = model
	}

	return out
}
