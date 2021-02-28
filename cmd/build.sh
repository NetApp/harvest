#!/bin/bash

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
    "harvest")
        harvest=true
        echo "build harvest-cli"
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
    "collector")
        collector=$2
        component="collector [$collector]"
        ;;
    "exporter")
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

# compile harvest-cli and manager
if [ $all == true ] || [ $harvest == true ]; then
    cd src/cli/
    go build -o ../../bin/harvest
    if [ $? -eq 0 ]; then
        info "compiled: /bin/harvest"
    else
        error "compilation failed"
        exit 1
    fi

    # compile manager and daemonize utils (manager uses daemonize)
    cd daemonize
    gcc daemonize.c -o ../../../bin/daemonize
    if [ $? -eq 0 ]; then
        info "compiled: /bin/daemonize"
    else
        error "compule failed"
        exit 1
    fi
    cd ../

    cd manager
    go build -o ../../../bin/manager
    if [ $? -eq 0 ]; then
        info "compiled: /bin/manager"
    else
        error "compule failed"
        exit 1
    fi
    cd ../

    cd config
    go build -o ../../bin/config
    if [ $? -eq 0 ]; then
        info "compiled: /bin/config"
    else
        error "compilation failed"
        exit 1
    fi
    cd ../../

    cd tools/zapi
    go build -o ../../../bin/zapitool
    if [ $? -eq 0 ]; then
        info "compiled: /bin/zapitool"
    else
        error "compilation failed"
        exit 1
    fi

    cd ../
    # @TODO migrate to GO
    cp grafana/grafana.py ../../bin/grafanatool
    info "copied /bin/grafanatool"

    cd ../
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

: '
# compile tool(s)
if [ $all == true ] || [ $tools == true ] || [ "$tool" != "" ]; then
    cd src/tools/
    declare -a files
    files=($(ls))
    for f in ${files[@]}; do
        if [ -d "$f" ]; then
            cd $f
            if [ $all == true ] || [ $tools == true ] || [ "$tool" == "$f" ]; then
                go build -o ../../../bin/"$f"
                if [ $? -eq 0 ]; then
                    info "compiled: /bin/$f"
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
'

# compile poller
if [ $all == true ] || [ $poller == true ]; then
    cd src/poller/
    go build -o ../../bin/poller
    cd ../../
    if [ $? -eq 0 ]; then
        info "compiled: /bin/poller"
    else
        error "compilation failed"
        exit 1
    fi
fi
