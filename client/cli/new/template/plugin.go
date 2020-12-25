package template

var (
	Plugin = `package main
{{if .Plugins}}
import ({{range .Plugins}}
	_ "github.com/lack-io/plugins/{{.}}"{{end}}
){{end}}
`
)
