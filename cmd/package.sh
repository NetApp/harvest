#!/bin/bash

ROOT=$(pwd)
BIN=`basename $0`

DIST=$1
ARCH=$2
VERSION=$3

GREEN='\033[0;32m'
RED='\033[0;31m'
BOLD='\033[1m'
END='\033[0m'

function usage {
    cat <<EOF_PRINT_HELP

    $BIN - Build Distribution Package

    Usage:
        $./$BIN DIST ARCH VERSION

    Arguments:
        DIST        package format (rpm or dep)
        ARCH        architecture (x86_64, amd64, etc)
        VERSION     package version

EOF_PRINT_HELP
}

function error {
    echo -e $RED$BOLD$1$END
}

function info {
    echo -e $GREEN$BOLD$1$END
}

function buildrpm {
    info "building RPM ($ARCH) in container"
    cd "$ROOT/cmd/rpm/centos"
    docker build -t harvest2/rpm .
    EXCODE=$?
    if [ ! $EXCODE -eq 0 ]; then
        error "build docker container failed, aborting"
    else
        docker run -it -v $ROOT:/tmp/src -e HARVEST_BUILD_ARCH="$ARCH" -e HARVEST_BUILD_VERSION="$VERSION" harvest2/rpm
        EXCODE=$?
        if [ ! $EXCODE -eq 0 ]; then
            error "run docker container failed"
        fi
    fi
    cd "$ROOT"
    exit $EXCODE
}

# defaults
if [ -z "$DIST" ] || [ "$DIST" == "help" ] || [ -z "$ARCH" ] || [ -z "$VERSION" ]; then
    usage
    exit 1
fi

# build package in container
if [ "$DIST" == "rpm" ]; then
    buildrpm
else
    error "invalid package format: $DIST"
fi
