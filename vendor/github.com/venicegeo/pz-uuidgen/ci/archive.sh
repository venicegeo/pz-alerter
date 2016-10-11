#!/bin/bash -ex

pushd `dirname $0`/.. > /dev/null
root=$(pwd -P)
popd > /dev/null
export GOPATH=$root/gogo

#----------------------------------------------------------------------

sh $root/ci/do_build.sh

#----------------------------------------------------------------------

# gather some data about the repo
source $root/ci/vars.sh

cd $root
cp $GOPATH/bin/$APP .
tar cvzf $APP.$EXT \
    $APP \
    uuidgen.cov \
    lint.txt \
    glide.lock \
    glide.yaml
tar tzf $APP.$EXT
