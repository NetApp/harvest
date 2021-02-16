
SRC="/tmp/src"
BUILD="/tmp/build"

echo "copying source files"
rm -rf "$BUILD"
mkdir -p "$BUILD"
cd "$BUILD"
cp -r "$SRC" "harvest2"

# clean non-publishable files
cd "$BUILD/harvest2"
echo "cleaning up package"
rm -f "config.yaml"
rm -rf "archive"
rm -rf "dist"
rm -rf "log/*"
rm -rf "cert/*"
rm -rf "var/*"
rm -rf "bin/*"

# build binaries
echo "building binaries"
./cmd/build.sh harvest
./cmd/build.sh collector zapi
./cmd/build.sh exporter prometheus
./cmd/build.sh poller
if [ ! $? -eq 0 ]; then
    echo "compiling binaries, failed, aborting"
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
echo "building tarball"
cd "$BUILD"
tar -czvf "rpm/SOURCES/harvest_$HARVEST_BUILD_VERSION.tgz" "harvest2"
if [ ! $? -eq 0] then
    echo "failed, aborting"
    exit 1
fi

# build rpm
echo "building rpm"
rpmbuild --target "$HARVEST_BUILD_ARCH" -bb "rpm/SPECS/harvest.spec"
if [ ! $? -eq 0 ]; then
    echo "rpmbuild failed, aborting"
    exit 1
fi

# copy files & clean up
echo "copying files"
mkdir -p "$SRC/dist/$HARVEST_BUILD_VERSION"
mv -vf "/root/rpmbuild/RPMS/*" "$SRC/dist/$HARVEST_BUILD_VERSION/"
mv -vf "rpm/SOURCES/*" "$SRC/dist/$HARVEST_BUILD_VERSION/"
echo "cleaning up"
rm -rf "$BUILD"

# liked this final message by Chris Madden :-)
echo "RPM/TGZ package ready for distribution. Have a nice day!"
exit 0

