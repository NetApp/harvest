#!/bin/bash

all=false
clean=false

poller=false

collectors=false
exporters=false
plugins=false

expected_name=""

collector=""
exporter=""
plugin=""

case $1 in
    "all"|"")
        all=true
        echo "build all"
        ;;
    "clean")
        clean=true
        echo "clean all"
        ;;
    "poller")
        poller=true
        echo "build poller"
        ;;
    "collectors")
        collectors=true
        echo "build collectors"
        ;;
    "exporters")
        exporters=true
        echo "build exporters"
        ;;
    "plugins")
        plugins=true
        echo "build plugins"
        ;;
    "collector")
        collector=$2
        expected_name="collector"
        ;;
    "exporter")
        exporter=$2
        expected_name="exporter"
        echo "build exporter [$exporter]"
        ;;
    "plugin")
        plugin=$2
        expected_name="plugin"
        echo "build plugin [$plugin]"
        ;;
esac

if [ "$expected_name" == "collector" ] && [ "$collector" == "" ]; then
    echo "missing collector name"
    exit 1
fi

if [ "$expected_name" == "exporter" ] && [ "$exporter" == "" ]; then
    echo "missing exporter name"
    exit 1
fi

if [ "$expected_name" == "plugin" ] && [ "$plugin" == "" ]; then
    echo "missing plugin name"
    exit 1
fi

if [ $clean == true ]; then
    cd bin
    rm poller
    rm collectors/*so
    rm exporters/*so
    rm plugins/*so
    cd ..
    exit 0
fi

# compile poller
if [ $all == true ] || [ $poller == true ]; then
    cd src/poller/
    go build -o ../../bin/poller
    cd ../../
fi

# compile collector(s)
if [ $all == true ] || [ $collectors == true ] || [ "$collector" != "" ]; then
    cd src/collectors/
    declare -a files
    files=($(ls))
    for f in ${files[@]}; do
        if [ -d "$f" ]; then
            cd $f
            if [ $all == true ] || [ $collectors == true ] || [ "$collector" == "$f" ]; then
                go build -buildmode=plugin -o ../../../bin/collectors/"$f".so
            fi
            cd ../
        fi
    done
    cd ../../
fi

# compile exporter(s)
if [ $all == true ] || [ $exporters == true ] || [ "$exporter" != "" ]; then
    cd src/exporters/
    declare -a files
    files=($(ls))
    for f in ${files[@]}; do
        if [ -d "$f" ]; then
            cd $f
            if [ $all == true ] || [ $exporters == true ] || [ "$exporter" == "$f" ]; then
                go build -buildmode=plugin -o ../../../bin/exporters/"$f".so
            fi
            cd ../
        fi
    done
    cd ../../
fi

# compile plugins(s)
if [ $all == true ] || [ $plugins == true ] || [ "$plugin" != "" ]; then
    cd src/plugins/
    declare -a files
    files=($(ls))
    for f in ${files[@]}; do
        if [ -d "$f" ]; then
            cd $f
            if [ $all == true ] || [ $plugins == true ] || [ "$plugin" == "$f" ]; then
                go build -buildmode=plugin -o ../../../bin/plugins/"$f".so
            fi
            cd ../
        fi
    done
    cd ../../
fi