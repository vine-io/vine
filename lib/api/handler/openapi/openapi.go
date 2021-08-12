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
	"bytes"
	"encoding/json"
	"html/template"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/vine-io/vine"
	openapipb "github.com/vine-io/vine/proto/apis/openapi"
	regpb "github.com/vine-io/vine/proto/apis/registry"
	maddr "github.com/vine-io/vine/util/addr"
)

type openAPI struct {
	svc    vine.Service
	prefix string
}

func (o *openAPI) OpenAPIHandler(ctx *fiber.Ctx) error {
	if ct := ctx.Get("Content-Type", ""); ct == "application/json" {
		ctx.Request().Header.Set("Content-Type", ct)
		return nil
	}
	var tmpl string
	style := ctx.Query("style", "")
	switch style {
	case "redoc":
		tmpl = redocTmpl
	default:
		tmpl = swaggerTmpl
	}

	return render(ctx, tmpl, nil)
}

func (o *openAPI) OpenAPIJOSNHandler(ctx *fiber.Ctx) error {
	services, err := o.svc.Options().Registry.ListServices()
	if err != nil {
		return fiber.NewError(500, err.Error())
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
			Version:     "latest",
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
	data, _ := json.MarshalIndent(openapi, "", " ")
	return ctx.Send(data)
}

func (o *openAPI) OpenAPIServiceHandler(ctx *fiber.Ctx) error {
	services, err := o.svc.Options().Registry.ListServices()
	if err != nil {
		return fiber.NewError(500, err.Error())
	}
	out := make([]*regpb.Service, 0)
	for _, item := range services {
		list, _ := o.svc.Options().Registry.GetService(item.Name)
		out = append(out, list...)
	}
	data, _ := json.MarshalIndent(out, "", " ")
	return ctx.Send(data)
}

func New(svc vine.Service) *openAPI {
	return &openAPI{svc: svc}
}

func render(ctx *fiber.Ctx, tmpl string, data interface{}) error {
	t, err := template.New("template").Funcs(template.FuncMap{
		//		"format": format,
	}).Parse(layoutTemplate)
	if err != nil {
		return fiber.NewError(500, "Error occurred:"+err.Error())
	}
	t, err = t.Parse(tmpl)
	if err != nil {
		return fiber.NewError(500, "Error occurred:"+err.Error())
	}
	buf := bytes.NewBuffer([]byte(""))
	if err := t.ExecuteTemplate(buf, "layout", data); err != nil {
		return fiber.NewError(500, "Error occurred:"+err.Error())
	}

	ctx.Set("Content-Type", "text/html")
	return ctx.SendString(buf.String())
}
