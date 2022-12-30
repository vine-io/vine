NAME=vine
IMAGE_NAME=vine-io/$(NAME)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_TAG=$(shell git describe --abbrev=0 --tags --always --match "v*")
GIT_VERSION=github.com/vine-io/vine/cmd/vine/version
CGO_ENABLED=0
BUILD_DATE=$(shell date +%s)
LDFLAGS=-X $(GIT_VERSION).GitCommit=$(GIT_COMMIT) -X $(GIT_VERSION).GitTag=$(GIT_TAG) -X $(GIT_VERSION).BuildDate=$(BUILD_DATE)
IMAGE_TAG=$(GIT_TAG)-$(GIT_COMMIT)
ROOT=github.com/vine-io/vine
TOOLS=$(shell echo "vine protoc-gen-gogo protoc-gen-vine protoc-gen-dao protoc-gen-validator protoc-gen-deepcopy protoc-gen-cli" )

all: build

vendor:
	go mod vendor

test-coverage:
	go test ./... -bench=. -coverage

lint:
	golint -set_exit_status ./..

release:
ifeq "$(TAG)" ""
	@echo "missing tag"
	exit 1
endif
	git tag $(TAG)
	make build-tag
	git add .
	git commit -m "$(TAG)"
	git tag -d $(TAG)
	git tag $(TAG)

build-tag:
	sed -i "" "s/GitCommit = ".*"/GitCommit = \"$(GIT_COMMIT)\"/g" cmd/vine/version/version.go
	sed -i "" "s/GitTag    = ".*"/GitTag    = \"$(GIT_TAG)\"/g" cmd/vine/version/version.go
	sed -i "" "s/BuildDate = ".*"/BuildDate = \"$(BUILD_DATE)\"/g" cmd/vine/version/version.go

tar-windows:
	mkdir -p _output/windows-amd64
	for i in $(TOOLS); do \
	    GOOS=windows GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-s -w ${LDFLAGS}" -o _output/windows-amd64/$$i.exe cmd/$$i/main.go ;\
	done && \
	cd _output && rm -fr $(NAME)-windows-amd64-$(GIT_TAG).zip && zip $(NAME)-windows-amd64-$(GIT_TAG).zip windows-amd64/* && rm -fr windows-amd64

tar-linux-amd64:
	mkdir -p _output/linux-amd64
	for i in $(TOOLS); do \
	    GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-s -w ${LDFLAGS}" -o _output/linux-amd64/$$i cmd/$$i/main.go ;\
	done && \
	cd _output && rm -fr $(NAME)-linux-amd64-$(GIT_TAG).tar.gz && tar -zcvf $(NAME)-linux-amd64-$(GIT_TAG).tar.gz linux-amd64/* && rm -fr linux-amd64

tar-linux-arm64:
	mkdir -p _output/linux-arm64
	for i in $(TOOLS); do \
	    GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -ldflags "-s -w ${LDFLAGS}" -o _output/linux-arm64/$$i cmd/$$i/main.go ;\
	done && \
	cd _output && rm -fr $(NAME)-linux-arm64-$(GIT_TAG).tar.gz && tar -zcvf $(NAME)-linux-arm64-$(GIT_TAG).tar.gz linux-arm64/* && rm -fr linux-arm64

tar-darwin-amd64:
	mkdir -p _output/darwin-amd64
	for i in $(TOOLS); do \
	    GOOS=darwin GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-s -w ${LDFLAGS}" -o _output/darwin-amd64/$$i cmd/$$i/main.go ;\
	done && \
	cd _output && rm -fr $(NAME)-darwin-amd64-$(GIT_TAG).tar.gz && tar -zcvf $(NAME)-darwin-amd64-$(GIT_TAG).tar.gz darwin-amd64/* && rm -fr darwin-amd64

tar-darwin-arm64:
	mkdir -p _output/darwin-arm64
	for i in $(TOOLS); do \
	    GOOS=darwin GOARCH=arm64 go build -a -installsuffix cgo -ldflags "-s -w ${LDFLAGS}" -o _output/darwin-arm64/$$i cmd/$$i/main.go ;\
	done && \
	cd _output && rm -fr $(NAME)-darwin-arm64-$(GIT_TAG).tar.gz && tar -zcvf $(NAME)-darwin-arm64-$(GIT_TAG).tar.gz darwin-arm64/* && rm -fr darwin-arm64

tar: tar-windows tar-linux-amd64 tar-linux-arm64 tar-darwin-amd64 tar-darwin-arm64

build:
	go build -a -installsuffix cgo -ldflags "-s -w $(LDFLAGS)" -o $(NAME) cmd/vine/main.go

build-windows-tool:
	mkdir -p _output
	GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-gogo.exe cmd/protoc-gen-gogo/main.go
	GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-vine.exe cmd/protoc-gen-vine/main.go
	GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-dao.exe cmd/protoc-gen-dao/main.go
	GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-validator.exe cmd/protoc-gen-validator/main.go
	GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-deepcopy.exe cmd/protoc-gen-deepcopy/main.go
	GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-cli.exe cmd/protoc-gen-cli/main.go
	GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w ${LDFLAGS}" -o _output/vine.exe cmd/vine/main.go

build-linux-tool:
	mkdir -p _output
	GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-gogo cmd/protoc-gen-gogo/main.go
	GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-vine cmd/protoc-gen-vine/main.go
	GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-dao cmd/protoc-gen-dao/main.go
	GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-validator cmd/protoc-gen-validator/main.go
	GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-deepcopy cmd/protoc-gen-deepcopy/main.go
	GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o _output/protoc-gen-cli cmd/protoc-gen-cli/main.go
	GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w ${LDFLAGS}" -o _output/vine.exe cmd/vine/main.go

install:
	go get github.com/rakyll/statik
	go get github.com/gogo/protobuf
	go get github.com/vine-io/vine/cmd/vine
	go get github.com/vine-io/vine/cmd/protoc-gen-gogo
	go get github.com/vine-io/vine/cmd/protoc-gen-vine
	go get github.com/vine-io/vine/cmd/protoc-gen-validator
	go get github.com/vine-io/vine/cmd/protoc-gen-deepcopy
	go get github.com/vine-io/vine/cmd/protoc-gen-dao
	go get github.com/vine-io/vine/cmd/protoc-gen-cli

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

.PHONY: build build-tag release tar-windows tar-linux-arm64 tar-darwin-amd64 tar-darwin-arm64 tar-darwin-amd64 tar build-windows-tool clean vet test docker install protoc openapi
