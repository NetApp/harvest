#!/bin/bash

#
# Copyright NetApp Inc, 2021 All rights reserved
#

BUILD="/tmp/build"
SRC=$HARVEST_BUILD_SRC

if [ -z "$SRC" ]; then
    echo "build source missing (\$HARVEST_BUILD_SRC)"
    exit 1
fi

echo "\033[1m\033[46m--> building DEB/$HARVEST_ARCH for \
    harvest $HARVEST_VERSION-$HARVEST_RELEASE \033[0m"

# create build directory tree
echo " --> create package directories"
rm -rf "$BUILD"
mkdir -p "$BUILD"
mkdir -p "$BUILD/opt/harvest/bin/"
cp -r "$SRC/grafana" "$SRC/conf" "$BUILD/opt/harvest/"
cp "$SRC/harvest.yml" "$SRC/prom-stack.yml" "$SRC/harvest.cue" "$BUILD/opt/harvest/"
cp -r "$SRC/pkg/" "$SRC/cmd/" "$SRC/docs/" "$SRC/docker/" "$BUILD/opt/harvest/"
cp -r "$SRC/rpm/" "$SRC/deb/" "$SRC/service/" "$SRC/autosupport/" "$SRC/.git" "$BUILD/opt/harvest/"
cp "$SRC/Makefile" "$SRC/README.md" "$SRC/LICENSE" "$SRC/go.mod" "$SRC/go.sum" "$BUILD/opt/harvest/"
if [ -d "$SRC/vendor" ]; then
    cp -r "$SRC/vendor" "$BUILD/opt/harvest/"
fi

# copy and modify debian packaging files
echo " --> create DEB control file"
mkdir "$BUILD/DEBIAN/"
cp "$SRC/deb/preinst" "$SRC/deb/postinst" "$SRC/deb/prerm" "$SRC/deb/postrm" "$BUILD/DEBIAN/"
echo "Package: harvest" > "$BUILD/DEBIAN/control"
echo "Version: $HARVEST_VERSION-$HARVEST_RELEASE" >> "$BUILD/DEBIAN/control"
echo "Architecture: $HARVEST_ARCH" >> "$BUILD/DEBIAN/control"
cat "$SRC/deb/control" >> "$BUILD/DEBIAN/control"

echo " --> update version & build info in [cmd/harvest/version/version.go]"
cd "$BUILD/opt/harvest/cmd/harvest/version"

# build binaries, since arch of build machine might not be the same as the target machine
# export a variable that Makefile will pass to the go compiler
cd "$BUILD/opt/harvest/"
export GOOS="linux"
export GOARCH="$HARVEST_ARCH"
if [ "$HARVEST_ARCH" = "armhf" ]; then
    export GOARCH="arm"
    export GOARM="7"
fi
echo " --> build harvest with envs [GOOS=$GOOS, GOARCH=$GOARCH, GOARM=$GOARM]"
if [ -n "$ASUP_MAKE_TARGET" ] && [ -z "$GIT_TOKEN" ]
then
      echo "GIT_TOKEN is required when ASUP_MAKE_TARGET is passed!"
      exit 1
fi

if [ -n "$ASUP_MAKE_TARGET" ] && [ -n "$GIT_TOKEN" ]
then
      make asup build VERSION=$VERSION RELEASE=$RELEASE ASUP_MAKE_TARGET=$ASUP_MAKE_TARGET GIT_TOKEN=$GIT_TOKEN
else
      make build VERSION=$HARVEST_VERSION RELEASE=$HARVEST_RELEASE
fi
if [ ! $? -eq 0 ]; then
    error "     build failed"
    exit 1
fi

# build deb package
PACKAGE_DIR="$SRC/dist/$HARVEST_VERSION-$HARVEST_RELEASE"
PACKAGE_NAME="harvest-${HARVEST_VERSION}-${HARVEST_RELEASE}.${HARVEST_ARCH}.deb"
mkdir -p "$PACKAGE_DIR"
rm -f "$PACKAGE_DIR/$PACKAGE_NAME"
dpkg-deb --build "$BUILD" "$PACKAGE_DIR/$PACKAGE_NAME"
if [ ! $? -eq 0 ]; then
    error "dpkg build failed"
    exit 1
fi
echo " --> created package: [$PACKAGE_DIR/$PACKAGE_NAME]"
echo " --> cleanup"
rm -Rf "$BUILD"
echo "DEB package ready for distribution. Have a nice evening!"
exit 0
