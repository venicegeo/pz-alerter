#!/bin/bash -ex

pushd `dirname $0`/.. > /dev/null
root=$(pwd -P)
popd > /dev/null

export GOPATH=$root/gogo
mkdir -p $GOPATH

###

go get github.com/stretchr/testify/suite
go get github.com/stretchr/testify/assert

go get gopkg.in/olivere/elastic.v3

go get github.com/venicegeo/pz-gocommon

go get github.com/venicegeo/pz-workflow
go test -v github.com/venicegeo/pz-workflow/server

###
