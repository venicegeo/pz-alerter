#!/bin/bash

curl="curl -S -s -u $PZKEY: -H Content-Type:application/json"

url="http://pz-gateway.int.geointservices.io"

eventTypeId="77bbe4c6-b1ac-4bbb-8e86-0f6e6a731c39"
serviceId="61985d9c-d4d0-45d9-a655-7dcf2dc08fad"

args='{\"name\":\"ME\", \"count\":5}'
echo $args

json='{
    "type": "execute-service",
    "data": {
        "dataInputs": {
            "": {
                "content": "'"$args"'",
                "type":     "body",
                "mimeType": "application/json"
            }
        },
        "dataOutput": [
            {
                "mimeType": "application/json",
                "type":     "text"
            }
        ],
        "serviceId": "'"$serviceId"'"
    }
}'

echo
echo POST /job
echo "$json"

$curl -XPOST -d "$json" "$url"/job > tmp

echo RETURN:
cat tmp
echo


jobId=`cat tmp | grep jobId | cut -f 2 -d ":" | cut -f 1 -d "," | cut -d \" -f 2`
echo JOB ID: $jobId


# get status on the job
$curl -XGET $url/job/$jobId
