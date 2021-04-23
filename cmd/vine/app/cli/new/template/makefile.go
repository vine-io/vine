package template

var (
	Makefile = `
GOPATH:=$(shell go env GOPATH)
{{if ne .Type "web"}}
.PHONY: proto
proto:
	cd ${GOPATH}/src && \
{{if eq .UseGoPath true}}	protoc -I=. -I=${GOPATH}/src:. --vine_out=:. --gogo_out=:. {{.Dir}}/proto/apis/apis.proto && \
	protoc -I=. -I=${GOPATH}/src:. --vine_out=:. --gogo_out=:. {{.Dir}}/proto/service/{{.Name}}/{{.Name}}.proto
{{end}}
.PHONY: build
build: proto
{{else}}
.PHONY: build
build:{{end}}	go build -a -installsuffix cgo -ldflags "-s -w" -o {{.Name}} {{.Dir}}/cmd/main.go

.PHONY: install
install:
	go get github.com/gogo/protobuf
	go get github.com/lack-io/vine/cmd/protoc-gen-gogo
	go get github.com/lack-io/vine/cmd/protoc-gen-vine
	go get github.com/lack-io/vine/cmd/protoc-gen-validator
	go get github.com/lack-io/vine/cmd/protoc-gen-deepcopy
	go get github.com/lack-io/vine/cmd/protoc-gen-dao

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t {{.Name}}:latest
`

	GenerateFile = `package main
//go:generate make proto
`
)
