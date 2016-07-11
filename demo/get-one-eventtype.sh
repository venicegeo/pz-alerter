#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

eventTypeId=$1
if [ "$eventTypeId" == "" ]
then
    echo "error: \$eventTypeId missing"
    exit 1
fi

echo
echo GET /eventType/$eventTypeId

ret=$(curl -S -s -XGET "$url"/eventType/$eventTypeId)

echo RETURN:
echo "$ret"
echo
