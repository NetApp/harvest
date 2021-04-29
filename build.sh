#!/bin/bash

#
# Copyright NetApp Inc, 2021 All rights reserved
#

COLOR_GREEN='\033[0;32m'
COLOR_RED='\033[0;31m'
COLOR_BOLD='\033[1m'
COLOR_END='\033[0m'

ROOT_DIR=$(pwd)

all=false
clean=false

harvest=false
poller=false

collectors=false
exporters=false
tools=false

component=""

collector=""
exporter=""
tool=""

function error {
    echo -e $COLOR_RED$COLOR_BOLD$1$COLOR_END
}

function info {
    echo -e $COLOR_GREEN$COLOR_BOLD$1$COLOR_END
}

# parse main command
case $1 in
    "all"|"")
        all=true
        echo "build all"
        ;;
    "clean")
        clean=true
        echo "clean"
        ;;
    "harvest"|"h")
        harvest=true
        echo "build harvest"
        ;;
    "poller"|"p")
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
    "collector"|"c")
        collector=$2
        component="collector [$collector]"
        ;;
    "exporter"|"e")
        exporter=$2
        component="exporter"
        echo "build exporter [$exporter]"
        ;;
    "tools")
        tools=true
        echo "build tools"
        ;;
    "tool")
        tool=$2
        component="tool"
        echo "build tool [$tool]"
        ;;
esac

# validate expected option
if [ "$component" == "collector" ] && [ "$collector" == "" ]; then
    echo "missing collector name"
    exit 1
fi

if [ "$component" == "exporter" ] && [ "$exporter" == "" ]; then
    echo "missing exporter name"
    exit 1
fi

if [ "$component" == "tool" ] && [ "$tool" == "" ]; then
    echo "missing tool name"
    exit 1
fi

# clean up binaries and exit
if [ $clean == true ]; then
    rm -rfv bin/*
    exit 0
fi

# compile tools --- @TODO
if [ "$tool" != "" ]; then
    error "not implemented yet"
    exit 0
fi

# compile harvest
if [ $all == true ] || [ $harvest == true ]; then
    cd src/harvest/
    go build -o ../../bin/harvest
    if [ $? -eq 0 ]; then
        info "compiled: /bin/harvest"
    else
        error "compilation failed"
        exit 1
    fi

    # compile manager and daemonize utils (manager uses daemonize)
    cd ../util/daemonize
    gcc daemonize.c -o ../../../bin/daemonize
    if [ $? -eq 0 ]; then
        info "compiled: /bin/daemonize"
    else
        error "compule failed"
        exit 1
    fi
    cd ../../

    cd tools/zapi
    go build -o ../../../bin/zapi
    if [ $? -eq 0 ]; then
        info "compiled: /bin/zapi"
    else
        error "compilation failed"
        exit 1
    fi

    cd ../grafana
    go build -o ../../../bin/grafana
    if [ $? -eq 0 ]; then
        info "compiled: /bin/grafana"
    else
        error "compilation failed"
        exit 1
    fi

    cd ../../../
fi

# compile collector(s)
if [ $all == true ] || [ $collectors == true ] || [ "$collector" != "" ]; then
    cd src/collectors/
    declare -a files
    files=($(ls))
    for f in ${files[@]}; do
        if [ -d "$f" ]; then
            if [ $all == true ] || [ $collectors == true ] || [ "$collector" == "$f" ]; then
                cd $f
                go build -buildmode=plugin -o ../../../bin/collectors/"$f".so

                if [ $? -eq 0 ]; then
                    info "compiled: /bin/collectors/$f.so"
                else
                    error "compiling [/src/collectors/$f] failed"
                    exit 1
                fi

                if [ -d "plugins" ]; then
                    echo "compiling plugins..."
                    cd plugins/
                    plugins=($(ls))
                    for p in ${plugins[@]}; do
                        if [ -d "$p" ]; then
                            cd $p
                            go build -buildmode=plugin -o ../../../../../bin/plugins/"$f"/"$p".so
                            if [ $? -eq 0 ]; then
                                echo -e "  compiled bin/plugins/$f/$p.so"
                            else
                                echo -e "  compiling [/src/collectors/$f/$p] failed"
                                exit 1
                            fi
                            cd ../
                        fi
                    done
                    cd ../
                fi
                cd ../
            fi
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
                if [ $? -eq 0 ]; then
                    info "compiled: /bin/exporters/$f.so"
                else
                    error "compilation failed"
                    exit 1
                fi
            fi
            cd ../
        fi
    done
    cd ../../
fi


# compile poller
if [ $all == true ] || [ $poller == true ]; then
    cd src/poller/
    go build -o ../../bin/poller
    if [ $? -eq 0 ]; then
        info "compiled: /bin/poller"
    else
        error "compilation failed"
        exit 1
    fi
    cd ../../
fi
