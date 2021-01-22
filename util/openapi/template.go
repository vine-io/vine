// Copyright 2021 lack
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
<link rel="stylesheet" type="text/css" href="./swagger-ui.css" >
    <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16" />
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
    <script src="./swagger-ui-bundle.js" charset="UTF-8"> </script>
    <script src="./swagger-ui-standalone-preset.js" charset="UTF-8"> </script>
    <script>
    window.onload = function() {
      // Begin Swagger UI call region
      const ui = SwaggerUIBundle({
        url: "{{ .Url }}",
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
{{ define "title" }}Redoc{{end}}
{{ define "style"}}
    <link href="./redoc.css" rel="stylesheet">
    <style>
        body {
            margin: 0;
            padding: 0;
        }
    </style>
{{end}}
{{ define "scripts" }}
	<redoc spec-url='{{ .Url }}'></redoc>
	<script src="./redoc.standalone.js"></script>
{{end}}
`
)
