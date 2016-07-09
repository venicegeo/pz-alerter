#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

cat > tmp <<foo
{
    "eventTypeId": "1234",
    "name": "USDadtaEvent",
    "mapping": {
        "filename": "string",
        "code":     "string",
        "severity": "integer"
    }
}
foo

json=$(cat tmp)

echo
echo POST /eventType
echo "$json"

ret=$(curl -S -s -XPOST -d "$json" "$url"/eventType)

echo RETURN:
echo "$ret"
echo

rm tmp
