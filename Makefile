# Copyright 2021 NetApp, Inc.  All Rights Reserved
.DEFAULT_GOAL:=help

.PHONY: help deps clean build test fmt vet package asup dev fetch-asup

###############################################################################
#  Check for GCC, GO version, etc and anything we are dependent on.
###############################################################################
SHELL := /bin/bash
GCC_EXISTS := $(shell which gcc)
REQUIRED_GO_VERSION := 1.19
ifneq (, $(shell which go))
FOUND_GO_VERSION := $(shell go version | cut -d" " -f3 | cut -d"o" -f 2)
CORRECT_GO_VERSION := $(shell expr `go version | cut -d" " -f3 | cut -d"o" -f 2` \>= ${REQUIRED_GO_VERSION})
endif
TAG_COMMIT   ?= $(shell git rev-list --tags --max-count=1)
RELEASE      ?= $(shell git describe --tags $(TAG_COMMIT))
VERSION      ?= $(shell expr `date +%Y.%m.%d%H | cut -c 3-`)
COMMIT       := $(shell git rev-parse --short HEAD)
BUILD_DATE   := `date +%FT%T%z`
LD_FLAGS     := "-X 'github.com/netapp/harvest/v2/cmd/harvest/version.VERSION=$(VERSION)' -X 'github.com/netapp/harvest/v2/cmd/harvest/version.Release=$(RELEASE)' -X 'github.com/netapp/harvest/v2/cmd/harvest/version.Commit=$(COMMIT)' -X 'github.com/netapp/harvest/v2/cmd/harvest/version.BuildDate=$(BUILD_DATE)'"
GOARCH ?= amd64
GOOS ?= linux
HARVEST_PACKAGE := harvest-${VERSION}-${RELEASE}_${GOOS}_${GOARCH}
DIST := dist
TMP := /tmp/${HARVEST_PACKAGE}
ASUP_TMP := /tmp/asup
ASUP_MAKE_TARGET ?= build #one of build/production
GIT_TOKEN ?=
CURRENT_DIR = $(shell pwd)
ASUP_BIN = asup
ASUP_BIN_VERSION ?= main #change it to match tag of release branch
BIN_PLATFORM ?= linux
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
LINT_EXISTS := $(shell which golangci-lint)


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
	@echo "Harvest requires that Go is installed and at least version: ${REQUIRED_GO_VERSION}"
	@echo
	@exit 1
endif
	@# Check to make sure that GO is the correct version
ifeq ("${CORRECT_GO_VERSION}", "0")
	@echo
	@echo "Required Go version is ${REQUIRED_GO_VERSION}, but found ${FOUND_GO_VERSION}"
	@echo
	@exit 1
endif

clean: header ## Cleanup the project binary (bin) folders
	@echo "Cleaning harvest files"
	@if test -d bin; then ls -d ./bin/* | grep -v "asup" | xargs rm -f; fi

test: ## run tests
	@echo "Running tests"
	go test ./...

fmt: ## format the go source files
	@echo "Running gofmt"
	go fmt ./...

vet: ## run go vet on the source files
	@echo "Running go vet"
	go vet ./...

lint: ## run golangci-lint on the source files
ifeq (${LINT_EXISTS}, )
	@echo
	@echo "Lint task requires that you have https://golangci-lint.run/ installed."
	@echo
	@exit 1
endif
	golangci-lint run

build: clean deps fmt harvest fetch-asup ## Build the project

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
	@cp -r .git cmd bin conf docker docs grafana pkg service cert autosupport go.mod go.sum Makefile README.md LICENSE prom-stack.tmpl harvest.cue ${TMP}
	@cp harvest.yml ${TMP}/harvest.yml
	@tar --directory /tmp --create --gzip --file ${DIST}/${HARVEST_PACKAGE}.tar.gz ${HARVEST_PACKAGE}
	@rm -rf ${TMP}
	@echo "tar artifact @" ${DIST}/${HARVEST_PACKAGE}.tar.gz

asup:
	@echo "Building AutoSupport"
	@rm -rf autosupport/asup
	@rm -rf ${ASUP_TMP}
	@mkdir ${ASUP_TMP}
	# check if there is an equivalent branch name to harvest. If branch name is not found then take autosupport code from main branch.
	@if [[ $(shell git ls-remote --heads  https://${GIT_TOKEN}@github.com/NetApp/harvest-private.git ${BRANCH} | wc -l | xargs) == 0 ]]; then\
		git clone -b main https://${GIT_TOKEN}@github.com/NetApp/harvest-private.git ${ASUP_TMP};\
	else\
		git clone -b ${BRANCH} https://${GIT_TOKEN}@github.com/NetApp/harvest-private.git ${ASUP_TMP};\
	fi
	@cd ${ASUP_TMP}/harvest-asup && make ${ASUP_MAKE_TARGET} VERSION=${VERSION} RELEASE=${RELEASE}
	@mkdir -p ${CURRENT_DIR}/autosupport
	@cp ${ASUP_TMP}/harvest-asup/bin/asup ${CURRENT_DIR}/autosupport

dev: build lint
	@echo "Deleting AutoSupport binary"
	@rm -rf autosupport/asup

fetch-asup:
	@./.github/fetch-asup ${ASUP_BIN} ${ASUP_BIN_VERSION} ${BIN_PLATFORM} 2>/dev/null   #Suppress Error in case of internet connectivity



