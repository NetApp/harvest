#!/bin/sh

# optional env variable telling us root of filesystem
ROOT=${BUILD_ROOT}
HARVEST_USER="harvestu"
HARVEST_GROUP="harvestu"

function info {
    # print message in green
    echo -e "\033[1m\033[45m$1\033[0m"

}

function error {
    # print message in red
    echo -e "\033[1m\033[41m$1\033[0m"
}

function install {
    echo "creating harvest user and group [$HARVEST_USER:$HARVEST_GROUP]"
    addgroup --quiet --system "$HARVEST_GROUP"
    adduser --quiet --system --no-create-home --ingroup "$HARVEST_GROUP" --disabled-password --shell /bin/false "$HARVEST_USER"

    echo "creating package directories"
    mkdir -p $ROOT/opt/harvest
    mkdir -p $ROOT/etc/harvest
    mkdir -p $ROOT/var/log/harvest
    mkdir -p $ROOT/var/run/harvest

    echo "setting user permissions"
    chown -R $HARVEST_USER:$HARVEST_GROUP $ROOT/opt/harvest
    chown -R $HARVEST_USER:$HARVEST_GROUP $ROOT/etc/harvest
    chown -R $HARVEST_USER:$HARVEST_GROUP $ROOT/var/log/harvest
    chown -R $HARVEST_USER:$HARVEST_GROUP $ROOT/var/run/harvest

    mv harvest.yaml config/ $ROOT/etc/harvest/
    mv * $ROOT/opt/harvest
    ln -s $ROOT/opt/harvest/bin/harvest $ROOT/usr/local/bin/harvest
    info "installation complete"
}

function uninstall {
    echo "stopping harvest"
    harvest stop
    unlink /usr/local/bin/harvest
    echo "cleaning log files"
    rm -Rf $ROOT/var/log/harvest
    echo "cleaning pid files"
    rm -Rf $ROOT/var/run/harvest
    echo "config and certificate files left in [$ROOT/etc/harvest]"
    echo "please remove manually if you don't need them"

    echo "deleting harvest user and group"
    userdel "$HARVEST_USER"
    groupdel "$HARVEST_GROUP"
    
    info "uninstall complete"
}

if [ "$1" == "uninstall" ]; then
    uninstall
else
    install
fi
