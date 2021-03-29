# Copyright 2021 NetApp, Inc.  All Rights Reserved

VERSION := "2.0"

all: header harvest collectors exporters poller

header:
	@echo "    _  _                     _     ___   __   "
	@echo "   | || |__ _ _ ___ _____ __| |_  |_  ) /  \  "
	@echo "   | __ / _\` | '_\ V / -_|_-<  _|  / / | () | "
	@echo "   |_||_\__,_|_|  \_/\___/__/\__| /___(_)__/  "
	@echo


# Anything that needs to be done before we build everything
precheck:

clean:
	@echo "Cleaning harvest files"
	@rm -rf bin

###############################################################################
# Main Harvest sub services
###############################################################################
harvest: precheck
	@# Build the harvest cli
	@echo "Building harvest cli"
	@cd cmd/cli; go build -o ../../bin/harvest

	@# Build the daemonizer for the pollers
	@echo "Building daemonizer"
	@cd cmd/cli/daemonize; gcc daemonize.c -o ../../../bin/daemonize

	@# Build the manager
	@echo "Building manager"
	@cd cmd/cli/manager; go build -o ../../../bin/manager

	@# Build the config
	@echo "Building config"
	@cd cmd/cli/config; go build -o ../../../bin/config

	@# Build the zapi tool
	@echo "Building zapi tool"
	@cd cmd/tools/zapi; go build -o ../../../bin/zapi

	@# Build the grafana tool
	@echo "Building grafana tool"
	@cd cmd/tools/grafana; go build -o ../../../bin/grafana

###############################################################################
# Collectors
###############################################################################
COLLECTORS := $(shell ls cmd/collectors)
collectors:
	@echo "Building collectors:"
	@for collector in ${COLLECTORS}; do                                                   \
		cd cmd/collectors/$${collector};                                              \
		echo "  Building $${collector}";                                              \
		go build -buildmode=plugin -o ../../../bin/collectors/"$${collector}".so;     \
		if [ -d plugins ]; then                                                       \
			echo "    Building plugins for $${collector}";                        \
	        	cd plugins;                                                           \
	        	for plugin in `ls`; do                                                \
				echo "        Building: $${plugin}";                          \
				cd $${plugin};                                                \
				go build -buildmode=plugin -o ../../../../../bin/plugins/"$${collector}"/"$${plugin}".so; \
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
		go build -buildmode=plugin -o ../../../bin/exporters/"$${exporter}".so;       \
	       	cd - > /dev/null;                                                             \
	done

###############################################################################
# Poller
###############################################################################
poller: precheck
	@echo "Building poller"
	@cd cmd/poller/;                                                                      \
	go build -o ../../bin/poller


packages: precheck all

###############################################################################
# Install targets
###############################################################################
ROOT := ${BUILD_ROOT}
SUDO := sudo
HARVEST_USER := harvestu
HARVEST_GROUP := harvestu
GROUP_EXISTS := $(shell grep -c "^${HARVEST_GROUP}" /etc/group)
USER_EXISTS := $(shell grep -c "^${HARVEST_USER}" /etc/passwd)
install:
	@echo "Installing Harvest: ${VERSION}"

	@echo "  Creating harvest user and group [${HARVEST_USER}:${HARVEST_GROUP}]"
	@if [ ${GROUP_EXISTS} -eq 0 ]; then              \
		${SUDO} groupadd -r "${HARVEST_GROUP}";  \
	else                                             \
		echo "    Harvest group already exists"; \
	fi;

	@# Make sure that the user does not already exist
	@if [ ${USER_EXISTS} -eq 0 ]; then               \
		${SUDO} adduser --ingroup ${HARVEST_GROUP} --shell=/sbin/nologin ${HARVEST_USER}; \
	else                                             \
		echo "    Harvest user already exists";  \
	fi;

	@echo "  Creating package directories"
	@${SUDO} mkdir -p ${ROOT}/opt/harvest
	@${SUDO} mkdir -p ${ROOT}/etc/harvest
	@${SUDO} mkdir -p ${ROOT}/var/log/harvest
	@${SUDO} mkdir -p ${ROOT}/var/run/harvest

	@echo "  Setting user permissions"
	@${SUDO} chown -R ${HARVEST_USER}:${HARVEST_GROUP} ${ROOT}/opt/harvest
	@${SUDO} chown -R ${HARVEST_USER}:${HARVEST_GROUP} ${ROOT}/etc/harvest
	@${SUDO} chown -R ${HARVEST_USER}:${HARVEST_GROUP} ${ROOT}/var/log/harvest
	@${SUDO} chown -R ${HARVEST_USER}:${HARVEST_GROUP} ${ROOT}/var/run/harvest

	@echo "  Copying config and binaries"
	@${SUDO} mv config/ ${ROOT}/etc/harvest/
	@${SUDO} mv grafana/ ${ROOT}/etc/harvest/
	@${SUDO} mv harvest.yml ${ROOT}/etc/harvest/
	@${SUDO} mv * ${ROOT}/opt/harvest
	#ln -s $ROOT/opt/harvest/bin/harvest $ROOT/usr/local/bin/harvest
	@echo "  Installation complete"



