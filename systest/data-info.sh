#!/bin/bash
set -e

url="http://pz-gateway.$PZDOMAIN"

dataId=$1
[ "$dataId" != "" ] || ( echo error: \$dataId missing ; exit 1 )

curl -X GET -u $PZUSER:$PZPASS -H 'Content-Type: application/json' \
   $url/data/$dataId
echo
