#!/bin/bash

#
# Copyright NetApp Inc, 2021 All rights reserved
#

ROOT=$(pwd)
PROJECT="HarvestDocler"
BASENAME=`basename $0`

function usage {
    cat <<EOF_PRINT_HELP

    $PROJECT - Build & run Harvest pollers in containers

    Usage:
        $./$BASENAME COMMAND [ARG]

    The commands are:
        build       build image and configure Harvest
        start       start all configured pollers
        help        show help message from harvest
        ...         any other command accepted by Harvest

    Note that this tool acts as a wrapper around the, so once
    you have successfully built the image, it will accept
    all commands normally accepted by harvest.
EOF_PRINT_HELP
}

function print_pollers {
    status | while read DATACENTER POLLER PORT STATUS PID ; do
        echo "$DATACENTER -- $POLLER -- $PORT -- $STATUS -- $PID"
        case "$PORT" in
            [0-9]*) echo "!!!"
        esac
    done
}

function get_pollers {
    status | while read DATACENTER POLLER PORT STATUS PID ; do
        case "$PORT" in
            [0-9]*) echo "$POLLER;$PORT"
            ;;
        esac
    done
}

function build {
    ls centos/Dockerfile
    if [ ! $? -eq 0 ]; then
        echo "docker file in [$ROOT/centos/] missing"
        exit 1
    fi

    ls centos/*.rpm
    if [ ! $? -eq 0 ]; then
        echo "rpm package missing"
        exit 1
    fi

    rm -rf "$ROOT/conf" && mkdir "$ROOT/conf"
    rm -rf "$ROOT/pid" && mkdir "$ROOT/pid"
    rm -rf "$ROOT/log" && mkdir "$ROOT/log"

    cd centos
    docker build -t harvest2/poller .
    if [ ! $? -eq 0 ]; then
        echo "docker build failed"
        exit 1
    fi
    cd ../
    run "config" "welcome"
}

function run {
    cmd=$1
    arg=$2
    docker run -it -v $ROOT:/harvest -e HARVEST_CMD="$cmd" -e HARVEST_ARG="$arg" -e HARVEST_DOCKER="yes" harvest2/poller
}

function status {
    run "status" "-v"
}

function start {

    declare -a pollers=( `get_pollers` )
    declare -a params

    for p in "${pollers[@]}"; do
        params=(`echo $p | tr ";" " "`)
        POLLER="${params[0]}"
        PORT="${params[1]}"

        image=(`docker container ls -a | grep "harvest2-$POLLER"`)
        
        if [ -z "$image" ]; then
            docker run -d -p $PORT:$PORT -v $ROOT:/harvest -e HARVEST_CMD="start" -e HARVEST_ARG="$POLLER" -e HARVEST_DOCKER="yes" --name "harvest2-$POLLER" harvest2/poller
        else
            docker container start "harvest2-$POLLER"
        fi
        
        if [ ! $? -eq 0 ]; then
            echo "docker build failed"
            exit 1
        fi
    done
}

function stop {
    declare -a pollers=( `get_pollers` )
    declare -a params

    for p in "${pollers[@]}"; do
        params=(`echo $p | tr ";" " "`)
        POLLER="${params[0]}"
        docker container stop "harvest2-$POLLER"
    done
}

case "$1" in
    "help"|"-h"|"--help","")
        usage
        exit 0
        ;;
    "build")
        build
        ;;
    "start")
        start $2 $3
        ;;
    "stop")
        stop
        ;;
    "config")
        run "config" "welcome"
        ;;
    *)
        echo "unknown comand: $1"
        exit 1
        ;;
esac
