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
info "creating package directories"
rm -rf "$BUILD" && mkdir -p "$BUILD" && cd "$BUILD"
mkdir -p "$BUILD/opt/harvest/"
mkdir -p "$BUILD/etc/harvest/"
mkdir -p "$BUILD/var/log/harvest/"
mkdir -p "$BUILD/var/run/harvest/"
cp -r "$SRC/src/" "opt/harvest/"
cp -r "$SRC/cmd/" "opt/harvest/"
cp -r "$SRC/docs/" "opt/harvest/"
cp -r "$SRC/README.md" "opt/harvest/"
cp -r "$SRC/conf/" "etc/harvest/"
cp -r "$SRC/grafana/" "etc/harvest/"
cp -r "$SRC/harvest.yml" "etc/harvest/"

info "creating DEB control file"
mkdir "DEBIAN"
echo "Package: harvest" > "DEBIAN/control"
echo "Version: $HARVEST_VERSION-$HARVEST_RELEASE" >> "DEBIAN/control"
#echo "Section: base" > "DEBIAN/control"
echo "Priority: optional" > "DEBIAN/control"
echo "Architecture: $HARVEST_ARCH" >> "DEBIAN/control"
cat "$SRC/cmd/deb/harvest.control" >> "DEBIAN/control"
cp "$SRC/cmd/install.sh" "DEBIAN/preinst"

# build binaries
info "compiling binaries..."
cd "$BUILD/opt/harvest/"
sh cmd/build.sh all
if [ ! $? -eq 0 ]; then
    error "compile failed"
    exit 1
fi

# build deb package
mkdir -p "$SRC/dist/$HARVEST_VERSION-$HARVEST_RELEASE"
dpkg-deb --build "$BUILD" "$SRC/dist/$HARVEST_VERSION-$HARVEST_RELEASE_$HARVEST_ARCH.deb"
if [ ! $? -eq 0 ]; then
    error "dpkg build failed"
    exit 1
fi

rm -Rf "$BUILDIR"
info "DEB package ready for distribution. Have a nice evening!"
exit 0