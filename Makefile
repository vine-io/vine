NAME=vine
IMAGE_NAME=lack-io/$(NAME)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_TAG=$(shell git describe --abbrev=0 --tags --always --match "v*")
GIT_VERSION=github.com/lack-io/vine/cmd/vine/version
CGO_ENABLED=0
BUILD_DATE=$(shell date +%s)
LDFLAGS=-X $(GIT_VERSION).GitCommit=$(GIT_COMMIT) -X $(GIT_VERSION).GitTag=$(GIT_TAG) -X $(GIT_VERSION).BuildDate=$(BUILD_DATE)
IMAGE_TAG=$(GIT_TAG)-$(GIT_COMMIT)
ROOT=github.com/lack-io/vine

all: build

vendor:
	go mod vendor

build:
	sed -i "" "s/GitCommit = ".*"/GitCommit = \"$(GIT_COMMIT)\"/g" cmd/vine/version/version.go
	sed -i "" "s/GitTag    = ".*"/GitTag    = \"$(GIT_TAG)\"/g" cmd/vine/version/version.go
	sed -i "" "s/BuildDate = ".*"/BuildDate = \"$(BUILD_DATE)\"/g" cmd/vine/version/version.go
	go build -a -installsuffix cgo -ldflags "-s -w ${LDFLAGS}" -o $(NAME) cmd/vine/main.go

build-windows-tool:
	mkdir -p _output
	GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-gogo.exe cmd/protoc-gen-gogo/main.go
	GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-vine.exe cmd/protoc-gen-vine/main.go
	GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-dao.exe cmd/protoc-gen-dao/main.go
	GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-validator.exe cmd/protoc-gen-validator/main.go
	GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-deepcopy.exe cmd/protoc-gen-deepcopy/main.go
	GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w ${LDFLAGS}" -o _output/vine.exe cmd/vine/main.go

install:
	go get github.com/rakyll/statik
	go get github.com/gogo/protobuf
	go get github.com/lack-io/vine/cmd/protoc-gen-gogo
	go get github.com/lack-io/vine/cmd/protoc-gen-vine
	go get github.com/lack-io/vine/cmd/protoc-gen-validator
	go get github.com/lack-io/vine/cmd/protoc-gen-deepcopy
	go get github.com/lack-io/vine/cmd/protoc-gen-dao

protoc:
	cd $(GOPATH)/src && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. ${ROOT}/proto/apis/api/api.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. ${ROOT}/proto/apis/openapi/openapi.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. ${ROOT}/proto/apis/registry/registry.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. ${ROOT}/proto/apis/errors/errors.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. ${ROOT}/proto/apis/usage/usage.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. --vine_out=:. ${ROOT}/proto/services/bot/bot.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. --vine_out=:. ${ROOT}/proto/services/broker/broker.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. --vine_out=:. ${ROOT}/proto/services/config/config.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. --vine_out=:. ${ROOT}/proto/services/debug/debug.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. --vine_out=:. ${ROOT}/proto/services/debug/log/log.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. --vine_out=:. ${ROOT}/proto/services/debug/stats/stats.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. --vine_out=:. ${ROOT}/proto/services/debug/trace/trace.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. --vine_out=:. ${ROOT}/proto/services/file/file.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. --vine_out=:. ${ROOT}/proto/services/router/router.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. --vine_out=:. ${ROOT}/proto/services/network/dns/dns.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. --vine_out=:. ${ROOT}/proto/services/network/network.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. --vine_out=:. ${ROOT}/proto/services/registry/registry.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. --vine_out=:. ${ROOT}/proto/services/runtime/runtime.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. --vine_out=:. ${ROOT}/proto/services/store/store.proto && \
	protoc -I=. -I=$(GOPATH)/src --gogo_out=:. --vine_out=:. ${ROOT}/proto/services/transport/transport.proto


	sed -i "" "s/ref,omitempty/\$$ref,omitempty/g" proto/apis/openapi/openapi.pb.go
	sed -i "" "s/applicationJson,omitempty/application\/json,omitempty/g" proto/apis/openapi/openapi.pb.go
	sed -i "" "s/applicationXml,omitempty/application\/xml,omitempty/g" proto/apis/openapi/openapi.pb.go

openapi:
	statik -m -f -src third_party/OpenAPI/ -dest service/api/handler/openapi

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
	rm -fr ./_output

.PHONY: build build-windows-tool clean vet test docker install protoc openapi
