SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCRIPT_DIR="$(dirname "$SCRIPT_DIR")"
echo "Dir : $SCRIPT_DIR"
cd $SCRIPT_DIR/test
tag=${1?Specify valid test tag}
export PATH=$PATH:/usr/local/go/bin
if [ ! -n "$VERSION" ]; then
  echo "VERSION not supplied."
  VERSION="$(date +%Y.%m.%d%H | cut -c 3-)"
fi

LD_FLAGS="-X ""'""goharvest2/cmd/harvest/version.VERSION=${VERSION}""'"""
echo $LD_FLAGS

go mod tidy
go test -timeout 30m -tags=$tag -ldflags="$LD_FLAGS"