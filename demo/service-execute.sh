#!/bin/bash
set -e

url="http://pz-gateway.$PZDOMAIN"

serviceId=$1
[ "$serviceId" != "" ] || ( echo error: \$serviceId missing ; exit 1 )

job='{
    "type": "execute-service",
    "data": {
        "serviceId": "'"$serviceId"'",
        "dataInputs": { },
        "dataOutput": [ { "mimeType":"application/json", "type":"text" } ]
    }
}'

curl -X POST -u $PZUSER:$PZPASS -H 'Content-Type: application/json' \
    -d "$job" \
    $url/job
