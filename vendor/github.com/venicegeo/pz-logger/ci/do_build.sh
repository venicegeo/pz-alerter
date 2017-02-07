#!/bin/bash -ex

pushd `dirname $0`/.. > /dev/null
root=$(pwd -P)
popd > /dev/null
export GOPATH=$root/gogo

#----------------------------------------------------------------------

mkdir -p "$GOPATH" "$GOPATH"/bin "$GOPATH"/src "$GOPATH"/pkg

PATH=$PATH:"$GOPATH"/bin

go version

# install metalinter
go get -u github.com/alecthomas/gometalinter
gometalinter --install

# build ourself, and go there
go get github.com/venicegeo/pz-logger
cd $GOPATH/src/github.com/venicegeo/pz-logger

# run unit tests w/ coverage collection
go test -v -coverprofile=$root/logger.cov -coverpkg github.com/venicegeo/pz-logger/logger github.com/venicegeo/pz-logger/logger
go tool cover -func=$root/logger.cov -o $root/logger.cov.txt

# lint
# sh ci/metalinter.sh | tee $root/lint.txt
# wc -l $root/lint.txt
