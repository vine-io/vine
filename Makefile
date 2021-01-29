NAME=vine
IMAGE_NAME=lack-io/$(NAME)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_TAG=$(shell git describe --abbrev=0 --tags --always --match "v*")
GIT_IMPORT=github.com/lack-io/vine/cmd/vine/service
CGO_ENABLED=0
BUILD_DATE=$(shell date +%s)
LDFLAGS=-X $(GIT_IMPORT).GitCommit=$(GIT_COMMIT) -X $(GIT_IMPORT).GitTag=$(GIT_TAG) -X $(GIT_IMPORT).BuildDate=$(BUILD_DATE)
IMAGE_TAG=$(GIT_TAG)-$(GIT_COMMIT)
ROOT=github.com/lack-io/vine

all: build

vendor:
	go mod vendor

build:
	go build -a -installsuffix cgo -ldflags "-s -w ${LDFLAGS}" -o $(NAME) cmd/vine/main.go

install:
	go get github.com/gogo/protobuf
	go get github.com/lack-io/vine/cmd/protoc-gen-gogofaster
	go get github.com/lack-io/vine/cmd/protoc-gen-vine
	go get github.com/lack-io/vine/cmd/protoc-gen-validator

protoc:
	cd $(GOPATH)/src && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. ${ROOT}/proto/apis/registry/registry.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. ${ROOT}/proto/apis/errors/errors.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. ${ROOT}/proto/apis/usage/usage.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/api/api.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/api/auth/auth.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/auth/auth.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/bot/bot.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/broker/broker.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/client/client.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/config/config.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/debug/debug.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/debug/log/log.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/debug/stats/stats.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/debug/trace/trace.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/file/file.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/router/router.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/network/dns/dns.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/network/network.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/registry/registry.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/runtime/runtime.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/server/server.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/store/store.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogofaster_out=plugins=grpc:. --vine_out=. ${ROOT}/proto/services/transport/transport.proto


	sed -i "" "s/ref,omitempty/$$\ref,omitempty/g" proto/apis/registry/registry.pb.go
	sed -i "" "s/applicationJson,omitempty/application\/json,omitempty/g" proto/apis/registry/registry.pb.go
	sed -i "" "s/applicationXml,omitempty/application\/xml,omitempty/g" proto/apis/registry/registry.pb.go

docker:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):latest
	docker push $(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(IMAGE_NAME):latest

vet:
	go vet ./...

test: vet
	go test -v ./...

clean:
	rm -rf ./vine

.PHONY: build clean vet test docker install protoc
