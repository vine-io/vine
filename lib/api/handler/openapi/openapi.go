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
	"net/http"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/core/registry"
	"github.com/vine-io/vine/third_party"

	log "github.com/vine-io/vine/lib/logger"
)

var (
	DefaultPrefix = "/openapi-ui/"
)

func RegisterOpenAPI(name string, co client.Client, router *gin.Engine) {
	mime.AddExtensionType(".svg", "image/svg+xml")
	router.StaticFS(filepath.Join(DefaultPrefix, "static/"), http.FS(third_party.GetStatic()))

	router.GET(DefaultPrefix, swagger())
	router.GET(path.Join(DefaultPrefix, "redoc"), redoc())
	router.GET("/openapi.json", openAPIJOSN(name, co))
	router.GET("/services", openAPIService(co))
	log.Infof("Starting OpenAPI at %v", DefaultPrefix)
}

func swagger() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if ct := ctx.GetHeader("Content-Type"); ct == "application/json" {
			ctx.Request.Header.Set("Content-Type", ct)
			return
		}
		render(ctx, swaggerTmpl, nil)
	}
}

func redoc() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if ct := ctx.GetHeader("Content-Type"); ct == "application/json" {
			ctx.Request.Header.Set("Content-Type", ct)
			return
		}
		render(ctx, redocTmpl, nil)
	}
}

func openAPIJOSN(name string, co client.Client) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		reg := co.Options().Registry
		services, err := reg.ListServices(ctx)
		if err != nil {
			ctx.JSON(500, err.Error())
			return
		}
		doc.discovery(name, co, reg, services)
		ctx.JSON(200, doc.output())
	}
}

func openAPIService(co client.Client) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		reg := co.Options().Registry
		services, err := reg.ListServices(ctx)
		if err != nil {
			ctx.JSON(500, err.Error())
			return
		}
		out := make([]*registry.Service, 0)
		for _, item := range services {
			list, _ := reg.GetService(ctx, item.Name)
			out = append(out, list...)
		}
		ctx.JSON(200, out)
	}
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
