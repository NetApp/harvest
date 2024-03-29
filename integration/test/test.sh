SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCRIPT_DIR="$(dirname "$SCRIPT_DIR")"
echo "Dir : $SCRIPT_DIR"
cd "$SCRIPT_DIR"/test || exit
export PATH=$PATH:/usr/local/go/bin
if [ -z "$VERSION" ]; then
  VERSION="$(date +%Y.%m.%d%H | cut -c 3-)"
  echo "VERSION not supplied, using $VERSION"
fi

LD_FLAGS="-X ""'""github.com/netapp/harvest/v2/cmd/harvest/version.VERSION=${VERSION}""'"""
echo "$LD_FLAGS"

go mod tidy
go test -timeout 30m -ldflags="$LD_FLAGS"