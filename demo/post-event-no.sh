#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

eventTypeId=$1
if [ "$eventTypeId" == "" ]
then
    echo "error: \$eventTypeId missing"
    exit 1
fi

json='{
    "eventTypeId": "'"$eventTypeId"'",
    "data": {
        "filename": "dataset-a",
        "severity": 5,
        "code": "BBOX"
    }
}'

echo
echo POST /event
echo "$json"

ret=$(curl -S -s -XPOST -d "$json" "$url"/event)

echo RETURN:
echo "$ret"
echo
