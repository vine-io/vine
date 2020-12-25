package template

var (
	Makefile = `
GOPATH:=$(shell go env GOPATH)
{{if ne .Type "web"}}
.PHONY: proto
proto:
    {{if eq .UseGoPath true}}
	protoc --proto_path=${GOPATH}/src:. --vine_out=${MODIFY}:. --gogofaster_out=${MODIFY}. proto/{{.Alias}}/{{.Alias}}.proto
    {{end}}

.PHONY: build
build: proto
{{else}}
.PHONY: build
build:
{{end}}
	go build -o {{.Alias}}-{{.Type}} *.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t {{.Alias}}-{{.Type}}:latest
`

	GenerateFile = `package main
//go:generate make proto
`
)
