package template

var (
	Makefile = `
GOPATH:=$(shell go env GOPATH)
.PHONY: init
init:
	go get github.com/gogo/protobuf
	go get github.com/gogo/googleapis
	go get github.com/lack-io/vine/cmd/protoc-gen-gogofaster
	go get github.com/lack-io/vine/cmd/protoc-gen-vine
.PHONY: proto
proto:
	protoc -I=. \
	  -I=($GOPATH)/src \
	  -I=($GOPATH)/src/github.com/gogo/protobuf/protobuf \
	  -I=($GOPATH)/src/github.com/gogo/googleapis \
	  --gogofaster_out=plugins=grpc:. --vine_out=. {{.Alias}}.proto
	
.PHONY: build
build:
	go build -o {{.Alias}} *.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t {{.Alias}}:latest
`

	GenerateFile = `package main
//go:generate make proto
`
)
