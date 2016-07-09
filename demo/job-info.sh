#!/bin/bash
set -e

url="http://pz-gateway.$PZDOMAIN"

jobId=$1
[ "$jobId" != "" ] || ( echo error: \$jobId missing ; exit 1 )

curl -X GET -u $PZUSER:$PZPASS -H 'Content-Type: application/json' \
   $url/job/$jobId

echo
