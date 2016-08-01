#!/bin/bash
set -e

source setup.sh

dataId=$1
[ "$dataId" != "" ] || ( echo error: \$dataId missing ; exit 1 )

ret=$($curl -X GET $url/data/$dataId)

echo "$ret"
