#!/bin/bash

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


function get_pollers {
    cat status.txt | while read DATACENTER POLLER PORT STATUS PID ; do
        case "$PORT" in
            [0-9]*) echo "$POLLER $PORT"
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
    #rm -rf "$ROOT/src" && mkdir "$ROOT/src" && mv $RPM "$ROOT/src/"

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
    port=$3
    if [ -z "$port" ]; then
        docker run -it -v $ROOT:/harvest -e HARVEST_CMD="$cmd" -e HARVEST_ARG="$arg" harvest2/poller
    else
        docker run -it -v $ROOT:/harvest -p "$port":"$port" -e HARVEST_CMD="$cmd" -e HARVEST_ARG="$arg" harvest2/poller
    fi
}

function status {
    run "status"
}

function start {

    if [ -f "$ROOT/conf/pollers.txt" ]; then
        echo "retrieving list of pollers"
    else
        echo "no pollers defined"
        exit 1
    fi

    declare -a pollers=( `get_pollers` )

    for POLLERPORT in "${pollers[@]}"; do
        docker run -itd -v $ROOT:/harvest -p "$PORT":"$PORT" -e HARVEST_POLLER="$POLLER" harvest2/poller
        if [ ! $? -eq 0 ]; then
            echo "docker build failed"
            exit 1
        fi
        echo "started poller [$POLLER] serving at port [:"$PORT"]"
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
    "status")
        status
        ;;
    *)
        echo "unknown comand: $1"
        exit 1
        ;;
esac
