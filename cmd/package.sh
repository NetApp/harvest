#!/bin/bash

ROOT=$(pwd)
BIN=`basename $0`

DIST=$1
ARCH=$2
VERSION=$3
RELEASE=$4

function usage {
    cat <<EOF_PRINT_HELP

    $BIN - Build Distribution Package

    Usage:
        $./$BIN DIST ARCH VERSION

    Arguments:
        DIST        package format (rpm or dep)
        ARCH        architecture (x86_64, amd64, etc)
        VERSION     version
        RELEASE     release

EOF_PRINT_HELP
}

function info {
    echo -e "\033[1m\033[45m$1\033[0m"
}

function error {
    echo -e "\033[1m\033[41m$1\033[0m"
}

function buildrpm {
    info "building RPM ($ARCH) in container"
    cd "$ROOT/cmd/rpm/centos"
    docker build -t harvest2/rpm .
    EXCODE=$?
    if [ ! $EXCODE -eq 0 ]; then
        error "build docker container failed, aborting"
    else
        docker run -it -v $ROOT:/tmp/src -e HARVEST_ARCH="$ARCH" -e HARVEST_VERSION="$VERSION" -e HARVEST_RELEASE="$RELEASE" harvest2/rpm
        EXCODE=$?
        if [ ! $EXCODE -eq 0 ]; then
            error "run docker container failed"
        fi
    fi
    cd "$ROOT"
    exit $EXCODE
}

# defaults
if [ -z "$DIST" ] || [ "$DIST" == "help" ] || [ -z "$ARCH" ] || [ -z "$VERSION" ] || [ -z "$RELEASE" ]; then
    usage
    exit 1
fi

# build package in container
if [ "$DIST" == "rpm" ]; then
    buildrpm
else
    error "invalid package format: $DIST"
fi
