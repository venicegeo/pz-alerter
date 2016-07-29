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

service='{
"serviceId" : "0bcc6896-642e-4a30-a01e-6bd0467b57ba",
"url" : "http://pzsvc-hello.int.geointservices.io",
"contractUrl" : "http://pzsvc-hello.int.geointservices.io",
"method" : "GET",
"resourceMetadata" : {
"name" : "HELLO World Test",
"description" : "Hello world test",
"classType" : "U"
}
}'

curl -S -s -X POST -u $PZKEY:"" -H 'Content-Type: application/json' \
    -d "$service" \
    $url/service
