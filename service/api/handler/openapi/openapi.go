// MIT License
//
// Copyright (c) 2020 Lack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package openapi

import (
	"html/template"
	"net/http"
	"strings"

	json "github.com/json-iterator/go"

	"github.com/lack-io/vine"
	openapipb "github.com/lack-io/vine/proto/apis/openapi"
	maddr "github.com/lack-io/vine/util/addr"
)

type openAPI struct {
	svc    vine.Service
	prefix string
}

func (o *openAPI) OpenAPIHandler(w http.ResponseWriter, r *http.Request) {
	if ct := r.Header.Get("Content-Type"); ct == "application/json" {
		w.Header().Set("Content-Type", ct)
		//w.Write(b)
		return
	}
	var tmpl string
	style := r.URL.Query().Get("style")
	switch style {
	case "redoc":
		tmpl = redocTmpl
	default:
		tmpl = swaggerTmpl
	}

	render(w, r, tmpl, nil)
}

func (o *openAPI) OpenAPIJOSNHandler(w http.ResponseWriter, r *http.Request) {
	services, err := o.svc.Options().Registry.ListServices()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	tags := make(map[string]*openapipb.OpenAPITag, 0)
	paths := make(map[string]*openapipb.OpenAPIPath, 0)
	schemas := make(map[string]*openapipb.Model, 0)
	security := &openapipb.SecuritySchemes{}
	servers := make([]*openapipb.OpenAPIServer, 0)
	for _, item := range services {
		list, err := o.svc.Options().Registry.GetService(item.Name)
		if err != nil {
			continue
		}
		if item.Name == "go.vine.api" {
			for _, node := range item.Nodes {
				if v, ok := node.Metadata["api-address"]; ok {
					if strings.HasPrefix(v, ":") {
						for _, ip := range maddr.IPv4s() {
							if ip == "localhost" || ip == "127.0.0.1" {
								continue
							}
							v = ip + v
						}
					}
					if !strings.HasPrefix(v, "http://") || !strings.HasPrefix(v, "https://") {
						v = "http://" + v
					}
					servers = append(servers, &openapipb.OpenAPIServer{
						Url:         v,
						Description: item.Name,
					})
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
				for _, tag := range api.Tags {
					tags[tag.Name] = tag
				}
				for name, path := range api.Paths {
					paths[name] = path
				}
				for name, schema := range api.Components.Schemas {
					schemas[name] = schema
				}
				if api.Components.SecuritySchemes.Basic != nil {
					security.Basic = api.Components.SecuritySchemes.Basic
				}
				if api.Components.SecuritySchemes.Bearer != nil {
					security.Bearer = api.Components.SecuritySchemes.Bearer
				}
				if api.Components.SecuritySchemes.ApiKeys != nil {
					security.ApiKeys = api.Components.SecuritySchemes.ApiKeys
				}
			}
		}
	}
	openapi := &openapipb.OpenAPI{
		Openapi: "3.0.1",
		Info: &openapipb.OpenAPIInfo{
			Title:       "Vine Document",
			Description: "OpenAPI3.0",
		},
		Tags:    []*openapipb.OpenAPITag{},
		Paths:   paths,
		Servers: servers,
		Components: &openapipb.OpenAPIComponents{
			SecuritySchemes: security,
			Schemas:         schemas,
		},
	}
	for _, tag := range tags {
		openapi.Tags = append(openapi.Tags, tag)
	}
	v, _ := json.Marshal(openapi)
	w.Write(v)
	w.WriteHeader(200)
}

func (o *openAPI) ServeHTTP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

func New(svc vine.Service) *openAPI {
	return &openAPI{svc: svc}
}

func render(w http.ResponseWriter, r *http.Request, tmpl string, data interface{}) {
	t, err := template.New("template").Funcs(template.FuncMap{
		//		"format": format,
	}).Parse(layoutTemplate)
	if err != nil {
		http.Error(w, "Error occurred:"+err.Error(), 500)
		return
	}
	t, err = t.Parse(tmpl)
	if err != nil {
		http.Error(w, "Error occurred:"+err.Error(), 500)
		return
	}
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Error occurred:"+err.Error(), 500)
	}
}
