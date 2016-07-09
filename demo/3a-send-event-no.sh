#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

eventtypeId=$1
[ "$eventTypeId" != "" ] || ( echo error: \$eventTypeId missing ; exit 1 )

json='{
    "eventTypeId": "'"$eventTypeId"'",
    "createdOn": "2007-05-05T14:30:00Z",
    "data": {
        "filename": "dataset-a",
        "severity": 5,
        "code": "BBOX"
    }
}'

echo
echo POST /v2/event
echo "$json"

ret=$(curl -S -s -XPOST -d "$json" "$url"/event)

echo RETURN:
echo "$ret"
echo

rm tmp
