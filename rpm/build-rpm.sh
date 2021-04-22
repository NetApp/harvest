#!/bin/bash

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
cp -r "$SRC/harvest.yml" "$BUILD/harvest/"
cp -r "$SRC/go.mod" "$BUILD/harvest/"
cp -r "$SRC/Makefile" "$BUILD/harvest/"
cp -r "$SRC/README.md" "$BUILD/harvest/"
cp -r "$SRC/SUPPORT.md" "$BUILD/harvest/"
cp -r "$SRC/CONTRIBUTING.md" "$BUILD/harvest/"
cp -r "$SRC/LICENSE" "$BUILD/harvest/"

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
echo "%define release $HARVEST_RELEASE" > "rpm/SPECS/harvest.spec"
echo "%define version $HARVEST_VERSION" >> "rpm/SPECS/harvest.spec"
echo "%define arch $HARVEST_ARCH" >> "rpm/SPECS/harvest.spec"
cat "$SRC/rpm/harvest.spec" >> "rpm/SPECS/harvest.spec"

# create tarball
echo "building tarball"
cd "$BUILD"
TGZ_FILEPATH="$BUILD/rpm/SOURCES/harvest_$HARVEST_VERSION-$HARVEST_RELEASE.tgz"
tar -czvf "$TGZ_FILEPATH" "harvest"
if [ ! $? -eq 0 ]; then
    echo "failed, aborting"
    exit 1
fi
echo "  -> [$TGZ_FILEPATH]"
file "$TGZ_FILEPATH" # DEBUG

# build rpm
echo "building rpm"
rpmbuild --target "$HARVEST_ARCH" -bb "rpm/SPECS/harvest.spec"
if [ ! $? -eq 0 ]; then
    echo "rpmbuild failed, aborting"
    exit 1
fi

# copy files & clean up
cd $BUILD
echo "copying packages"
TARGET_DIR="$SRC/dist/$HARVEST_VERSION-$HARVEST_RELEASE"
mkdir -p $TARGET_DIR
mv -vf /root/rpmbuild/RPMS/* $TARGET_DIR/
mv -vf rpm/SOURCES/* $TARGET_DIR/
echo "cleaning up..."
rm -rf "$BUILD"

# liked this final message by Chris Madden :-)
echo "RPM and TGZ packages ready for distribution. Have a nice day!"
exit 0
