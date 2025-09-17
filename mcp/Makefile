.DEFAULT_GOAL := help

.PHONY: help deps clean build test fmt lint govulncheck package all

SHELL := /bin/bash
GOLANGCI_LINT_VERSION := latest
GOVULNCHECK_VERSION := latest
HARVEST_ENV := ../.harvest.env

# Read the environment file if it exists and export the uncommented variables
ifneq (,$(wildcard $(HARVEST_ENV)))
    include $(HARVEST_ENV)
	export $(shell sed '/^\#/d; s/=.*//' $(HARVEST_ENV))
endif

REQUIRED_GO_VERSION := $(GO_VERSION)

ifneq (, $(shell which go))
FOUND_GO_VERSION := $(shell go version | cut -d" " -f3 | cut -d"o" -f 2)
CORRECT_GO_VERSION := $(shell expr `go version | cut -d" " -f3 | cut -d"o" -f 2` \>= ${REQUIRED_GO_VERSION})
endif

RELEASE      ?= 1
VERSION      ?= $(shell expr `date +%Y.%m.%d | cut -c 3-`)
COMMIT       := $(shell git rev-parse --short HEAD)
BUILD_DATE   := `date +%FT%T%z`
LD_FLAGS     := "-X 'mcp-server/cmd/server/version.VERSION=$(VERSION)' -X 'mcp-server/cmd/server/version.Release=$(RELEASE)' -X 'mcp-server/cmd/server/version.Commit=$(COMMIT)' -X 'mcp-server/cmd/server/version.BuildDate=$(BUILD_DATE)'"
GOARCH       ?= amd64
GOOS         ?= $(shell go env GOOS)
CGO_ENABLED  ?= 0
BINARY_NAME  := harvest-mcp
BIN_DIR      := bin
DIST         := dist
PACKAGE_NAME := harvest-mcp-${VERSION}-${RELEASE}_${GOOS}_${GOARCH}

help: ## Display this help
	@echo "    _  _                     _     ___   __   "
	@echo "   | || |__ _ _ ___ _____ __| |_  |_  ) /  \  "
	@echo "   | __ / _\` | '_\ V / -_|_-<  _|  / / | () | "
	@echo "   |_||_\__,_|_|  \_/\___/__/\__| /___(_)__/  "
	@echo ""
	@echo "Harvest MCP Server"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

deps: ## Check dependencies
	@# Make sure that go exists
ifeq (${FOUND_GO_VERSION}, )
	$(error Harvest MCP requires that Go is installed and at least version: ${REQUIRED_GO_VERSION})
endif
	@# Check to make sure that GO is the correct version
ifeq ("${CORRECT_GO_VERSION}", "0")
	$(error Required Go version is ${REQUIRED_GO_VERSION}, but found ${FOUND_GO_VERSION})
endif

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf ${BIN_DIR} ${DIST}
	@go clean

fmt: ## Format the Go source files
	@echo "Formatting..."
	@go fmt ./...

lint: ## Run golangci-lint on the source files
	@echo "Linting..."
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} run ./...
	@go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest ./...

govulncheck: ## Run govulncheck on the source files
	@echo "Running vulnerability check..."
	@go run golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION} ./...

test: ## Run tests
	@echo "Running tests..."
	@go test -race -shuffle=on -v ./...

build: clean deps fmt lint govulncheck ## Build the MCP server binary with development checks
	@echo "Building MCP server..."
	@mkdir -p ${BIN_DIR}
	@GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build -trimpath -ldflags=$(LD_FLAGS) -o ${BIN_DIR}/${BINARY_NAME} ./cmd/server
	@echo "✅ Build complete: ${BIN_DIR}/${BINARY_NAME}"


package: clean deps build test ## Package MCP server binary with documentation
	@echo "Creating package..."
	@rm -rf ${DIST}
	@mkdir -p ${DIST}/${PACKAGE_NAME}
	@mkdir -p ${DIST}/${PACKAGE_NAME}/bin
	@mkdir -p ${DIST}/${PACKAGE_NAME}/docs
	@mkdir -p ${DIST}/${PACKAGE_NAME}/metadata
	@mkdir -p ${DIST}/${PACKAGE_NAME}/prompts
	@cp ${BIN_DIR}/${BINARY_NAME} ${DIST}/${PACKAGE_NAME}/bin/
	@cp README.md ${DIST}/${PACKAGE_NAME}/ 2>/dev/null || true
	@cp ../docs/ontap-metrics.md ${DIST}/${PACKAGE_NAME}/docs/ 2>/dev/null || echo "Warning: ontap-metrics.md not found"
	@cp ../docs/cisco-switch-metrics.md ${DIST}/${PACKAGE_NAME}/docs/ 2>/dev/null || echo "Warning: cisco-switch-metrics.md not found"
	@cp ../docs/storagegrid-metrics.md ${DIST}/${PACKAGE_NAME}/docs/ 2>/dev/null || echo "Warning: storagegrid-metrics.md not found"
	@cp metadata/*.json ${DIST}/${PACKAGE_NAME}/metadata/ 2>/dev/null || echo "Warning: No metadata files found"
	@cp -r prompts/* ${DIST}/${PACKAGE_NAME}/prompts/ 2>/dev/null || echo "Warning: No prompts found"
	@cd ${DIST} && tar --create --gzip --file ${PACKAGE_NAME}.tar.gz ${PACKAGE_NAME}/
	@rm -rf ${DIST}/${PACKAGE_NAME}
	@echo "✅ Package created: ${DIST}/${PACKAGE_NAME}.tar.gz"

all: package ## Build, test, and package everything
	@echo "✅ All tasks complete"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@cd .. && docker build -f mcp/Dockerfile --build-arg GO_VERSION=$(GO_VERSION) -t harvest-mcp:latest .
	@echo "✅ Docker image built: harvest-mcp:latest"

