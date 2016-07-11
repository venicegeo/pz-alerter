#!/bin/bash
set -e

url="http://pz-gateway.$PZDOMAIN"

serviceId=$1
[ "$serviceId" != "" ] || ( echo error: \$serviceId missing ; exit 1 )

curl -X GET -u $PZUSER:$PZPASS -H 'Content-Type: application/json' \
   $url/service/$serviceId

echo
