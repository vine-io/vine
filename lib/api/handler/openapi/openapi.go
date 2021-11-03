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
	"html/template"
	"mime"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rakyll/statik/fs"
	"github.com/vine-io/vine/core/registry"
	log "github.com/vine-io/vine/lib/logger"
	maddr "github.com/vine-io/vine/util/addr"

	_ "github.com/vine-io/vine/lib/api/handler/openapi/statik"
)

var DefaultPrefix = "/openapi-ui/"

func RegisterOpenAPI(router *gin.Engine) {
	mime.AddExtensionType(".svg", "image/svg+xml")
	statikFs, err := fs.New()
	if err != nil {
		log.Fatalf("Starting OpenAPI: %v", err)
	}
	router.GET(DefaultPrefix, swaggerHandler)
	router.GET(path.Join(DefaultPrefix, "redoc"), redocHandler)
	router.StaticFS(filepath.Join(DefaultPrefix, "static/"), statikFs)
	router.GET("/openapi.json", openAPIJOSNHandler)
	router.GET("/services", openAPIServiceHandler)
	log.Infof("Starting OpenAPI at %v", DefaultPrefix)
}

func swaggerHandler(ctx *gin.Context) {
	if ct := ctx.GetHeader("Content-Type"); ct == "application/json" {
		ctx.Request.Header.Set("Content-Type", ct)
		return
	}
	render(ctx, swaggerTmpl, nil)
}

func redocHandler(ctx *gin.Context) {
	if ct := ctx.GetHeader("Content-Type"); ct == "application/json" {
		ctx.Request.Header.Set("Content-Type", ct)
		return
	}
	render(ctx, redocTmpl, nil)
}

func openAPIJOSNHandler(ctx *gin.Context) {
	services, err := registry.ListServices()
	if err != nil {
		ctx.JSON(500, err.Error())
		return
	}
	tags := make(map[string]*registry.OpenAPITag, 0)
	paths := make(map[string]*registry.OpenAPIPath, 0)
	schemas := make(map[string]*registry.Model, 0)
	security := &registry.SecuritySchemes{}
	servers := make([]*registry.OpenAPIServer, 0)
	for _, item := range services {
		list, err := registry.GetService(item.Name)
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
					servers = append(servers, &registry.OpenAPIServer{
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
	openapi := &registry.OpenAPI{
		Openapi: "3.0.1",
		Info: &registry.OpenAPIInfo{
			Title:       "Vine Document",
			Description: "OpenAPI3.0",
			Version:     "latest",
		},
		Tags:    []*registry.OpenAPITag{},
		Paths:   paths,
		Servers: servers,
		Components: &registry.OpenAPIComponents{
			SecuritySchemes: security,
			Schemas:         schemas,
		},
	}
	for _, tag := range tags {
		openapi.Tags = append(openapi.Tags, tag)
	}
	ctx.JSON(200, openapi)
}

func openAPIServiceHandler(ctx *gin.Context) {
	services, err := registry.ListServices()
	if err != nil {
		ctx.JSON(500, err.Error())
		return
	}
	out := make([]*registry.Service, 0)
	for _, item := range services {
		list, _ := registry.GetService(item.Name)
		out = append(out, list...)
	}
	ctx.JSON(200, out)
}

func render(ctx *gin.Context, tmpl string, data interface{}) {
	t, err := template.New("template").Funcs(template.FuncMap{
		//		"format": format,
	}).Parse(layoutTemplate)
	if err != nil {
		ctx.Data(500, "text/html", []byte("Error occurred:"+err.Error()))
		return
	}
	t, err = t.Parse(tmpl)
	if err != nil {
		ctx.Data(500, "text/html", []byte("Error occurred:"+err.Error()))
		return
	}
	buf := bytes.NewBuffer([]byte(""))
	if err := t.ExecuteTemplate(buf, "layout", data); err != nil {
		ctx.Data(500, "text/html", []byte("Error occurred:"+err.Error()))
		return
	}

	ctx.Data(200, "text/html", buf.Bytes())
}
