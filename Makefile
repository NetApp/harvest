# Copyright 2021 NetApp, Inc.  All Rights Reserved
.DEFAULT_GOAL:=help

.PHONY: help deps clean build test fmt lint tidy tidy-check package asup dev fetch-asup ci

SHELL := /bin/bash
HARVEST_ENV := .harvest.env
HELM_CHART_DIR := deploy/helm/harvest
# Every Go module in the repo. integration/ and mcp/ consume the root module from
# disk (replace github.com/netapp/harvest/v2 => ../), so a root dependency bump
# raises their indirect requirements too and their go.mod files have to follow.
GO_MODULES := . integration mcp cmd/tools/grafana/promqlfmt
PROMQLFMT_DIR := cmd/tools/grafana/promqlfmt
# Absolute so it can be passed to a check running in any module directory
WWHRD_CONFIG := $(CURDIR)/.wwhrd.yml

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
LD_FLAGS     := "-X 'github.com/netapp/harvest/v2/cmd/harvest/version.VERSION=$(VERSION)' -X 'github.com/netapp/harvest/v2/cmd/harvest/version.Release=$(RELEASE)' -X 'github.com/netapp/harvest/v2/cmd/harvest/version.Commit=$(COMMIT)' -X 'github.com/netapp/harvest/v2/cmd/harvest/version.BuildDate=$(BUILD_DATE)'"
GOARCH ?= amd64
GOOS ?= linux
CGO_ENABLED ?= 0
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
MKDOCS_EXISTS := $(shell which mkdocs)
FETCH_ASUP_EXISTS := $(shell which ./.github/fetch-asup)

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
		find ./bin -type f -not -name "*asup*" -exec rm -f {} +; \
	fi

test: ## Run tests
	@echo "Testing"
	@go test -race -shuffle=on ./...
	@echo "Testing mcp module"
	@$(MAKE) --no-print-directory -C mcp test

fmt: ## Format the go source files
	@echo "Formatting"
	@go fmt ./...

tidy: ## Tidy go.mod and go.sum in every module
	@echo "Tidying"
	@for m in $(GO_MODULES); do echo "  $$m"; (cd $$m && go mod tidy) || exit 1; done
	@go mod vendor

tidy-check: ## Check that every module's go.mod and go.sum are tidy
	@echo "Checking go.mod files are tidy"
	@for m in $(GO_MODULES); do \
		(cd $$m && go mod tidy -diff) || { echo "ERROR: $$m/go.mod is not tidy (see diff above) -- run 'make tidy'"; exit 1; }; \
	done

lint: tidy-check ## Run golangci-lint on the source files
	@echo "Linting"
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} run ./...
	@cd integration && go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} run ./...
	@cd $(PROMQLFMT_DIR) && go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} run ./...
	@echo "Linting mcp module"
	@$(MAKE) --no-print-directory -C mcp lint

govulncheck: tidy-check ## Run govulncheck on the source files
	@echo "Govulnchecking"
	@go run golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION} ./...
	@cd integration && go run golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION} ./...
	@cd $(PROMQLFMT_DIR) && go run golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION} ./...
	@echo "Govulnchecking mcp module"
	@$(MAKE) --no-print-directory -C mcp govulncheck

mkdocs:
ifeq (${MKDOCS_EXISTS}, )
	$(error mkdocs task requires that you have https://squidfunk.github.io/mkdocs-material/getting-started/ installed.)
endif
	mkdocs serve

build: clean deps fmt harvest fetch-asup ## Build the project

package: clean deps build test dist-tar ## Package Harvest binary

all: package ## Build, Test, Package

harvest: deps
	@mkdir -p bin
	@# Build the harvest and poller cli
	@echo "Building"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build -trimpath -o bin -ldflags=$(LD_FLAGS) ./cmd/harvest ./cmd/poller
	@cp service/contrib/grafana bin; chmod +x bin/grafana

###############################################################################
# Build tar gz distribution
###############################################################################
dist-tar:
	@echo "Building tar"
	@rm -rf ${TMP}
	@rm -rf ${DIST}
	@mkdir ${TMP}
	@mkdir ${DIST}
	@cp -r bin conf container grafana service cert autosupport README.md LICENSE prom-stack.tmpl ${TMP}
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

dev: build lint govulncheck
	@echo "Deleting AutoSupport binary"
	@rm -rf autosupport/asup

fetch-asup:
ifneq (${FETCH_ASUP_EXISTS}, )
	@./.github/fetch-asup ${ASUP_BIN} ${ASUP_BIN_VERSION} ${BIN_PLATFORM} 2>/dev/null   #Suppress Error in case of internet connectivity
endif

docs: mkdocs ## Serve docs for local dev

license-check: tidy-check
	@echo "Licence checking"
	@go run github.com/frapposelli/wwhrd@latest check -q -t
	@cd integration && go run github.com/frapposelli/wwhrd@latest check -q -t -f $(WWHRD_CONFIG)
	@cd $(PROMQLFMT_DIR) && go run github.com/frapposelli/wwhrd@latest check -q -t -f $(WWHRD_CONFIG)
	@echo "Licence checking mcp module"
	@$(MAKE) --no-print-directory -C mcp license-check WWHRD_CONFIG=$(WWHRD_CONFIG)

ci: clean deps fmt harvest lint test govulncheck license-check

ci-local: ## Run CI locally
ifeq ($(origin ci), undefined)
	@echo ci-local requires that both the ci and admin variables are defined at the CLI, ci is missing. e.g.:
	@echo make ci=/path/to/harvest.yml admin=/path/to/harvest_admin.yml ci-local
	@exit 1
endif
ifeq ($(origin admin), undefined)
	@echo ci-local requires that both the ci and admin variables are defined at the CLI, admin is missing. e.g.
	@echo make ci=/path/to/harvest.yml admin=/path/to/harvest_admin.yml ci-local
	@exit 1
else
# Both variables are defined
	-@docker stop $$(docker ps -a --format '{{.ID}} {{.Names}}' | grep -E 'grafana|prometheus|poller') 2>/dev/null || true
	-@docker rm $$(docker ps -a --format '{{.ID}} {{.Names}}' | grep -E 'grafana|prometheus|poller') 2>/dev/null || true
	-@docker volume rm harvest_grafana_data harvest_prometheus_data 2>/dev/null || true
	@cp "${admin}" integration/test/harvest_admin.yml
	@cp "${ci}" integration/test/harvest.yml
	@./bin/harvest generate docker full --config "${ci}" --port --output harvest-compose.yml
	@docker build -f container/onePollerPerContainer/Dockerfile -t ghcr.io/netapp/harvest:latest . --no-cache --build-arg GO_VERSION=${GO_VERSION} --build-arg VERSION=${VERSION}
	@docker compose -f prom-stack.yml -f harvest-compose.yml up -d --remove-orphans
	VERSION=${VERSION} INSTALL_DOCKER=1 ./integration/test/test.sh
	VERSION=${VERSION} REGRESSION=1 ./integration/test/test.sh
	VERSION=${VERSION} ANALYZE_DOCKER_LOGS=1 ./integration/test/test.sh
	VERSION=${VERSION} CHECK_METRICS=1 ./integration/test/test.sh
	bin/harvest generate metrics --config "${ci}" --poller dc1 --prom-url http://localhost:9090
endif

helm-lint: ## Lint Helm chart against all CI values files
	@echo "Linting Helm chart"
	helm lint --strict $(HELM_CHART_DIR)

	@echo "Linting Helm chart with CI values files"
	@for f in $(HELM_CHART_DIR)/ci/*.yaml; do \
		echo "case: $$f"; \
		helm lint --strict $(HELM_CHART_DIR) --values $$f || exit 1; \
	done

helm-docs: ## Generate Helm chart documentation
	@echo "Generating Helm chart documentation"
	helm-docs --chart-search-root $(HELM_CHART_DIR)
