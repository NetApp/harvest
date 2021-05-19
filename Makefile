# Copyright 2021 NetApp, Inc.  All Rights Reserved

HARVEST_VERSION := "2.0"

all: header harvest collectors exporters

header:
	@echo "    _  _                     _     ___   __   "
	@echo "   | || |__ _ _ ___ _____ __| |_  |_  ) /  \  "
	@echo "   | __ / _\` | '_\ V / -_|_-<  _|  / / | () | "
	@echo "   |_||_\__,_|_|  \_/\___/__/\__| /___(_)__/  "
	@echo


###############################################################################
# Anything that needs to be done before we build everything
#  Check for GCC, GO version, etc and anything else we are dependent on.
###############################################################################
GCC_EXISTS := $(shell which gcc)
REQUIRED_GO_VERSION := 1.15
FOUND_GO_VERSION := $(shell go version | cut -d" " -f3 | cut -d"o" -f 2)
CORRECT_GO_VERSION := $(shell expr `go version | cut -d" " -f3 | cut -d"o" -f 2` \>= ${REQUIRED_GO_VERSION})
ifndef RELEASE
    RELEASE      := $(shell git describe --tags --abbrev=0)
endif
ifndef VERSION
    VERSION      := $(shell expr `date +%Y.%m.%d%H | cut -c 3-`)
endif
COMMIT       := $(shell git rev-parse --short HEAD)
BUILD_DATE   := `date +%FT%T%z`
LD_FLAGS     := "-X 'goharvest2/cmd/harvest/version.VERSION=$(VERSION)' -X 'goharvest2/cmd/harvest/version.Release=$(RELEASE)' -X 'goharvest2/cmd/harvest/version.Commit=$(COMMIT)' -X 'goharvest2/cmd/harvest/version.BuildDate=$(BUILD_DATE)'"
GOARCH ?= amd64
GOOS ?= linux

precheck:
	@# Check for GCC
ifeq (${GCC_EXISTS}, "")
	@echo
	@echo "Harvest requires that you have gcc installed."
	@echo
	@exit
endif
	@# Make sure that go exists
ifeq (${FOUND_GO_VERSION}, "")
	@echo
	@echo "Harvest requires that the go lang is installed and is at least version: ${REQUIRED_GO_VERSION}"
	@echo
	@exit
endif
	@# Check to make sure that GO is the correct version
ifeq ("${CORRECT_GO_VERSION}", "0")
	@echo
	@echo "Required go lang version is ${REQUIRED_GO_VERSION}, but found ${FOUND_GO_VERSION}"
	@echo
	@exit
endif

###############################################################################
# Clean the code base for rebuilding.
###############################################################################
clean:
	@echo "Cleaning harvest files"
	@rm -rf bin

###############################################################################
# Main Harvest sub services
###############################################################################
harvest: precheck
	@# Build the harvest cli
	@echo "Building harvest"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/harvest -ldflags=$(LD_FLAGS) cmd/harvest/harvest.go

	@# Build the harvest poller
	@echo "Building poller"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/poller -ldflags=$(LD_FLAGS) cmd/poller/poller.go

	@# Build the daemonizer for the pollers
	@echo "Building daemonizer"
	@cd cmd/tools/daemonize; gcc daemonize.c -o ../../../bin/daemonize

	@# Build the zapi tool
	@echo "Building zapi tool"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/zapi -ldflags=$(LD_FLAGS) cmd/tools/zapi/main/main.go

	@# Build the grafana tool
	@echo "Building grafana tool"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/grafana -ldflags=$(LD_FLAGS) cmd/tools/grafana/main/main.go

###############################################################################
# Collectors
###############################################################################
COLLECTORS := $(shell ls cmd/collectors)
collectors:
	@echo "Building collectors:"
	@for collector in ${COLLECTORS}; do                                                   \
		cd cmd/collectors/$${collector};                                              \
		echo "  Building $${collector}";                                              \
		GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags=$(LD_FLAGS) -buildmode=plugin -o ../../../bin/collectors/"$${collector}".so;     \
		if [ -d plugins ]; then                                                       \
			echo "    Building plugins for $${collector}";                        \
	        	cd plugins;                                                           \
	        	for plugin in `ls`; do                                                \
				echo "        Building: $${plugin}";                          \
				cd $${plugin};                                                \
				GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags=$(LD_FLAGS) -buildmode=plugin -o ../../../../../bin/plugins/"$${collector}"/"$${plugin}".so; \
				cd ../;                                                       \
			done;                                                                 \
			cd ../../../../;                                                      \
		else                                                                          \
	       		cd - > /dev/null;                                                     \
		fi;                                                                           \
	done

###############################################################################
# Exporters
###############################################################################
EXPORTERS := $(shell ls cmd/exporters)
exporters: precheck
	@echo "Building exporters:"
	@for exporter in ${EXPORTERS}; do                                                     \
		cd cmd/exporters/$${exporter};                                                \
		echo "  Building $${exporter}";                                               \
		GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags=$(LD_FLAGS) -buildmode=plugin -o ../../../bin/exporters/"$${exporter}".so;       \
	       	cd - > /dev/null;                                                             \
	done

packages: precheck all

###############################################################################
# Build tar gz distribution
###############################################################################
HARVEST_RELEASE := harvest-${VERSION}-${RELEASE}
DIST := dist
TMP := /tmp/${HARVEST_RELEASE}
dist-tar: all
	-rm -rf ${TMP}
	-rm -rf ${DIST}
	@mkdir ${TMP}
	@mkdir ${DIST}
	@cp -a bin conf docs grafana README.md LICENSE ${TMP}
	@cp -a harvest.example.yml ${TMP}/harvest.yml
	@tar --directory /tmp --create --gzip --file ${DIST}/${HARVEST_RELEASE}.tar.gz ${HARVEST_RELEASE}
	-rm -rf ${TMP}
