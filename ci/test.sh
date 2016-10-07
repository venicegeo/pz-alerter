#!/bin/bash -ex

pushd "$(dirname "$0")"/.. > /dev/null
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
go get github.com/venicegeo/pz-workflow
cd $GOPATH/src/github.com/venicegeo/pz-workflow

#----------------------------------------------------------------------

go test -v -coverprofile=workflow.cov github.com/venicegeo/pz-workflow/workflow
