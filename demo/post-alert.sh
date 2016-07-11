#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

json='{
    "alertId": "11",
    "triggerId": "22",
    "eventId": "33",
    "jobId": "44"
}'

echo
echo POST /alert
echo "$json"

ret=$(curl -S -s -XPOST -d "$json" "$url"/alert)

echo RETURN:
echo "$ret"
echo
