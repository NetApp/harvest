#!/bin/bash

BUILD_SOURCE=$(pwd)
BIN=`basename $0`

function usage {
    cat <<EOF_PRINT_HELP

    $BIN - Build Distribution Package

    Usage:
        $./$BIN <dist> <arch> <version> <release> [options]

    Arguments:
        dist            package format (rpm or dep)
        arch            architecture (x86_64, amd64, etc)
        version         version
        release         release

    Options
        -d, --docker    build in docker container

EOF_PRINT_HELP
}

function info {
    echo -e "\033[1m\033[45m$1\033[0m"
}

function error {
    echo -e "\033[1m\033[41m$1\033[0m"
}


function build {
    if [ $DOCKER ]; then

        info "building [harvest_$VERSION-$RELEASE_$ARCH.$DIST] in container"

        
        if [ "$DIST" == "rpm" ]; then 
            $DOCKER_IMAGE="centos"
        else
            $DOCKER_IMAGE="debian"
        fi

        cd "$BUILD_SOURCE/cmd/$DIST/$DOCKER_IMAGE"
        if [ ! $? -eq 0 ]; then
            error "docker image"
            exit 1
        fi

        docker build -t harvest2/$DIST .
        if [ ! $? -eq 0 ]; then
            error "build docker container"
            exit 1
        fi

        docker run -it -v $BUILD_SOURCE:/tmp/src -e $HARVEST_BUILD_SRC="/tmp/src" -e HARVEST_ARCH="$ARCH" -e HARVEST_VERSION="$VERSION" -e HARVEST_RELEASE="$RELEASE" harvest2/$DIST
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

        export HARVEST_BUILD_SRC="/tmp/src"
        export HARVEST_ARCH="$ARCH"
        export HARVEST_VERSION="$VERSION"
        export HARVEST_RELEASE="$RELEASE"
        
        sh "$BUILD_SOURCE/cmd/$DIST/build-$DIST.sh"

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
    usage
    exit 0
fi

DIST=$1
ARCH=$2
VERSION=$3
RELEASE=$4

if [ -z "$DIST" ] || [ -z "$ARCH" ] || [ -z "$VERSION" ] || [ -z "$RELEASE" ]; then
    usage
    exit 1
fi

if [ "$5" == "-d" ] || [ "$5" == "--docker" ]; then
    DOCKER=true
else
    DOCKER=false
fi

# build package in container
if [ "$DIST" == "rpm" ] || [ "$DIST" == "deb" ]; then
    build
else
    error "unknown package format: $DIST"
    exit 1
fi
