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

var (
	layoutTemplate = `
{{ define "layout" }}
<!DOCTYPE html>
<html>
<head>
    {{ template "title" . }}
    <!-- needed for adaptive design -->
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    {{ template "style" .}}
</head>
<body>
	{{ template "scripts" . }}
</body>
</html>
{{end}}
{{ define "style" }}{{end}}
{{ define "scripts" }}{{end}}
{{ define "title" }}{{end}}
`

	swaggerTmpl = `
{{ define "title"}}<title>Swagger UI</title>{{end}}
{{ define "style" }}
<link rel="stylesheet" type="text/css" href="./static/swagger/swagger-ui.css" >
    <link rel="icon" type="image/png" href="./static/swagger/favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="./static/swagger/favicon-16x16.png" sizes="16x16" />
    <style>
      html
      {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
      }

      *,
      *:before,
      *:after
      {
        box-sizing: inherit;
      }

      body
      {
        margin:0;
        background: #fafafa;
      }
    </style>
{{end}}
{{ define "scripts" }}
<div id="swagger-ui"></div>
    <script src="./static/swagger/swagger-ui-bundle.js" charset="UTF-8"> </script>
    <script src="./static/swagger/swagger-ui-standalone-preset.js" charset="UTF-8"> </script>
    <script>
    window.onload = function() {
      // Begin Swagger UI call region
      const ui = SwaggerUIBundle({
        url: "/openapi.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      })
      // End Swagger UI call region
      window.ui = ui
    }
  </script>
{{end}}
`

	redocTmpl = `
{{ define "title" }}<title>Redoc</title>{{end}}
{{ define "style" }}
    <link href="./static/redoc/redoc.css" rel="stylesheet">
    <style>
        body {
            margin: 0;
            padding: 0;
        }
    </style>
{{end}}
{{ define "scripts" }}
	<redoc spec-url='/openapi.json'></redoc>
	<script src="./static/redoc/redoc.standalone.js"></script>
{{end}}
`
)
