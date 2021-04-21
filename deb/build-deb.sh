#!/bin/bash

function info {
    echo -e "\033[1m\033[45m$1\033[0m"
}

function error {
    echo -e "\033[1m\033[41m$1\033[0m"
}

BUILD="/tmp/build"
SRC=$HARVEST_BUILD_SRC

if [ -z "$SRC" ]; then
    error "build source missing (\$HARVEST_BUILD_SRC)"
    exit 1
fi

# create build directory tree
rm -rf "$BUILD"
mkdir -p "$BUILD/opt/harvest/"
cd "$BUILD"
cp -r "$SRC/*" "opt/harvest/"
mkdir "DEBIAN"
echo "Package: harvest" > "DEBIAN/control"
echo "Version: $HARVEST_VERSION-$HARVEST_RELEASE" >> "DEBIAN/control"
echo "Architecture: $HARVEST_ARCH" >> "DEBIAN/control"
cat "$SRC/cmd/deb/harvest.control" >> "DEBIAN/control"

# build binaries
info "compiling binaries..."
cd "$BUILD/opt/harvest/"
sh cmd/build.sh all
if [ ! $? -eq 0 ]; then
    error "compile failed"
    exit 1
fi

# install directores
info "creating package directories"
cd "$BUILD"
mkdir -p "etc/harvest/"
mkdir -p "var/log/harvest/"
mkdir -p "var/run/harvest/"
mv -r "opt/harvest/config" "etc/harvest/"

# build deb package
mkdir -p "$SRC/dist/$HARVEST_VERSION-$HARVEST_RELEASE"
dpkg-deb --build "$BUILD" "$SRC/dist/$HARVEST_VERSION-$HARVEST_RELEASE"
if [ ! $? -eq 0 ]; then
    error "dpkg build failed"
    exit 1
fi

info "DEB package ready for distribution. Have a nice evening!"
exit 0