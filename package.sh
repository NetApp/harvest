#!/bin/bash

#
# Copyright NetApp Inc, 2021 All rights reserved
#

BUILD_SOURCE=$(pwd)
BIN=`basename $0`

help() {
    cat <<EOF_PRINT_HELP

    $BIN - Build Distribution Package

    Usage:
        $./$BIN <dist> <arch> <version> <release> [options]

    Arguments:
        dist            package format (rpm or dep)
        arch            target architecture, one of:
                            x86_64/amd64, amd64, arm64, arm32, armhf
                            (other architectures should work as well,
                            but we haven't tested)
        version         version
        release         release

    Options
        -d, --docker    build in docker container

EOF_PRINT_HELP
}

usage() {
    echo "Usage:"
    echo "$./$BIN <dist> <arch> <version> <release> [options]"
}

info() {
    echo -e "\033[1m\033[45m$1\033[0m"
}

error() {
    echo -e "\033[1m\033[41m$1\033[0m"
}


build () {
    if [ "$DOCKER" == "true" ]; then

        info "building [harvest_$VERSION-$RELEASE_$ARCH.$DIST] in container"

        if [ "$DIST" == "rpm" ]; then 
            cd "$BUILD_SOURCE/$DIST/centos"
        else
            cd "$BUILD_SOURCE/$DIST/debian"
        fi

        if [ ! $? -eq 0 ]; then
            error "docker image"
            exit 1
        fi

        docker build -t harvest2/$DIST .
        if [ ! $? -eq 0 ]; then
            error "build docker container"
            exit 1
        fi

        docker run -it -v $BUILD_SOURCE:/tmp/src -e HARVEST_BUILD_SRC="/tmp/src" -e HARVEST_ARCH="$ARCH" -e HARVEST_VERSION="$VERSION" -e HARVEST_RELEASE="$RELEASE" harvest2/$DIST
        if [ ! $? -eq 0 ]; then
            error "run docker container"
            exit 1
        else
            info "build in docker complete"
        fi

        cd $BUILD_SOURCE
        exit 0

    else

        info "building [harvest_$VERSION-$RELEASE_$ARCH.$DIST] on local system"

        export HARVEST_BUILD_SRC="$BUILD_SOURCE"
        export HARVEST_ARCH="$ARCH"
        export HARVEST_VERSION="$VERSION"
        export HARVEST_RELEASE="$RELEASE"
        
        sh "$BUILD_SOURCE/$DIST/build-$DIST.sh"

        if [ ! $? -eq 0 ]; then
            error "run build script"
            exit 1
        else
            info "build complete"
        fi

        cd $BUILD_SOURCE
        exit 0
    fi
}

# defaults

if [ "$1" == "help" ] || [ "$1" == "-h" ] || [ "$1" == "--help" ]; then
    help
    exit 0
fi

DIST=$1
ARCH=$2
VERSION=$3
RELEASE=$4

if [[ -z "$DIST" || -z "$ARCH" || -z "$VERSION" || -z "$RELEASE" ]]; then
    usage
    exit 1
fi

if [[ "$5" == "-d" || "$5" == "--docker" ]]; then
    echo " -- build in docker"
    DOCKER=true
else
    echo " -- build locally"
    DOCKER=false
fi

if [[ "$ARCH" == "x86_64" && "$DIST" == "deb" ]]; then
    ARCH="amd64"
fi

# build package in container
if [[ "$DIST" == "rpm" || "$DIST" == "deb" ]]; then
    build
else
    error "unknown package format: $DIST"
    exit 1
fi
