# Copyright 2021 NetApp, Inc.  All Rights Reserved
.DEFAULT_GOAL:=help

.PHONY: help deps clean build test fmt lint package dev ci

SHELL := /bin/bash
REQUIRED_GO_VERSION := 1.24
GOLANGCI_LINT_VERSION := latest
GOVULNCHECK_VERSION := latest
ifneq (, $(shell which go))
FOUND_GO_VERSION := $(shell go version | cut -d" " -f3 | cut -d"o" -f 2)
CORRECT_GO_VERSION := $(shell expr `go version | cut -d" " -f3 | cut -d"o" -f 2` \>= ${REQUIRED_GO_VERSION})
endif
RELEASE      ?= 1
VERSION      ?= $(shell expr `date +%Y.%m.%d | cut -c 3-`)
COMMIT       := $(shell git rev-parse --short HEAD)
BUILD_DATE   := `date +%FT%T%z`
LD_FLAGS     := "-X 'github.com/netapp/harvest/v2/cmd/harvest/version.VERSION=$(VERSION)' -X 'github.com/netapp/harvest/v2/cmd/harvest/version.Release=$(RELEASE)' -X 'github.com/netapp/harvest/v2/cmd/harvest/version.Commit=$(COMMIT)' -X 'github.com/netapp/harvest/v2/cmd/harvest/version.BuildDate=$(BUILD_DATE)'"
GOARCH ?= amd64
GOOS ?= linux
CGO_ENABLED ?= 0
HARVEST_PACKAGE := harvest-${VERSION}-${RELEASE}_${GOOS}_${GOARCH}
DIST := dist
TMP := /tmp/${HARVEST_PACKAGE}
CURRENT_DIR = $(shell pwd)
BIN_PLATFORM ?= linux
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
MKDOCS_EXISTS := $(shell which mkdocs)
HARVEST_ENV := .harvest.env
TIMESTAMP := $(shell date +%Y%m%d%H%M%S)

# Read the environment file if it exists and export the uncommented variables
ifneq (,$(wildcard $(HARVEST_ENV)))
    include $(HARVEST_ENV)
	export $(shell sed '/^\#/d; s/=.*//' $(HARVEST_ENV))
endif

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-11s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

header:
	@echo "    _  _                     _     ___   __   "
	@echo "   | || |__ _ _ ___ _____ __| |_  |_  ) /  \  "
	@echo "   | __ / _\` | '_\ V / -_|_-<  _|  / / | () | "
	@echo "   |_||_\__,_|_|  \_/\___/__/\__| /___(_)__/  "

deps: header ## Check dependencies
	@# Make sure that go exists
ifeq (${FOUND_GO_VERSION}, )
	$(error Harvest requires that Go is installed and at least version: ${REQUIRED_GO_VERSION})
endif
	@# Check to make sure that GO is the correct version
ifeq ("${CORRECT_GO_VERSION}", "0")
	$(error Required Go version is ${REQUIRED_GO_VERSION}, but found ${FOUND_GO_VERSION})
endif

clean: ## Cleanup the project binary (bin) folders
	@echo "Cleaning Harvest files"
	@if [ -d bin ]; then \
		find ./bin -type f -exec rm -f {} +; \
	fi

test: ## Run tests
	@echo "Testing"
	@# The ldflags force the old Apple linker to suppress ld warning messages on MacOS
	@# See https://github.com/golang/go/issues/61229#issuecomment-1988965927
	@go test -ldflags=-extldflags=-Wl,-ld_classic -race -shuffle=on ./...

fmt: ## Format the go source files
	@echo "Formatting"
	@go fmt ./...

lint: ## Run golangci-lint on the source files
	@echo "Linting"
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} run ./...
	@cd integration && go mod tidy && go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} run ./...

govulncheck: ## Run govulncheck on the source files
	@echo "Govulnchecking"
	@go run golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION} ./...
	@cd integration && go mod tidy && go run golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION} ./...

mkdocs:
ifeq (${MKDOCS_EXISTS}, )
	$(error mkdocs task requires that you have https://squidfunk.github.io/mkdocs-material/getting-started/ installed.)
endif
	mkdocs serve

build: clean deps fmt harvest ## Build the project

package: clean deps build test dist-tar ## Package Harvest binary

all: package ## Build, Test, Package

harvest: deps
	@mkdir -p bin
	@# Build the harvest and poller cli
	@echo "Building"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build -trimpath -o bin -ldflags=$(LD_FLAGS) ./cmd/harvest ./cmd/poller

###############################################################################
# Build tar gz distribution
###############################################################################
dist-tar:
	@echo "Building tar"
	@rm -rf ${TMP}
	@rm -rf ${DIST}
	@mkdir ${TMP}
	@mkdir ${DIST}
	@cp -r bin conf container grafana service cert README.md LICENSE prom-stack.tmpl ${TMP}
	@cp harvest.yml ${TMP}/harvest.yml
	@tar --directory /tmp --create --gzip --file ${DIST}/${HARVEST_PACKAGE}.tar.gz ${HARVEST_PACKAGE}
	@rm -rf ${TMP}
	@echo "tar artifact @" ${DIST}/${HARVEST_PACKAGE}.tar.gz

dev: build lint govulncheck

docs: mkdocs ## Serve docs for local dev

test-local: test promtool

license-check:
	@echo "Licence checking"
	@go run github.com/frapposelli/wwhrd@latest check -q -t
	@cd integration && go mod tidy && go run github.com/frapposelli/wwhrd@latest check -q -t -f ../.wwhrd.yml

ci: clean deps fmt harvest lint test govulncheck license-check

build-image: ## Run CI locally
	@docker build --label "source_repository=https://github.com/sapcc/dme-storage-harvest" -f container/onePollerPerContainer/Dockerfile -t ${IMAGE_TAG}:${HARVEST_RELEASE}-${TIMESTAMP} . --no-cache --build-arg GO_VERSION=${GO_VERSION} --build-arg VERSION=${VERSION}

