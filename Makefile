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
RELEASE      := $(shell git describe --tags --abbrev=0)
COMMIT       := $(shell git rev-parse --short HEAD)
BUILD_DATE   := `date +%FT%T%z`
LD_FLAGS     := "-X 'goharvest2/cmd/harvest/version.Release=$(RELEASE)' -X 'goharvest2/cmd/harvest/version.Commit=$(COMMIT)' -X 'goharvest2/cmd/harvest/version.BuildDate=$(BUILD_DATE)'"
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
	@cd cmd/harvest; GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags=$(LD_FLAGS) -o ../../bin/harvest

	@# Build the harvest poller
	@echo "Building poller"
	@cd cmd/poller/; GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags=$(LD_FLAGS) -o ../../bin/poller

	@# Build the daemonizer for the pollers
	@echo "Building daemonizer"
	@cd cmd/tools/daemonize; gcc daemonize.c -o ../../../bin/daemonize

	@# Build the zapi tool
	@echo "Building zapi tool"
	@cd cmd/tools/zapi; GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags=$(LD_FLAGS) -o ../../../bin/zapi

	@# Build the grafana tool
	@echo "Building grafana tool"
	@cd cmd/tools/grafana; GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags=$(LD_FLAGS) -o ../../../bin/grafana

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
HARVEST_RELEASE := harvest-${RELEASE}
TMP := /tmp/${HARVEST_RELEASE}
DIST := dist
dist-tar: all
	-rm -rf ${TMP}
	-rm -rf ${DIST}
	@mkdir ${TMP}
	@mkdir ${DIST}
	@cp -a bin conf docs grafana README.md LICENSE ${TMP}
	@cp -a harvest.example.yml ${TMP}/harvest.yml
	@tar --directory /tmp --create --gzip --file ${DIST}/${HARVEST_RELEASE}.tar.gz ${HARVEST_RELEASE}
	-rm -rf ${TMP}

###############################################################################
# Install targets
# If the ROOT is not set to "", then this is a development deploy which means
# we will be creating different users, and linking the deploy directory to
# the system setup.
###############################################################################
ROOT := ${BUILD_ROOT}
SUDO := sudo
HARVEST_USER := harvestu
HARVEST_GROUP := harvestu
GROUP_EXISTS := $(shell grep -c "^${HARVEST_GROUP}" /etc/group)
USER_EXISTS := $(shell grep -c "^${HARVEST_USER}" /etc/passwd)

install:
	@echo "Installing Harvest: ${HARVEST_VERSION}"

ifeq (${ROOT},)
	@echo "  Creating harvest user and group [${HARVEST_USER}:${HARVEST_GROUP}]"
	@if [ ${GROUP_EXISTS} -eq 0 ]; then                                     \
		${SUDO} groupadd -r ${HARVEST_GROUP};                           \
	else                                                                    \
		echo "    Harvest group already exists";                        \
	fi;

	@# Make sure that the user does not already exist
	@if [ ${USER_EXISTS} -eq 0 ]; then                                      \
	    ${SUDO} adduser --quiet --ingroup ${HARVEST_GROUP} --shell=/sbin/nologin ${HARVEST_USER}; \
	else                                                                    \
		echo "    Harvest user already exists";                         \
	fi;
endif

	@echo "  Creating package directories"
ifeq (${ROOT},)
	@${SUDO} mkdir -p /opt/harvest
	@${SUDO} mkdir -p /var/log/harvest
	@${SUDO} mkdir -p /var/run/harvest
else
	@mkdir -p ${ROOT}/deploy/opt/harvest
	@mkdir -p ${ROOT}/deploy/var/log/harvest
	@mkdir -p ${ROOT}/deploy/var/run/harvest
endif

ifeq (${ROOT},)
	@echo "  Setting user permissions"
	@${SUDO} chown -R ${HARVEST_USER}:${HARVEST_GROUP} /opt/harvest
	@${SUDO} chown -R ${HARVEST_USER}:${HARVEST_GROUP} /var/log/harvest
	@${SUDO} chown -R ${HARVEST_USER}:${HARVEST_GROUP} /var/run/harvest
endif

	@echo "  Copying config and binaries"
ifeq (${ROOT},)
	@${SUDO} cp -r  conf/ /opt/harvest
	@${SUDO} cp -r grafana/ /opt/harvest
	@${SUDO} cp harvest.example.yml /opt/harvest/harvest.yml
	@${SUDO} cp -r bin /opt/harvest
	@${SUDO} ln -sf /opt/harvest/bin/harvest /usr/bin/harvest
else
	@cp -r  conf/ ${ROOT}/deploy/opt/harvest/
	@cp -r grafana/ ${ROOT}/deploy/opt/harvest
	@cp harvest.example.yml ${ROOT}/deploy/opt/harvest/harvest.yml
	@cp -r bin ${ROOT}/deploy/opt/harvest/
	@${SUDO} ln -sf ${ROOT}/deploy/opt/harvest/ /opt
	@${SUDO} ln -sf ${ROOT}/deploy/var/log/harvest /var/log
	@${SUDO} ln -sf ${ROOT}/deploy/var/run/harvest /var/run
endif
	@echo "  Installation complete"


###############################################################################
# Uninstall target
###############################################################################
uninstall:
	@echo "Stopping Harvest"
	@/opt/harvest/bin/harvest stop

	@echo "Cleaning install files"
	@${SUDO} rm -rf /opt/harvest/bin
	@${SUDO} rm -rf /var/log/harvest
	@${SUDO} rm -rf /var/run/harvest
	@${SUDO} unlink /usr/bin/harvest
	@echo
	@echo "Configuration and Certificate files not removed in [${ROOT}/opt/harvest]"
	@echo "please remove manually if no longer needed."
	@echo
ifeq (${ROOT}, "")
	@echo "Removing harvest user and group"
	@${SUDO} userdel ${HARVEST_USER}
	@${SUDO} groupdel ${HARVEST_GROUP}
	@echo
endif
	@echo "Uninstall complete."
