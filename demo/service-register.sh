#!/bin/bash
set -e

url="http://pz-gateway.$PZDOMAIN"

service='{
    "url": "'"http://pzsvc-hello.$PZDOMAIN/"'",
    "contractUrl": "http://helloContract",
    "serviceId": "",
    "method": "GET",
    "resourceMetadata": {
        "name": "pzsvc-hello service",
        "description": "Hello World Example"
    }
}'

curl -X POST -u $PZUSER:$PZPASS -H 'Content-Type: application/json' \
    -d "$service" \
    $url/service
