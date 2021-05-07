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

# copy files and directories
echo "copying source files"
rm -rf "$BUILD"
mkdir -p "$BUILD/harvest/bin"
cp -r "$SRC/pkg/" "$BUILD/harvest/"
cp -r "$SRC/cmd/" "$BUILD/harvest/"
cp -r "$SRC/grafana/" "$BUILD/harvest/"
cp -r "$SRC/docs/" "$BUILD/harvest/"
cp -r "$SRC/conf/" "$BUILD/harvest/"
cp -r "$SRC/rpm/" "$BUILD/harvest/"
cp "$SRC/harvest.example.yml" "$BUILD/harvest/"
cp "$SRC/go.mod" "$BUILD/harvest/"
if [ -d "$SRC/vendor" ]; then
    cp -r "$SRC/vendor" "$BUILD/harvest/"
fi
cp "$SRC/Makefile" "$BUILD/harvest/"
cp "$SRC/README.md" "$BUILD/harvest/"
cp "$SRC/LICENSE" "$BUILD/harvest/"

# update build and package version
sed -i -E "s/(\s*BUILD\s*=\s*\")\w*(\")/\1rpm $HARVEST_ARCH\2/" $BUILD/harvest/cmd/harvest/version/version.go
sed -i -E "s/(\s*VERSION\s*=\s*\")\w*(\")/\1$HARVEST_VERSION\2/" $BUILD/harvest/cmd/harvest/version/version.go
sed -i -E "s/(\s*RELEASE\s*=\s*\")\w*(\")/\1$HARVEST_RELEASE\2/" $BUILD/harvest/cmd/harvest/version/version.go

# build binaries
echo "building binaries"
cd "$BUILD/harvest"
make all
if [ ! $? -eq 0 ]; then
    echo "build failed, aborting"
    exit 1
fi

# create rpm build package
cd "$BUILD"
rm -rf "rpm"
mkdir -p "rpm/RPMS"
mkdir "rpm/SOURCES"
mkdir "rpm/SRPMS"
mkdir "rpm/SPECS"
echo "%define release $HARVEST_RELEASE" > "rpm/SPECS/spec"
echo "%define version $HARVEST_VERSION" >> "rpm/SPECS/spec"
echo "%define arch $HARVEST_ARCH" >> "rpm/SPECS/spec"
cat "$SRC/rpm/spec" >> "rpm/SPECS/spec"

# create tarball
echo "building binary tarball"
cd "$BUILD"
TGZ_FILEPATH="$BUILD/rpm/SOURCES/harvest_$HARVEST_VERSION-$HARVEST_RELEASE.tgz"
tar -czvf "$TGZ_FILEPATH" "harvest"
if [ ! $? -eq 0 ]; then
    echo "failed, aborting"
    exit 1
fi
echo "  -> [$TGZ_FILEPATH]"

echo "building source tarball"
rm -rf "harvest/bin/"
TARGET_DIR="$SRC/dist/$HARVEST_VERSION-$HARVEST_RELEASE"
mkdir -p $TARGET_DIR
TGZ_SOURCE="$TARGET_DIR/harvest_${HARVEST_VERSION}-${HARVEST_RELEASE}_source.tgz"
tar -czvf "$TGZ_SOURCE" "harvest"
if [ ! $? -eq 0 ]; then
    echo "failed, aborting"
    exit 1
fi
echo "  -> [$TGZ_SOURCE]"

# build rpm
echo "building rpm"
rpmbuild --target "$HARVEST_ARCH" -bb "rpm/SPECS/spec"
if [ ! $? -eq 0 ]; then
    echo "rpmbuild failed, aborting"
    exit 1
fi

# copy files & clean up
cd $BUILD
echo "copying packages"
mv -vf /root/rpmbuild/RPMS/* $TARGET_DIR/
mv -vf rpm/SOURCES/* $TARGET_DIR/
echo "cleaning up..."
rm -rf "$BUILD"

# liked this final message by Chris Madden :-)
echo "RPM and TGZ packages ready for distribution. Have a nice day!"
exit 0
