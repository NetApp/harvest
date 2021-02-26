
SRC="/tmp/src"
BUILD="/tmp/build"

function alert {
    echo -e "\033[1m\033[45m$1\033[0m"
}

function error {
    echo -e "\033[1m\033[41m$1\033[0m"
}

# copy files and directories
alert "copying source files"
rm -rf "$BUILD"
mkdir -p "$BUILD/harvest/bin"
cp -r "$SRC/src/" "$BUILD/harvest/"
cp -r "$SRC/cmd/" "$BUILD/harvest/"
cp -r "$SRC/grafana/" "$BUILD/harvest/"
cp -r "$SRC/docs/" "$BUILD/harvest/"
cp -r "$SRC/config/" "$BUILD/harvest/"
cp -r "$SRC/ReadMe.md" "$BUILD/harvest/"

# build binaries
alert "building binaries"
cd "$BUILD/harvest"
./cmd/build.sh harvest
./cmd/build.sh collector zapi
./cmd/build.sh collector zapiperf
./cmd/build.sh exporter prometheus
./cmd/build.sh poller
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
alert "building tarball"
cd "$BUILD"
TGZ_FILEPATH="$BUILD/rpm/SOURCES/harvest_$HARVEST_VERSION-$HARVEST_RELEASE.tgz"
tar -czvf "$TGZ_FILEPATH" "harvest"
if [ ! $? -eq 0 ]; then
    error "failed, aborting"
    exit 1
fi
alert "  -> [$TGZ_FILEPATH]"
file "$TGZ_FILEPATH" # DEBUG


# build rpm
alert "building rpm"
rpmbuild --target "$HARVEST_ARCH" -bb "rpm/SPECS/harvest.spec"
if [ ! $? -eq 0 ]; then
    error "rpmbuild failed, aborting"
    exit 1
fi

# copy files & clean up
cd $BUILD
alert "copying packages"
TARGET_DIR="$SRC/dist/$HARVEST_VERSION-$HARVEST_RELEASE"
mkdir -p $TARGET_DIR
mv -vf /root/rpmbuild/RPMS/* $TARGET_DIR/
mv -vf rpm/SOURCES/* $TARGET_DIR
alert "cleaning up..."
rm -rf "$BUILD"

# liked this final message by Chris Madden :-)
alert "RPM and TGZ packages ready for distribution. Have a nice day!"
exit 0
