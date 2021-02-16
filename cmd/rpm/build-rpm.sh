
SRC="/tmp/src"
BUILD="/tmp/build"

function alert {
    echo -e "\033[1m\033[45m$1\033[0m"

}

function error {
    echo -e "\033[1m\033[41m$1\033[0m"
}

alert "copying source files"
rm -rf "$BUILD"
mkdir -p "$BUILD"
cd "$BUILD"
cp -r "$SRC" "harvest2"

# clean non-publishable files
cd "$BUILD/harvest2"
alert "cleaning up package"
rm -f "config.yaml"
rm -rf "archive"
rm -rf "dist"
rm -rf "log/*"
rm -rf "cert/*"
rm -rf "var/*"
rm -rf "bin/*"
rm -rf ".git"

# build binaries
alert "building binaries"
./cmd/build.sh harvest
./cmd/build.sh collector zapi
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
echo "%define version $HARVEST_BUILD_VERSION" > "rpm/SPECS/harvest.spec"
echo "%define arch $HARVEST_BUILD_ARCH" >> "rpm/SPECS/harvest.spec"
cat "$SRC/cmd/rpm/harvest.spec" >> "rpm/SPECS/harvest.spec"

# create tarball
alert "building tarball"
cd "$BUILD"
tar -czvf "rpm/SOURCES/harvest_$HARVEST_BUILD_VERSION.tgz" "harvest2"
alert "  -> [$BUILD/rpm/SOURCES/harvest_$HARVEST_BUILD_VERSION.tgz]"
file "$BUILD/rpm/SOURCES/harvest_$HARVEST_BUILD_VERSION.tgz"
if [ ! $? -eq 0 ]; then
    error "failed, aborting"
    exit 1
fi

# build rpm
alert "building rpm"
rpmbuild --target "$HARVEST_BUILD_ARCH" -bb "rpm/SPECS/harvest.spec"
if [ ! $? -eq 0 ]; then
    error "rpmbuild failed, aborting"
    exit 1
fi

# copy files & clean up
cd $BUILD
alert "copying files"
mkdir -p $SRC/dist/$HARVEST_BUILD_VERSION
mv -vf /root/rpmbuild/RPMS/* $SRC/dist/$HARVEST_BUILD_VERSION/
mv -vf rpm/SOURCES/* $SRC/dist/$HARVEST_BUILD_VERSION/
alert "cleaning up"
rm -rf "$BUILD"

# liked this final message by Chris Madden :-)
alert "RPM/TGZ package ready for distribution. Have a nice day!"
exit 0

