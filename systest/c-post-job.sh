#!/bin/bash
set -e

source setup.sh

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

ret=$($curl -X POST -d "$job" $url/job)

echo "$ret"
