package template

var (
	SingleMakefile = `
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_TAG=$(shell git describe --abbrev=0 --tags --always --match "v*")
GIT_VERSION={{.Dir}}/pkg/version/version
CGO_ENABLED=0
BUILD_DATE=$(shell date +%s)
LDFLAGS=-X $(GIT_VERSION).GitCommit=$(GIT_COMMIT) -X $(GIT_VERSION).GitTag=$(GIT_TAG) -X $(GIT_VERSION).BuildDate=$(BUILD_DATE)

build-tag:
	sed -i "" "s/GitTag     = ".*"/GitTag     = \"$(GIT_TAG)\"/g" pkg/version/version.go
	sed -i "" "s/GitCommit  = ".*"/GitCommit  = \"$(GIT_COMMIT)\"/g" pkg/version/version.go
	sed -i "" "s/BuildDate  = ".*"/BuildDate  = \"$(BUILD_DATE)\"/g" pkg/version/version.go

install:

build:

clean:

.PHONY: build-tag install build clean
`

	ClusterMakefile = `
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_TAG=$(shell git describe --abbrev=0 --tags --always --match "v*")
GIT_VERSION={{.Dir}}/pkg/version/version
CGO_ENABLED=0
BUILD_DATE=$(shell date +%s)
LDFLAGS=-X $(GIT_VERSION).GitCommit=$(GIT_COMMIT) -X $(GIT_VERSION).GitTag=$(GIT_TAG) -X $(GIT_VERSION).BuildDate=$(BUILD_DATE)

build-tag:
	sed -i "" "s/GitTag     = ".*"/GitTag     = \"$(GIT_TAG)\"/g" pkg/version/version.go
	sed -i "" "s/GitCommit  = ".*"/GitCommit  = \"$(GIT_COMMIT)\"/g" pkg/version/version.go
	sed -i "" "s/BuildDate  = ".*"/BuildDate  = \"$(BUILD_DATE)\"/g" pkg/version/version.go

install:

build:

clean:

.PHONY: build-tag install build clean
`

	GenerateFile = `package main
//go:generate make proto
`
)
