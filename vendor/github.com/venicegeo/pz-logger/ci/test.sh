#!/bin/bash -ex

pushd `dirname $0`/.. > /dev/null
root=$(pwd -P)
popd > /dev/null

#----------------------------------------------------------------------

export GOPATH=$root/gogo
mkdir -p "$GOPATH"

# glide expects these to already exist
mkdir "$GOPATH"/bin "$GOPATH"/src "$GOPATH"/pkg

PATH=$PATH:"$GOPATH"/bin

curl https://glide.sh/get | sh

# get ourself, and go there
go get github.com/venicegeo/pz-logger
cd $GOPATH/src/github.com/venicegeo/pz-logger

#----------------------------------------------------------------------

go test -v -coverprofile=logger.cov -coverpkg github.com/venicegeo/pz-logger/logger github.com/venicegeo/pz-logger/logger

###
