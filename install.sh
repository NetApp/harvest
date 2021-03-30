#!/bin/sh

function install {
    make install $@
}

function uninstall {
    make uninstall $@
}

if [ "$1" == "uninstall" ]; then
    uninstall
else
    install
fi
