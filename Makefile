# Copyright 2021 NetApp, Inc.  All Rights Reserved
.DEFAULT_GOAL:=help

.PHONY: help deps clean build test fmt vet package

###############################################################################
# Anything that needs to be done before we build everything
#  Check for GCC, GO version, etc and anything else we are dependent on.
###############################################################################
GCC_EXISTS := $(shell which gcc)
REQUIRED_GO_VERSION := 1.15
ifneq (, $(shell which go))
FOUND_GO_VERSION := $(shell go version | cut -d" " -f3 | cut -d"o" -f 2)
CORRECT_GO_VERSION := $(shell expr `go version | cut -d" " -f3 | cut -d"o" -f 2` \>= ${REQUIRED_GO_VERSION})
endif
RELEASE      ?= $(shell git describe --tags --abbrev=0)
VERSION      ?= $(shell expr `date +%Y.%m.%d%H | cut -c 3-`)
COMMIT       := $(shell git rev-parse --short HEAD)
BUILD_DATE   := `date +%FT%T%z`
LD_FLAGS     := "-X 'goharvest2/cmd/harvest/version.VERSION=$(VERSION)' -X 'goharvest2/cmd/harvest/version.Release=$(RELEASE)' -X 'goharvest2/cmd/harvest/version.Commit=$(COMMIT)' -X 'goharvest2/cmd/harvest/version.BuildDate=$(BUILD_DATE)'"
GOARCH ?= amd64
GOOS ?= linux
HARVEST_PACKAGE := harvest-${VERSION}-${RELEASE}_${GOOS}_${GOARCH}
DIST := dist
TMP := /tmp/${HARVEST_PACKAGE}

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

header:
	@echo "    _  _                     _     ___   __   "
	@echo "   | || |__ _ _ ___ _____ __| |_  |_  ) /  \  "
	@echo "   | __ / _\` | '_\ V / -_|_-<  _|  / / | () | "
	@echo "   |_||_\__,_|_|  \_/\___/__/\__| /___(_)__/  "
	@echo

deps: header ## Check dependencies
	@echo "checking Harvest dependencies"
ifeq (${GCC_EXISTS}, )
	@echo
	@echo "Harvest requires that you have gcc installed."
	@echo
	@exit 1
endif
	@# Make sure that go exists
ifeq (${FOUND_GO_VERSION}, )
	@echo
	@echo "Harvest requires that the go lang is installed and is at least version: ${REQUIRED_GO_VERSION}"
	@echo
	@exit 1
endif
	@# Check to make sure that GO is the correct version
ifeq ("${CORRECT_GO_VERSION}", "0")
	@echo
	@echo "Required go lang version is ${REQUIRED_GO_VERSION}, but found ${FOUND_GO_VERSION}"
	@echo
	@exit 1
endif

clean: header ## Cleanup the project binary (bin) folders
	@echo "Cleaning harvest files"
	@rm -rf bin

test: ## run tests
	@echo "Running tests"
	go test -v ./...

fmt: ## format the go source files
	@echo "Running gofmt"
	go fmt ./...

vet: ## run go vet on the source files
	@echo "Running go vet"
	go vet ./...

build: clean deps fmt harvest ## Build the project

package: clean deps build test dist-tar ## Package Harvest binary

all: package ## Build, Test, Package

harvest: deps
	@# Build the harvest cli
	@echo "Building harvest"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -o bin/harvest -ldflags=$(LD_FLAGS) cmd/harvest/harvest.go

	@# Build the harvest poller
	@echo "Building poller"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -o bin/poller -ldflags=$(LD_FLAGS) cmd/poller/poller.go

	@# Build the daemonize for the pollers
	@echo "Building daemonize"
	@cd cmd/tools/daemonize; gcc daemonize.c -o ../../../bin/daemonize

	@# Build the zapi tool
	@echo "Building zapi tool"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -o bin/zapi -ldflags=$(LD_FLAGS) cmd/tools/zapi/main/main.go

	@# Build the grafana tool
	@echo "Building grafana tool"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -o bin/grafana -ldflags=$(LD_FLAGS) cmd/tools/grafana/main/main.go

	@# Build the rest tool
	@echo "Building rest tool"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -o bin/rest -ldflags=$(LD_FLAGS) cmd/tools/rest/main/main.go

###############################################################################
# Build tar gz distribution
###############################################################################
dist-tar:
	@echo "Building tar"
	@rm -rf ${TMP}
	@rm -rf ${DIST}
	@mkdir ${TMP}
	@mkdir ${DIST}
	@cp -r .git cmd bin conf docker docs grafana pkg service go.mod go.sum Makefile README.md LICENSE ${TMP}
	@cp harvest.yml ${TMP}/harvest.yml
	@tar --directory /tmp --create --gzip --file ${DIST}/${HARVEST_PACKAGE}.tar.gz ${HARVEST_PACKAGE}
	@rm -rf ${TMP}
	@echo "tar artifact @" ${DIST}/${HARVEST_PACKAGE}.tar.gz
