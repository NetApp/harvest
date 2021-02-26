#!/bin/bash

function get_pollers {
    cat status.txt | while read DATACENTER POLLER PORT STATUS PID ; do
        case "$PORT" in
            [0-9]*) echo "$POLLER $PORT"
            ;;
        esac
    done
}

ROOT=$(pwd)

ls centos/
if [ ! $? -eq 0 ]; then
    echo "docker file in [$ROOT/centos/] missing"
    exit 1
fi

RPM=$(ls *.rpm)
if [ -z "$RPM" ]; then
    echo "rpm package missing"
    exit 1
fi

rm -rf "$ROOT/conf" && mkdir "$ROOT/conf"
rm -rf "$ROOT/pid" && mkdir "$ROOT/pid"
rm -rf "$ROOT/log" && mkdir "$ROOT/log"
rm -rf "$ROOT/src" && mkdir "$ROOT/src" && mv $RPM "$ROOT/src/"

docker build -v $ROOT:/harvest/ -t harvest2 .
if [ ! $? -eq 0 ]; then
    echo "docker build failed"
    exit 1
fi

if [ -f "$ROOT/conf/pollers.txt" ]; then
    echo "retrieving list of pollers"
else
    echo "no pollers defined"
    exit 1
fi

declare -a pollers=( `get_pollers` )

for POLLER PORT in "${pollers[@]}"; do
    docker run -itd -v $ROOT:/harvest -p "$PORT":"$PORT" -e HARVEST_POLLER="$POLLER" harvest2
    if [ ! $? -eq 0 ]; then
        echo "docker build failed"
        exit 1
    fi
    echo "started poller [$POLLER] serving at port [:"$PORT"]"
done