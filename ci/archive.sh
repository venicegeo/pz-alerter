#!/bin/bash
set -e

pushd "$(dirname "$0")/.." > /dev/null
root=$(pwd -P)
popd > /dev/null
export GOPATH=$root/gogo

#----------------------------------------------------------------------

sh $root/ci/do_build.sh

#----------------------------------------------------------------------

# gather some data about the repo
source $root/ci/vars.sh

cd $root
tar cvzf $APP.$EXT \
    $GOPATH/bin/pz-workflow \
    workflow.cov \
    lint.txt \
    glide.lock \
    glide.yaml
tar tzf $APP.$EXT
