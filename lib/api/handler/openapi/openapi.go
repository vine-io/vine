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

	"github.com/gin-gonic/gin"
	"github.com/rakyll/statik/fs"
	"github.com/vine-io/vine/core/registry"
	_ "github.com/vine-io/vine/lib/api/handler/openapi/statik"
	log "github.com/vine-io/vine/lib/logger"
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
	services, err := registry.ListServices(ctx)
	if err != nil {
		ctx.JSON(500, err.Error())
		return
	}
	doc.Init(services...)
	ctx.JSON(200, doc.Out())
}

func openAPIServiceHandler(ctx *gin.Context) {
	services, err := registry.ListServices(ctx)
	if err != nil {
		ctx.JSON(500, err.Error())
		return
	}
	out := make([]*registry.Service, 0)
	for _, item := range services {
		list, _ := registry.GetService(ctx, item.Name)
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
