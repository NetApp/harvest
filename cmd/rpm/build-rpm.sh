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

# copy files and directories
info "copying source files"
rm -rf "$BUILD"
mkdir -p "$BUILD/harvest/bin"
cp -r "$SRC/src/" "$BUILD/harvest/"
cp -r "$SRC/cmd/" "$BUILD/harvest/"
cp -r "$SRC/grafana/" "$BUILD/harvest/"
cp -r "$SRC/docs/" "$BUILD/harvest/"
cp -r "$SRC/conf/" "$BUILD/harvest/"
cp -r "$SRC/README.md" "$BUILD/harvest/"
cp -r "$SRC/harvest.yml" "$BUILD/harvest/"

# build binaries
info "building binaries"
cd "$BUILD/harvest"
./cmd/build.sh all
if [ ! $? -eq 0 ]; then
    error "compiling binaries, failed, aborting"
    exit 1
fi
cd "$BUILD"

# create rpm build package
rm -rf "rpm"
mkdir -p "rpm/RPMS"
mkdir "rpm/SOURCES"
mkdir "rpm/SRPMS"
mkdir "rpm/SPECS"
echo "%define release $HARVEST_RELEASE" > "rpm/SPECS/harvest.spec"
echo "%define version $HARVEST_VERSION" >> "rpm/SPECS/harvest.spec"
echo "%define arch $HARVEST_ARCH" >> "rpm/SPECS/harvest.spec"
cat "$SRC/cmd/rpm/harvest.spec" >> "rpm/SPECS/harvest.spec"

# create tarball
info "building tarball"
cd "$BUILD"
TGZ_FILEPATH="$BUILD/rpm/SOURCES/harvest_$HARVEST_VERSION-$HARVEST_RELEASE.tgz"
tar -czvf "$TGZ_FILEPATH" "harvest"
if [ ! $? -eq 0 ]; then
    error "failed, aborting"
    exit 1
fi
info "  -> [$TGZ_FILEPATH]"
file "$TGZ_FILEPATH" # DEBUG


# build rpm
info "building rpm"
rpmbuild --target "$HARVEST_ARCH" -bb "rpm/SPECS/harvest.spec"
if [ ! $? -eq 0 ]; then
    error "rpmbuild failed, aborting"
    exit 1
fi

# copy files & clean up
cd $BUILD
info "copying packages"
TARGET_DIR="$SRC/dist/$HARVEST_VERSION-$HARVEST_RELEASE"
mkdir -p $TARGET_DIR
mv -vf /root/rpmbuild/RPMS/* $TARGET_DIR/
mv -vf rpm/SOURCES/* $TARGET_DIR
info "cleaning up..."
rm -rf "$BUILD"

# liked this final message by Chris Madden :-)
info "RPM and TGZ packages ready for distribution. Have a nice day!"
exit 0
