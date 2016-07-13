#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

json='{
    "name": "USDadtaEvent2",
    "mapping": {
        "filename": "string",
        "code":     "string",
        "severity": "integer"
    }
}'

echo
echo POST /eventType
echo "$json"

ret=$(curl -S -s -XPOST -d "$json" "$url"/eventType)

echo RETURN:
echo "$ret"
echo
