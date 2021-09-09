SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCRIPT_DIR="$(dirname "$SCRIPT_DIR")"
echo "Dir : $SCRIPT_DIR"
cd $SCRIPT_DIR/test
tag=${1?Specify valid test tag}
export PATH=$PATH:/usr/local/go/bin
go test -tags=$tag