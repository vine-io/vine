package openapi

import (
	"strings"
	"sync"

	"github.com/vine-io/vine/core/registry"
	maddr "github.com/vine-io/vine/util/addr"
)

var once = sync.Once{}
var doc = &APIDoc{}

func init() {
	once.Do(func() {
		doc = &APIDoc{
			Doc: &registry.OpenAPI{
				Openapi: "3.0.1",
				Info: &registry.OpenAPIInfo{
					Title:       "Vine Document",
					Description: "OpenAPI3.0",
					Version:     "latest",
				},
				Servers: []*registry.OpenAPIServer{},
				Tags:    []*registry.OpenAPITag{},
				Paths:   map[string]*registry.OpenAPIPath{},
				Components: &registry.OpenAPIComponents{
					SecuritySchemes: &registry.SecuritySchemes{},
					Schemas:         map[string]*registry.Model{},
				},
			},
		}
	})
}

type APIDoc struct {
	once sync.Once
	sync.RWMutex

	Doc *registry.OpenAPI
}

func (ad *APIDoc) Init(services ...*registry.Service) {
	ad.once.Do(func() {
		for _, item := range services {
			list, err := registry.GetService(item.Name)
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

			for _, i := range list {
				if len(i.Apis) == 0 {
					continue
				}
				for _, api := range i.Apis {
					if api == nil || api.Components.SecuritySchemes == nil {
						continue
					}
					ad.AddEndpoint(api)
				}
			}
		}
	})
}

func (ad *APIDoc) AddServer(url, desc string) {
	ad.Lock()
	defer ad.Unlock()

	ad.Doc.Servers = append(ad.Doc.Servers, &registry.OpenAPIServer{
		Url:         url,
		Description: desc,
	})
}

func (ad *APIDoc) AddEndpoint(api *registry.OpenAPI) {
	tags := map[string]*registry.OpenAPITag{}
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
	ad.Doc.Tags = []*registry.OpenAPITag{}
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

func (ad *APIDoc) Out() *registry.OpenAPI {
	out := &registry.OpenAPI{
		Openapi: "3.0.1",
		Info: &registry.OpenAPIInfo{
			Title:       "Vine Document",
			Description: "OpenAPI3.0",
			Version:     "latest",
		},
		Servers: []*registry.OpenAPIServer{},
		Tags:    []*registry.OpenAPITag{},
		Paths:   map[string]*registry.OpenAPIPath{},
		Components: &registry.OpenAPIComponents{
			SecuritySchemes: &registry.SecuritySchemes{},
			Schemas:         map[string]*registry.Model{},
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
		out.Components.SecuritySchemes.Basic = new(registry.BasicSecurity)
		*out.Components.SecuritySchemes.Basic = *ad.Doc.Components.SecuritySchemes.Basic
	}
	if ad.Doc.Components.SecuritySchemes.ApiKeys != nil {
		out.Components.SecuritySchemes.ApiKeys = new(registry.APIKeysSecurity)
		*out.Components.SecuritySchemes.ApiKeys = *ad.Doc.Components.SecuritySchemes.ApiKeys
	}
	if ad.Doc.Components.SecuritySchemes.Bearer != nil {
		out.Components.SecuritySchemes.Bearer = new(registry.BearerSecurity)
		*out.Components.SecuritySchemes.Bearer = *ad.Doc.Components.SecuritySchemes.Bearer
	}
	for name, model := range ad.Doc.Components.Schemas {
		out.Components.Schemas[name] = model
	}

	return out
}
