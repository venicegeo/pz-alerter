#!/bin/bash
set -e

source setup.sh

hello=`echo $PZSERVER | sed -e sXpz-gatewayXhttp://pzsvc-helloX`

service='{
    "url": "'"$hello"'",
    "contractUrl": "http://helloContract",
    "method": "GET",
    "isAsynchronous": "false",
    "resourceMetadata": {
        "name": "pzsvc-hello service",
        "description": "Hello World Example",
        "classType": {
           "classification": "UNCLASSIFIED"
        }
    }
}'

#echo POST
#echo "$service"
ret=$($curl -XPOST -d "$service" $url/service)

#echo RETURN:
#echo "$ret"

echo `extract serviceId "$ret"`
