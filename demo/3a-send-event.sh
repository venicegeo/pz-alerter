#!/bin/bash

# shellcheck disable=SC1091
source 0-setup.sh

etId=$1

cat > tmp <<foo
{
    "eventTypeId": "$etId",
    "createdOn": "2007-05-05T14:30:00Z",
    "data": {
        "filename": "dataset-a",
        "severity": 5,
        "code": "BBOX"
    }
}
foo

json=$(cat tmp)

echo
echo POST /v2/event
echo "$json"

ret=$(curl -S -s -XPOST -d "$json" "$WHOST"/event)

echo RETURN:
echo "$ret"
echo

rm tmp
