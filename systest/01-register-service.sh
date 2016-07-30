#!/bin/bash
set -e

source setup.sh

service='{
    "url": "'"http://pzsvc-hello.$PZDOMAIN/"'",
    "contractUrl": "http://helloContract",
    "method": "POST",
    "resourceMetadata": {
        "name": "pzsvc-hello service",
        "description": "Hello World Example",
        "classType": "U"
    }
}'

#echo POST
#echo "$service"

ret=$($curl -XPOST -d "$service" $url/service)

#echo RETURN:
#echo "$ret"

echo `extract serviceId "$ret"`
