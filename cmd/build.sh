#!/bin/bash


COLOR_END='\033[0m'
COLOR_GREEN='\033[0;32m'
COLOR_RED='\033[0;31m'

ROOT_DIR=$(pwd)

all=false
clean=false

harvest=false
poller=false

collectors=false
exporters=false
plugins=false
tools=false

expected_name=""

collector=""
exporter=""
plugin=""
tool=""

case $1 in
    "all"|"")
        all=true
        echo "build all"
        ;;
    "clean")
        clean=true
        echo "clean all"
        ;;
    "harvest")
        harvest=true
        echo "build harvest"
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
    "tools")
        tools=true
        echo "build tools"
        ;;
    "tool")
        tool=$2
        expected_name="tool"
        echo "build tool [$tool]"
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

if [ "$expected_name" == "tool" ] && [ "$tool" == "" ]; then
    echo "missing tool name"
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

# compile harvest
if [ $all == true ] || [ $harvest == true ]; then
    cd src/harvest/
    go build -o ../../bin/harvest
    if [ $? -eq 0 ]; then
        echo -e "${COLOR_GREEN}compiled: /bin/harvest ${COLOR_END}"
    else
        echo -e "${COLOR_RED}compilation failed ${COLOR_END}"
    fi
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

                if [ $? -eq 0 ]; then
                    echo -e "${COLOR_GREEN}compiled: /bin/collectors/$f.so ${COLOR_END}"
                else
                    echo -e "${COLOR_RED}compiling [/src/collectors/$f] failed ${COLOR_END}"
                fi
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
                        fi
                        cd ../
                    fi
                done
                cd ../
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
                if [ $? -eq 0 ]; then
                    echo -e "${COLOR_GREEN}compiled: /bin/exporters/$f.so ${COLOR_END}"
                else
                    echo -e "${COLOR_RED}compilation failed ${COLOR_END}"
                fi
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
                if [ $? -eq 0 ]; then
                    echo -e "${COLOR_GREEN}compiled: /bin/plugins/$f.so ${COLOR_END}"
                else
                    echo -e "${COLOR_RED}compilation failed ${COLOR_END}"
                fi
            fi
            cd ../
        fi
    done
    cd ../../
fi

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
                    echo -e "${COLOR_GREEN}compiled: /bin/$f ${COLOR_END}"
                else
                    echo -e "${COLOR_RED}compilation failed ${COLOR_END}"
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
    cd ../../
    if [ $? -eq 0 ]; then
        echo -e "${COLOR_GREEN}compiled: /bin/poller ${COLOR_END}"
    else
        echo -e "${COLOR_RED}compilation failed ${COLOR_END}"
    fi
fi
pwd