#!/bin/bash
set -e

pushd "$(dirname "$0")/.." > /dev/null
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
go get github.com/venicegeo/pz-gocommon/gocommon
cd $GOPATH/src/github.com/venicegeo/pz-gocommon

# run unit tests w/ coverage collection
go test -v -coverprofile=$root/common.cov github.com/venicegeo/pz-gocommon/gocommon
go test -v -coverprofile=$root/elastic.cov github.com/venicegeo/pz-gocommon/elasticsearch
go test -v -coverprofile=$root/kafka.cov github.com/venicegeo/pz-gocommon/kafka

# lint
sh ci/metalinter.sh | tee $root/lint.txt
wc -l $root/lint.txt

#curl https://glide.sh/get | sh
#go tool cover -html=common.cov
