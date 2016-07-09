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
    "createdOn": "2010-05-05T14:30:00Z",
    "data": {
        "filename": "dataset-c",
        "severity": 5,
        "code": "PHONE"
    }
}'

echo
echo POST /event
echo "$json"

ret=$(curl -S -s -XPOST -d "$json" "$url"/event)

echo RETURN:
echo "$ret"
echo
